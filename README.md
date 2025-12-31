# Go 并发 Map 测试

这个程序用于测试并发访问 map 时是否会发生 panic。

## 问题场景

### 场景 1: 直接使用 json.Unmarshal

```go
type nacosData struct {
    Data map[string]string `json:"data"`
}
var nacosVal = &nacosData{}

// goroutine1:
json.Unmarshal(someBytes, &nacosVal)

// goroutine2:
_ = nacosVal.Data["key"]  // 读取 map
```

### 场景 2: 使用反射 + json.Unmarshal

```go
var testMap map[string]string

// goroutine1: 使用反射处理 map
func updateMapWithReflection(data []byte, result interface{}) error {
    rc := reflect.ValueOf(result)
    if strings.Contains(rc.Type().String(), "map") {
        rc = rc.Elem()
        if rc.Kind() == reflect.Map {
            rc.Set(reflect.MakeMap(rc.Type()))  // 清空 map
        }
    }
    err := json.Unmarshal(data, result)
    return err
}

// goroutine2:
_ = testMap["key"]  // 读取 map
```

## 运行测试

### 1. 运行不安全版本（可能 panic）
```bash
go run main.go
```

### 2. 运行安全版本（使用 Mutex）
```bash
go run main.go safe
```

### 3. 运行两个版本对比
```bash
go run main.go both
```

### 4. 运行反射版本测试（使用反射 + json.Unmarshal）
```bash
go run *.go reflect
```

### 5. 使用 race detector 检测数据竞争
```bash
go run -race *.go reflect
```

## 预期结果

### 不安全版本（场景 1）
在并发情况下，这个程序很可能会：
1. 发生 panic: `fatal error: concurrent map read and map write`
2. 或者数据竞争导致数据不一致

注意：即使 goroutine2 只是读取 map，与 goroutine1 的写操作（json.Unmarshal）并发时仍然会 panic。

### 反射版本（场景 2）
使用反射 + json.Unmarshal 的场景同样存在并发问题：
1. **反射的 `rc.Set(reflect.MakeMap(rc.Type()))` 会清空 map**：这是一个写操作
2. **json.Unmarshal 会写入新的数据**：这也是写操作
3. **goroutine2 读取 map**：这是读操作
4. **并发读写会导致 panic**：`fatal error: concurrent map read and map write`

使用 `go run -race *.go reflect` 可以检测到数据竞争。

### 安全版本
使用 `sync.RWMutex` 保护后，程序可以安全运行，不会发生 panic。

## 原因分析

1. **json.Unmarshal 会重新分配 map**：当 `json.Unmarshal` 执行时，它会创建一个新的 map 并赋值给 `nacosVal.Data`，这涉及到 map 的写操作。

2. **读取 map 是读操作**：`nacosVal.Data["key"]` 是读操作。

3. **Go 的 map 不是并发安全的**：多个 goroutine 同时对同一个 map 进行读写操作会导致数据竞争，Go 的运行时检测到这种情况会直接 panic。即使一个 goroutine 只读，另一个 goroutine 在写，也会触发 `concurrent map read and map write` 的 panic。

## 解决方案

需要使用同步机制来保护共享数据：
- 使用 `sync.Mutex` 或 `sync.RWMutex`（见 `main.go` 中的 `safeNacosData` 和 `runSafeTest`）
- 使用 channel
- 使用 `sync.Map`（适用于特定场景）

