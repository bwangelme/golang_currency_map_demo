- runUnsafeTest

```
ø> go run ./main.go
=== 不安全版本测试（可能 panic）===
goroutine1: 执行 json.Unmarshal
goroutine2: 读取 nacosVal.Data

fatal error: concurrent map read and map write

goroutine 7 [running]:
internal/runtime/maps.fatal({0x1022bfe72?, 0x1400018270b?})
        /Users/mico/.gvm/gos/go1.25.5/src/runtime/panic.go:1046 +0x20
main.runUnsafeTest.func2()
        /Users/mico/Github/golang_currency_map_demo/main.go:54 +0xcc
created by main.runUnsafeTest in goroutine 1
        /Users/mico/Github/golang_currency_map_demo/main.go:50 +0x2a0

goroutine 1 [sync.WaitGroup.Wait]:
sync.runtime_SemacquireWaitGroup(0x102231490?, 0xa0?)
        /Users/mico/.gvm/gos/go1.25.5/src/runtime/sema.go:114 +0x38
sync.(*WaitGroup).Wait(0x14000010750)
        /Users/mico/.gvm/gos/go1.25.5/src/sync/waitgroup.go:206 +0xa8
main.runUnsafeTest()
        /Users/mico/Github/golang_currency_map_demo/main.go:61 +0x2a8
main.main()
        /Users/mico/Github/golang_currency_map_demo/main.go:76 +0x90

goroutine 6 [sleep]:
time.Sleep(0x3e8)
        /Users/mico/.gvm/gos/go1.25.5/src/runtime/time.go:363 +0x150
main.runUnsafeTest.func1()
        /Users/mico/Github/golang_currency_map_demo/main.go:43 +0xc8
created by main.runUnsafeTest in goroutine 1
        /Users/mico/Github/golang_currency_map_demo/main.go:34 +0x24c
exit status 2
```

- runReflectionTest

```
ø> go run -race ./main.go reflect
=== 反射版本测试（可能 panic）===
goroutine1: 使用反射 + json.Unmarshal 更新 map
goroutine2: 读取 map

==================
WARNING: DATA RACE
Write at 0x00c000058030 by goroutine 6:
  reflect.Value.Set()
      /Users/mico/.gvm/gos/go1.25.5/src/reflect/value.go:2139 +0x168
  main.updateMapWithReflection()
      /Users/mico/Github/golang_currency_map_demo/main.go:74 +0x150
  main.runReflectionTest.func1()
      /Users/mico/Github/golang_currency_map_demo/main.go:107 +0x114

Previous read at 0x00c000058030 by goroutine 7:
  main.runReflectionTest.func2()
      /Users/mico/Github/golang_currency_map_demo/main.go:119 +0x100

Goroutine 6 (running) created at:
  main.runReflectionTest()
      /Users/mico/Github/golang_currency_map_demo/main.go:98 +0x37c
  main.main()
      /Users/mico/Github/golang_currency_map_demo/main.go:141 +0x63c

Goroutine 7 (running) created at:
  main.runReflectionTest()
      /Users/mico/Github/golang_currency_map_demo/main.go:115 +0x434
  main.main()
      /Users/mico/Github/golang_currency_map_demo/main.go:141 +0x63c
==================
==================
WARNING: DATA RACE
Write at 0x00c000190000 by goroutine 6:
  runtime.mapaccess2_faststr()
      /Users/mico/.gvm/gos/go1.25.5/src/internal/runtime/maps/runtime_faststr_swiss.go:162 +0x29c
  reflect.mapassign_faststr0()
      /Users/mico/.gvm/gos/go1.25.5/src/runtime/map_swiss.go:264 +0x24
  reflect.Value.SetMapIndex()
      /Users/mico/.gvm/gos/go1.25.5/src/reflect/map_swiss.go:427 +0x294
  encoding/json.(*decodeState).object()
      /Users/mico/.gvm/gos/go1.25.5/src/encoding/json/decode.go:811 +0x14d0
  encoding/json.(*decodeState).value()
      /Users/mico/.gvm/gos/go1.25.5/src/encoding/json/decode.go:380 +0x78
  encoding/json.(*decodeState).unmarshal()
      /Users/mico/.gvm/gos/go1.25.5/src/encoding/json/decode.go:183 +0x238
  encoding/json.Unmarshal()
      /Users/mico/.gvm/gos/go1.25.5/src/encoding/json/decode.go:113 +0x1c8
  main.updateMapWithReflection()
      /Users/mico/Github/golang_currency_map_demo/main.go:77 +0x168
  main.runReflectionTest.func1()
      /Users/mico/Github/golang_currency_map_demo/main.go:107 +0x114

Previous read at 0x00c000190000 by goroutine 7:
  runtime.mapassign_fast64ptr()
      /Users/mico/.gvm/gos/go1.25.5/src/internal/runtime/maps/runtime_fast64_swiss.go:372 +0x38c
  main.runReflectionTest.func2()
      /Users/mico/Github/golang_currency_map_demo/main.go:119 +0x11c

Goroutine 6 (running) created at:
  main.runReflectionTest()
      /Users/mico/Github/golang_currency_map_demo/main.go:98 +0x37c
  main.main()
      /Users/mico/Github/golang_currency_map_demo/main.go:141 +0x63c

Goroutine 7 (running) created at:
  main.runReflectionTest()
      /Users/mico/Github/golang_currency_map_demo/main.go:115 +0x434
  main.main()
      /Users/mico/Github/golang_currency_map_demo/main.go:141 +0x63c
==================
goroutine2 完成（反射版本）
goroutine1 完成（反射版本）
反射版本测试完成！
最终 map 大小: 1
Found 2 data race(s)
exit status 66
```