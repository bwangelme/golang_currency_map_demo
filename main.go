package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

type nacosData struct {
	Data map[string]string `json:"data"`
}

var nacosVal = &nacosData{}

// 使用 Mutex 保护的安全版本
type safeNacosData struct {
	mu   sync.RWMutex
	Data map[string]string `json:"data"`
}

var safeNacosVal = &safeNacosData{
	Data: make(map[string]string),
}

func (n *safeNacosData) UnmarshalJSON(data []byte) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	var temp struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	n.Data = temp.Data
	return nil
}

func (n *safeNacosData) Set(key, value string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Data[key] = value
}

func (n *safeNacosData) Get(key string) (string, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	val, ok := n.Data[key]
	return val, ok
}

func runSafeTest() {
	fmt.Println("\n=== 安全版本测试（使用 Mutex）===")
	fmt.Println("goroutine1: 执行 json.Unmarshal（带锁）")
	fmt.Println("goroutine2: 直接修改 Data（带锁）")
	fmt.Println()

	var wg sync.WaitGroup
	iterations := 1000

	// goroutine1: 执行 json.Unmarshal
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			jsonData := fmt.Sprintf(`{"data":{"key%d":"value%d"}}`, i, i)
			someBytes := []byte(jsonData)
			safeNacosVal.UnmarshalJSON(someBytes)
			time.Sleep(time.Microsecond)
		}
		fmt.Println("goroutine1 完成（安全版本）")
	}()

	// goroutine2: 直接修改 map
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			safeNacosVal.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
			time.Sleep(time.Microsecond)
		}
		fmt.Println("goroutine2 完成（安全版本）")
	}()

	wg.Wait()
	fmt.Println("安全版本测试完成！")
	fmt.Printf("最终 map 大小: %d\n", len(safeNacosVal.Data))
}

func runUnsafeTest() {
	fmt.Println("=== 不安全版本测试（可能 panic）===")
	fmt.Println("goroutine1: 执行 json.Unmarshal")
	fmt.Println("goroutine2: 读取 nacosVal.Data")
	fmt.Println()

	// 初始化 map 并添加一些初始数据，以便 goroutine2 可以读取
	nacosVal.Data = make(map[string]string)
	for i := 0; i < 100; i++ {
		nacosVal.Data[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	var wg sync.WaitGroup
	iterations := 1000

	// goroutine1: 执行 json.Unmarshal
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			// 模拟 JSON 数据
			jsonData := fmt.Sprintf(`{"data":{"key%d":"value%d"}}`, i, i)
			someBytes := []byte(jsonData)

			// 这里会发生并发问题：json.Unmarshal 会重新分配 map（写操作）
			json.Unmarshal(someBytes, &nacosVal)
			time.Sleep(time.Microsecond) // 稍微延迟以增加竞争机会
		}
		fmt.Println("goroutine1 完成")
	}()

	// goroutine2: 读取 map
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			// 读取 map，在并发情况下不安全（与 json.Unmarshal 的写操作冲突）
			_ = nacosVal.Data[fmt.Sprintf("key%d", i%100)]
			time.Sleep(time.Microsecond) // 稍微延迟以增加竞争机会
		}
		fmt.Println("goroutine2 完成")
	}()

	// 等待所有 goroutine 完成
	wg.Wait()
	fmt.Println("不安全版本测试完成！")
	fmt.Printf("最终 map 大小: %d\n", len(nacosVal.Data))
}

// 使用反射更新 map 的辅助函数
func updateMapWithReflection(data []byte, result interface{}) error {
	rc := reflect.ValueOf(result)
	if strings.Contains(rc.Type().String(), "map") {
		rc = rc.Elem()
		if rc.Kind() == reflect.Map {
			rc.Set(reflect.MakeMap(rc.Type()))
		}
	}
	err := json.Unmarshal(data, result)
	return err
}

func runReflectionTest() {
	fmt.Println("=== 反射版本测试（可能 panic）===")
	fmt.Println("goroutine1: 使用反射 + json.Unmarshal 更新 map")
	fmt.Println("goroutine2: 读取 map")
	fmt.Println()

	// 初始化 map 并添加一些初始数据，以便 goroutine2 可以读取
	testMap := make(map[string]string)
	for i := 0; i < 100; i++ {
		testMap[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	var wg sync.WaitGroup
	iterations := 10000 // 增加迭代次数以提高触发 panic 的概率

	// goroutine1: 使用反射 + json.Unmarshal 更新 map
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			// 模拟 JSON 数据
			jsonData := fmt.Sprintf(`{"key%d":"value%d"}`, i, i)
			data := []byte(jsonData)

			// 将 map 作为 interface{} 传入，使用反射处理
			// 这里会发生并发问题：反射会清空 map，然后 json.Unmarshal 会写入（写操作）
			updateMapWithReflection(data, &testMap)
			time.Sleep(time.Microsecond) // 稍微延迟以增加竞争机会
		}
		fmt.Println("goroutine1 完成（反射版本）")
	}()

	// goroutine2: 读取 map
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			// 读取 map，在并发情况下不安全（与反射+json.Unmarshal 的写操作冲突）
			_ = testMap[fmt.Sprintf("key%d", i%100)]
			time.Sleep(time.Microsecond) // 稍微延迟以增加竞争机会
		}
		fmt.Println("goroutine2 完成（反射版本）")
	}()

	// 等待所有 goroutine 完成
	wg.Wait()
	fmt.Println("反射版本测试完成！")
	fmt.Printf("最终 map 大小: %d\n", len(testMap))
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "safe" {
		// 只运行安全版本
		runSafeTest()
	} else if len(os.Args) > 1 && os.Args[1] == "both" {
		// 运行两个版本
		runUnsafeTest()
		runSafeTest()
	} else if len(os.Args) > 1 && os.Args[1] == "reflect" {
		// 运行反射版本测试
		runReflectionTest()
	} else {
		// 默认只运行不安全版本（用于观察 panic）
		runUnsafeTest()
		fmt.Println("\n提示: 使用 'go run main.go safe' 运行安全版本")
		fmt.Println("     使用 'go run main.go both' 运行两个版本对比")
		fmt.Println("     使用 'go run main.go reflect' 运行反射版本测试")
	}
}
