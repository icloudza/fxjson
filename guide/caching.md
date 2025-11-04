# 缓存系统指南

FxJSON 内置了强大的缓存系统，可以显著提升重复访问的性能。通过智能缓存机制，重复访问速度可提升 4-10 倍。

## 目录

- [缓存概述](#缓存概述)
- [内存缓存](#内存缓存)
- [解析缓存](#解析缓存)
- [懒加载](#懒加载)
- [性能监控](#性能监控)
- [缓存策略](#缓存策略)
- [实际应用场景](#实际应用场景)

## 缓存概述

FxJSON 的缓存系统包含：

1. **节点访问缓存**：自动缓存已访问的节点路径
2. **内存缓存**：可配置的外部缓存存储
3. **懒加载**：按需加载数据，延迟解析
4. **性能统计**：实时监控缓存效果

### 缓存特性

- **零开销**：缓存系统对性能影响极小
- **线程安全**：支持并发访问
- **自动管理**：无需手动维护缓存
- **可配置**：支持自定义缓存策略

## 内存缓存

### NewMemoryCache()

创建内存缓存实例。

```go
func NewMemoryCache(maxSize int) *MemoryCache
```

```go
package main

import (
    "fmt"
    "time"
    "github.com/icloudza/fxjson"
)

func basicCacheExample() {
    // 创建缓存，最多存储 1000 个项目
    cache := fxjson.NewMemoryCache(1000)

    // 存储 JSON 数据
    jsonData := `{"id": 1, "name": "张三", "email": "zhangsan@example.com"}`
    node := fxjson.FromString(jsonData)

    // 设置缓存，TTL 为 10 分钟
    cache.Set("user:1", node, 10*time.Minute)

    // 从缓存获取
    if cachedNode, found := cache.Get("user:1"); found {
        name := cachedNode.Get("name").StringOr("")
        fmt.Printf("从缓存获取用户: %s\n", name)
    }

    // 获取缓存统计
    stats := cache.Stats()
    fmt.Printf("缓存统计: %+v\n", stats)
}
```

### 缓存控制

```go
func cacheControl() {
    // 启用全局缓存
    cache := fxjson.NewMemoryCache(5000)
    fxjson.EnableCaching(cache)

    // 禁用缓存
    // fxjson.DisableCaching()

    // 使用缓存的解析
    jsonData := `{"users": [{"id": 1, "name": "张三"}, {"id": 2, "name": "李四"}]}`

    // 第一次解析
    start := time.Now()
    node1 := fxjson.FromBytesWithCache([]byte(jsonData), 5*time.Minute)
    parseTime1 := time.Since(start)

    // 第二次解析（从缓存）
    start = time.Now()
    node2 := fxjson.FromBytesWithCache([]byte(jsonData), 5*time.Minute)
    parseTime2 := time.Since(start)

    fmt.Printf("首次解析: %v\n", parseTime1)
    fmt.Printf("缓存解析: %v (提升 %.2fx)\n", parseTime2,
        float64(parseTime1)/float64(parseTime2))

    // 验证数据一致性
    name1 := node1.GetPath("users.0.name").StringOr("")
    name2 := node2.GetPath("users.0.name").StringOr("")
    fmt.Printf("数据一致性: %v (name1=%s, name2=%s)\n",
        name1 == name2, name1, name2)
}
```

## 解析缓存

### FromBytesWithCache()

带缓存的解析方法。

```go
func FromBytesWithCache(b []byte, ttl time.Duration) Node
```

```go
func parseWithCache() {
    // 准备测试数据
    jsonData := `{
        "company": "科技公司",
        "employees": [
            {"id": 1, "name": "张三", "department": "研发部"},
            {"id": 2, "name": "李四", "department": "产品部"},
            {"id": 3, "name": "王五", "department": "设计部"}
        ],
        "config": {
            "version": "1.0.0",
            "debug": true,
            "features": ["feature1", "feature2"]
        }
    }`

    // 设置缓存
    cache := fxjson.NewMemoryCache(100)
    fxjson.EnableCaching(cache)

    // 多次解析相同数据
    iterations := 100
    var totalTime time.Duration

    for i := 0; i < iterations; i++ {
        start := time.Now()

        // 前几次解析会创建缓存，��续从缓存读取
        node := fxjson.FromBytesWithCache([]byte(jsonData), 30*time.Second)

        // 使用数据
        _ = node.GetPath("employees.0.name").StringOr("")
        _ = node.GetPath("config.version").StringOr("")

        totalTime += time.Since(start)
    }

    avgTime := totalTime / time.Duration(iterations)
    fmt.Printf("平均解析时间: %v\n", avgTime)

    // 获取性能统计
    stats := fxjson.GetPerformanceStats()
    fmt.Printf("缓存命中率: %.2f%%\n", stats["cache_hit_rate"].(float64)*100)
}
```

### FromBytesWithMetrics()

带性能统计的解析。

```go
func fromBytesWithMetrics() {
    jsonData := `{"data": [{"value": 1}, {"value": 2}, {"value": 3}]}`

    // 解析并收集指标
    node := fxjson.FromBytesWithMetrics([]byte(jsonData))

    // 获取性能统计
    stats := fxjson.GetPerformanceStats()
    fmt.Printf("性能统计:\n")
    for key, value := range stats {
        fmt.Printf("  %s: %v\n", key, value)
    }

    // 访问数据
    values := node.Get("data")
    values.ArrayForEach(func(index int, item fxjson.Node) bool {
        value := item.Get("value").IntOr(0)
        fmt.Printf("Value %d: %d\n", index, value)
        return true
    })
}
```

## 懒加载

### NewLazyLoader()

创建懒加载器。

```go
func NewLazyLoader(loadFunc func() (Node, error)) *LazyLoader
```

```go
func lazyLoaderExample() {
    // 创建懒加载器
    loader := fxjson.NewLazyLoader(func() (fxjson.Node, error) {
        // 模拟耗时操作（如从数据库或网络加载）
        time.Sleep(100 * time.Millisecond)

        jsonData := `{
            "user_id": 1001,
            "profile": {
                "name": "张三",
                "email": "zhangsan@example.com",
                "preferences": {
                    "theme": "dark",
                    "language": "zh-CN"
                }
            },
            "permissions": ["read", "write", "admin"]
        }`

        return fxjson.FromString(jsonData), nil
    })

    // 检查是否已加载
    fmt.Printf("加载前: 已加载=%v\n", loader.IsLoaded())

    // 第一次访问会触发加载
    start := time.Now()
    node, err := loader.Get()
    if err != nil {
        panic(err)
    }
    firstLoadTime := time.Since(start)

    fmt.Printf("首次加载: %v\n", firstLoadTime)
    fmt.Printf("加载后: 已加载=%v\n", loader.IsLoaded())

    // 第二次访问使用缓存
    start = time.Now()
    node2, _ := loader.Get()
    secondAccessTime := time.Since(start)

    fmt.Printf("缓存访问: %v (提升 %.2fx)\n",
        secondAccessTime,
        float64(firstLoadTime)/float64(secondAccessTime))

    // 使用数据
    name := node.GetPath("profile.name").StringOr("")
    theme := node.GetPath("profile.preferences.theme").StringOr("")
    fmt.Printf("用户: %s, 主题: %s\n", name, theme)

    // 重置懒加载器
    loader.Reset()
    fmt.Printf("重置后: 已加载=%v\n", loader.IsLoaded())
}
```

## 性能监控

### 缓存统计

```go
func cacheMonitoring() {
    // 创建缓存并设置统计
    cache := fxjson.NewMemoryCache(1000)
    fxjson.EnableCaching(cache)

    // 生成测试数据
    dataSets := make(map[string][]byte)
    for i := 0; i < 100; i++ {
        key := fmt.Sprintf("data:%d", i)
        dataSets[key] = []byte(fmt.Sprintf(`{"id": %d, "value": %d}`, i, i*i))
    }

    // 随机访问数据
    for round := 0; round < 10; round++ {
        for key := range dataSets {
            // 随机选择访问
            if rand.Intn(2) == 0 {
                fxjson.FromBytesWithCache(dataSets[key], time.Minute)
            }
        }
    }

    // 获取详细统计
    stats := cache.Stats()
    fmt.Println("缓存统计详情:")
    fmt.Printf("  存储项目数: %d\n", stats.Size)
    fmt.Printf("  命中次数: %d\n", stats.Hits)
    fmt.Printf("  未命中次数: %d\n", stats.Misses)
    fmt.Printf("  命中率: %.2f%%\n", stats.HitRate()*100)
    fmt.Printf("  淘汰次数: %d\n", stats.Evictions)

    // 全局性能统计
    globalStats := fxjson.GetPerformanceStats()
    fmt.Println("\n全局性能统计:")
    for key, value := range globalStats {
        fmt.Printf("  %s: %v\n", key, value)
    }
}
```

### 性能基准测试

```go
func benchmarkCache() {
    // 准备测试数据
    largeJSON := generateLargeJSON(1000)

    // 测试无缓存性能
    fmt.Println("=== 无缓存测试 ===")
    start := time.Now()
    for i := 0; i < 100; i++ {
        node := fxjson.FromBytes(largeJSON)
        processData(node)
    }
    noCacheTime := time.Since(start)
    fmt.Printf("无缓存耗时: %v\n", noCacheTime)

    // 启用缓存
    cache := fxjson.NewMemoryCache(50)
    fxjson.EnableCaching(cache)
    defer fxjson.DisableCaching()

    // 测试有缓存性能
    fmt.Println("\n=== 有缓存测试 ===")
    start = time.Now()
    for i := 0; i < 100; i++ {
        node := fxjson.FromBytesWithCache(largeJSON, time.Minute)
        processData(node)
    }
    cacheTime := time.Since(start)
    fmt.Printf("有缓存耗时: %v\n", cacheTime)

    // 性能提升
    improvement := float64(noCacheTime) / float64(cacheTime)
    fmt.Printf("\n性能提升: %.2fx\n", improvement)

    // 缓存统计
    stats := cache.Stats()
    fmt.Printf("缓存命中率: %.2f%%\n", stats.HitRate()*100)
}

func generateLargeJSON(size int) []byte {
    var items []string
    for i := 0; i < size; i++ {
        items = append(items, fmt.Sprintf(
            `{"id": %d, "name": "Item%d", "value": %f, "tags": ["tag%d", "tag%d"]}`,
            i, i, rand.Float64()*1000, i%10, i%20,
        ))
    }
    return []byte(fmt.Sprintf(`{"items": [%s], "total": %d}`,
        strings.Join(items, ","), size))
}

func processData(node fxjson.Node) {
    // 模拟数据处理
    count := node.GetPath("items").Len()
    _ = count
}
```

## 缓存策略

### LRU 淘汰策略

```go
func lruCacheExample() {
    // 创建小容量缓存以观察 LRU 行为
    cache := fxjson.NewMemoryCache(3) // 最多3个项目

    // 添加项目
    items := []string{"A", "B", "C", "D", "E"}

    for _, item := range items {
        data := fmt.Sprintf(`{"item": "%s", "timestamp": %d}`,
            item, time.Now().UnixNano())

        node := fxjson.FromString(data)
        cache.Set(item, node, time.Minute)

        stats := cache.Stats()
        fmt.Printf("添加 %s 后: 大小=%d, 淘汰=%d\n",
            item, stats.Size, stats.Evictions)
    }

    // 访问顺序检查
    testItems := []string{"C", "D", "E", "A", "B"}
    for _, item := range testItems {
        if cached, found := cache.Get(item); found {
            fmt.Printf("找到 %s: %s\n", item,
                cached.Get("item").StringOr(""))
        } else {
            fmt.Printf("未找到 %s (已被淘汰)\n", item)
        }
    }
}
```

### 分层缓存

```go
func tieredCache() {
    // L1 缓存：热数据，小容量快速访问
    l1Cache := fxjson.NewMemoryCache(100)

    // L2 缓存：温数据，中等容量
    l2Cache := fxjson.NewMemoryCache(1000)

    // 自定义缓存策略
    type TieredCache struct {
        l1 *fxjson.MemoryCache
        l2 *fxjson.MemoryCache
    }

    tiered := &TieredCache{
        l1: l1Cache,
        l2: l2Cache,
    }

    // 实现获取逻辑
    getFromTiered := func(key string) (fxjson.Node, bool) {
        // 先从 L1 获取
        if node, found := tiered.l1.Get(key); found {
            return node, true
        }

        // 再从 L2 获取
        if node, found := tiered.l2.Get(key); found {
            // 提升到 L1
            tiered.l1.Set(key, node, time.Minute)
            return node, true
        }

        return fxjson.Node{}, false
    }

    // 使用示例
    jsonData := `{"id": 1, "name": "测试数据"}`
    node := fxjson.FromString(jsonData)

    // 存储到 L2
    tiered.l2.Set("test:1", node, time.Hour)

    // 获取时会提升到 L1
    if cached, found := getFromTiered("test:1"); found {
        fmt.Printf("获取成功: %s\n", cached.Get("name").StringOr(""))

        // 检查 L1 是否包含
        if _, found := tiered.l1.Get("test:1"); found {
            fmt.Println("数据已提升到 L1 缓存")
        }
    }
}
```

## 实际应用场景

### 1. API 响应缓存

```go
func apiResponseCache() {
    // 模拟 API 响应缓存
    cache := fxjson.NewMemoryCache(500)
    fxjson.EnableCaching(cache)

    // API 响应数据
    responses := map[string][]byte{
        "/api/users":     []byte(`{"users": [{"id": 1, "name": "张三"}]}`),
        "/api/products":  []byte(`{"products": [{"id": 101, "name": "产品A"}]}`),
        "/api/orders":    []byte(`{"orders": [{"id": 1001, "total": 99.99}]}`),
    }

    // 模拟 API 请求
    handleAPIRequest := func(endpoint string) fxjson.Node {
        // 尝试从缓存获取
        if cached, found := cache.Get(endpoint); found {
            fmt.Printf("缓存命中: %s\n", endpoint)
            return cached
        }

        // 模拟 API 调用延迟
        time.Sleep(50 * time.Millisecond)

        // 解析响应
        node := fxjson.FromBytes(responses[endpoint])

        // 缓存结果
        cache.Set(endpoint, node, 5*time.Minute)

        fmt.Printf("API 调用: %s\n", endpoint)
        return node
    }

    // 测试请求
    endpoints := []string{"/api/users", "/api/products", "/api/orders"}

    // 第一轮请求
    fmt.Println("=== 第一轮请求 ===")
    start := time.Now()
    for _, endpoint := range endpoints {
        handleAPIRequest(endpoint)
    }
    firstRound := time.Since(start)

    // 第二轮请求（从缓存）
    fmt.Println("\n=== 第二轮请求 ===")
    start = time.Now()
    for _, endpoint := range endpoints {
        handleAPIRequest(endpoint)
    }
    secondRound := time.Since(start)

    fmt.Printf("\n第一轮耗时: %v\n", firstRound)
    fmt.Printf("第二轮耗时: %v (提升 %.2fx)\n",
        secondRound, float64(firstRound)/float64(secondRound))
}
```

### 2. 配置文件缓存

```go
func configCache() {
    // 配置文件缓存系统
    type ConfigManager struct {
        cache  *fxjson.MemoryCache
        files  map[string]time.Time
        mutex  sync.RWMutex
    }

    manager := &ConfigManager{
        cache: fxjson.NewMemoryCache(100),
        files: make(map[string]time.Time),
    }

    loadConfig := func(filename string) (fxjson.Node, error) {
        manager.mutex.RLock()
        lastMod, exists := manager.files[filename]
        manager.mutex.RUnlock()

        // 检查文件是否修改
        if info, err := os.Stat(filename); err == nil {
            if exists && info.ModTime().Before(lastMod) {
                // 文件未修改，使用缓存
                if cached, found := manager.cache.Get(filename); found {
                    return cached, nil
                }
            }
        }

        // 读取并解析文件
        data, err := os.ReadFile(filename)
        if err != nil {
            return fxjson.Node{}, err
        }

        node := fxjson.FromBytes(data)

        // 更新缓存
        manager.cache.Set(filename, node, time.Hour)

        manager.mutex.Lock()
        manager.files[filename] = time.Now()
        manager.mutex.Unlock()

        return node, nil
    }

    // 使用示例
    node, err := loadConfig("config.json")
    if err == nil {
        debug := node.GetPath("debug.enabled").BoolOr(false)
        fmt.Printf("Debug mode: %v\n", debug)
    }
}
```

### 3. 数据库查询缓存

```go
func dbQueryCache() {
    // 数据库查询结果缓存
    cache := fxjson.NewMemoryCache(1000)

    queryAndCache := func(query string, params ...interface{}) (fxjson.Node, error) {
        // 生成缓存键
        cacheKey := fmt.Sprintf("query:%x", md5.Sum([]byte(fmt.Sprintf(query, params...))))

        // 尝试从缓存获取
        if cached, found := cache.Get(cacheKey); found {
            fmt.Printf("查询缓存命中: %s\n", query)
            return cached, nil
        }

        // 模拟数据库查询
        time.Sleep(100 * time.Millisecond)

        // 模拟查询结果
        result := map[string]interface{}{
            "rows": []map[string]interface{}{
                {"id": 1, "name": "张三", "score": 95.5},
                {"id": 2, "name": "李四", "score": 87.0},
            },
            "total": 2,
        }

        // 转换为 JSON Node
        jsonData, _ := fxjson.Marshal(result)
        node := fxjson.FromBytes(jsonData)

        // 缓存结果
        cache.Set(cacheKey, node, 10*time.Minute)

        fmt.Printf("数据库查询: %s\n", query)
        return node, nil
    }

    // 测试查询
    queries := []string{
        "SELECT * FROM users WHERE score > 90",
        "SELECT * FROM users WHERE department = '研发部'",
        "SELECT * FROM users WHERE active = true",
    }

    // 执行查询
    for _, query := range queries {
        node, _ := queryAndCache(query)
        total := node.Get("total").IntOr(0)
        fmt.Printf("  查询结果: %d 行\n", total)
    }

    // 重复执行（从缓存）
    fmt.Println("\n=== 重复查询 ===")
    for _, query := range queries {
        node, _ := queryAndCache(query)
        total := node.Get("total").IntOr(0)
        fmt.Printf("  查询结果: %d 行\n", total)
    }
}
```

## 总结

FxJSON 的缓存系统提供了：

1. **自动节点缓存**：无需配置，自动优化重复访问
2. **灵活的内存缓存**：可配置的缓存策略和容量
3. **懒加载机制**：按需加载，节省资源
4. **性能监控**：详细的缓存统计和性能指标
5. **线程安全**：支持高并发场景

通过合理使用缓存功能，可以显著提升应用程序的性能，特别是在重复访问相同数据的场景下。