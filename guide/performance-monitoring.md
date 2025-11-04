# 性能监控指南

FxJSON 提供了完善的性能监控工具，帮助您了解和优化 JSON 处理性能，实现数据驱动的性能优化。

## 目录

- [性能监控概述](#性能监控概述)
- [性能指标收集](#性能指标收集)
- [性能分析工具](#性能分析工具)
- [性能优化建议](#性能优化建议)
- [基准测试](#基准测试)
- [生产环境监控](#生产环境监控)

---

## 性能监控概述

### 为什么需要性能监控？

1. **识别瓶颈**: 找出性能热点
2. **量化改进**: 验证优化效果
3. **容量规划**: 预测资源需求
4. **问题诊断**: 快速定位性能问题

### 监控指标体系

```
性能监控
├── 解析性能
│   ├── 解析耗时
│   ├── 吞吐量
│   └── 内存分配
├── 访问性能
│   ├── Get/GetPath 耗时
│   ├── 缓存命中率
│   └── 遍历速度
└── 序列化性能
    ├── Marshal 耗时
    ├── 内存使用
    └── GC 压力
```

---

## 性能指标收集

### FromBytesWithMetrics()

自动收集解析性能指标。

```go
func FromBytesWithMetrics(b []byte) Node
```

### GetPerformanceStats()

获取累积的性能统计数据。

```go
func GetPerformanceStats() map[string]interface{}
```

### 基础监控示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/icloudza/fxjson"
)

func main() {
    // 生成测试数据
    testData := generateTestJSON(1000)

    // 使用性能监控解析
    start := time.Now()
    node := fxjson.FromBytesWithMetrics([]byte(testData))
    parseTime := time.Since(start)

    // 执行一些操作
    for i := 0; i < 100; i++ {
        _ = node.GetPath(fmt.Sprintf("users.%d.name", i%100)).StringOr("")
    }

    // 获取性能统计
    stats := fxjson.GetPerformanceStats()

    fmt.Println("=== 性能统计 ===")
    fmt.Printf("解析耗时: %v\n", parseTime)
    fmt.Printf("数据大小: %d 字节\n", len(testData))
    fmt.Printf("解析速度: %.2f MB/s\n",
        float64(len(testData))/parseTime.Seconds()/1024/1024)

    for key, value := range stats {
        fmt.Printf("%s: %v\n", key, value)
    }
}

func generateTestJSON(userCount int) string {
    var users []string
    for i := 0; i < userCount; i++ {
        user := fmt.Sprintf(`{
            "id": %d,
            "name": "用户%d",
            "email": "user%d@example.com",
            "age": %d,
            "active": %t
        }`, i, i, i, 20+i%50, i%2 == 0)
        users = append(users, user)
    }
    return fmt.Sprintf(`{"users": [%s]}`, strings.Join(users, ","))
}
```

### 详细性能指标

```go
type PerformanceMetrics struct {
    TotalParses       int64         // 总解析次数
    TotalParseTime    time.Duration // 总解析耗时
    AvgParseTime      time.Duration // 平均解析耗时
    TotalBytes        int64         // 总处理字节数
    CacheHits         int64         // 缓存命中次数
    CacheMisses       int64         // 缓存未命中次数
    CacheHitRate      float64       // 缓存命中率
    TotalAllocations  int64         // 总内存分配次数
    TotalAllocBytes   int64         // 总分配字节数
    GCCount           uint32        // GC 次数
}

func collectDetailedMetrics() {
    var m1, m2 runtime.MemStats

    // 初始状态
    runtime.ReadMemStats(&m1)
    startTime := time.Now()

    // 执行大量操作
    for i := 0; i < 1000; i++ {
        jsonData := fmt.Sprintf(`{"id": %d, "value": %d}`, i, i*100)
        node := fxjson.FromBytesWithMetrics([]byte(jsonData))
        _ = node.Get("value").IntOr(0)
    }

    // 最终状态
    runtime.ReadMemStats(&m2)
    totalTime := time.Since(startTime)

    // 获取性能统计
    stats := fxjson.GetPerformanceStats()

    metrics := PerformanceMetrics{
        TotalParses:      1000,
        TotalParseTime:   totalTime,
        AvgParseTime:     totalTime / 1000,
        TotalAllocations: int64(m2.Mallocs - m1.Mallocs),
        TotalAllocBytes:  int64(m2.TotalAlloc - m1.TotalAlloc),
        GCCount:          m2.NumGC - m1.NumGC,
    }

    // 从统计中提取缓存信息
    if cacheHits, ok := stats["cache_hits"].(int64); ok {
        metrics.CacheHits = cacheHits
    }
    if cacheMisses, ok := stats["cache_misses"].(int64); ok {
        metrics.CacheMisses = cacheMisses
    }
    if metrics.CacheHits+metrics.CacheMisses > 0 {
        metrics.CacheHitRate = float64(metrics.CacheHits) /
            float64(metrics.CacheHits+metrics.CacheMisses) * 100
    }

    printMetrics(metrics)
}

func printMetrics(m PerformanceMetrics) {
    fmt.Println("=== 详细性能指标 ===")
    fmt.Printf("总解析次数: %d\n", m.TotalParses)
    fmt.Printf("总耗时: %v\n", m.TotalParseTime)
    fmt.Printf("平均耗时: %v\n", m.AvgParseTime)
    fmt.Printf("吞吐量: %.0f ops/sec\n",
        float64(m.TotalParses)/m.TotalParseTime.Seconds())

    fmt.Println("\n缓存性能:")
    fmt.Printf("  命中次数: %d\n", m.CacheHits)
    fmt.Printf("  未命中次数: %d\n", m.CacheMisses)
    fmt.Printf("  命中率: %.2f%%\n", m.CacheHitRate)

    fmt.Println("\n内存使用:")
    fmt.Printf("  分配次数: %d\n", m.TotalAllocations)
    fmt.Printf("  分配字节: %d KB\n", m.TotalAllocBytes/1024)
    fmt.Printf("  平均每次: %d 字节\n", m.TotalAllocBytes/m.TotalParses)
    fmt.Printf("  GC 次数: %d\n", m.GCCount)
}
```

---

## 性能分析工具

### 1. 解析性能分析

```go
func analyzeParsePerformance() {
    // 测试不同大小的 JSON
    sizes := []int{100, 1000, 10000, 50000}

    fmt.Println("=== 解析性能分析 ===")
    fmt.Printf("%-10s %-15s %-15s %-15s\n", "大小", "耗时", "速度", "内存")
    fmt.Println(strings.Repeat("-", 60))

    for _, size := range sizes {
        testData := generateTestJSON(size)
        dataSize := len(testData)

        var m1, m2 runtime.MemStats
        runtime.ReadMemStats(&m1)

        start := time.Now()
        node := fxjson.FromBytesWithMetrics([]byte(testData))
        parseTime := time.Since(start)

        runtime.ReadMemStats(&m2)
        memUsed := m2.Alloc - m1.Alloc

        speed := float64(dataSize) / parseTime.Seconds() / 1024 / 1024

        fmt.Printf("%-10d %-15v %-15.2f %-15d\n",
            size, parseTime, speed, memUsed/1024)

        // 确保node被使用
        _ = node.Exists()
    }
}
```

### 2. 访问性能分析

```go
func analyzeAccessPerformance() {
    jsonData := generateTestJSON(10000)
    node := fxjson.FromString(jsonData)

    accessPatterns := []struct {
        name string
        fn   func()
    }{
        {
            "直接访问",
            func() {
                for i := 0; i < 1000; i++ {
                    _ = node.Get("users").Index(i % 100).Get("name").StringOr("")
                }
            },
        },
        {
            "路径访问",
            func() {
                for i := 0; i < 1000; i++ {
                    _ = node.GetPath(fmt.Sprintf("users.%d.name", i%100)).StringOr("")
                }
            },
        },
        {
            "缓存访问",
            func() {
                // 重复访问相同路径，测试缓存效果
                for i := 0; i < 1000; i++ {
                    _ = node.GetPath("users.0.name").StringOr("")
                }
            },
        },
        {
            "遍历访问",
            func() {
                count := 0
                node.Get("users").ArrayForEach(func(i int, user fxjson.Node) bool {
                    _ = user.Get("name").StringOr("")
                    count++
                    return count < 1000
                })
            },
        },
    }

    fmt.Println("=== 访问性能对比 ===")
    fmt.Printf("%-15s %-15s %-15s\n", "访问方式", "耗时", "速度(ops/s)")
    fmt.Println(strings.Repeat("-", 50))

    for _, pattern := range accessPatterns {
        start := time.Now()
        pattern.fn()
        duration := time.Since(start)

        opsPerSec := 1000.0 / duration.Seconds()

        fmt.Printf("%-15s %-15v %-15.0f\n",
            pattern.name, duration, opsPerSec)
    }
}
```

### 3. 序列化性能分析

```go
func analyzeSerializationPerformance() {
    type User struct {
        ID       int      `json:"id"`
        Name     string   `json:"name"`
        Email    string   `json:"email"`
        Age      int      `json:"age"`
        Active   bool     `json:"active"`
        Tags     []string `json:"tags"`
    }

    // 生成测试数据
    users := make([]User, 1000)
    for i := 0; i < 1000; i++ {
        users[i] = User{
            ID:     i,
            Name:   fmt.Sprintf("用户%d", i),
            Email:  fmt.Sprintf("user%d@example.com", i),
            Age:    20 + i%50,
            Active: i%2 == 0,
            Tags:   []string{"tag1", "tag2", "tag3"},
        }
    }

    methods := []struct {
        name string
        fn   func() []byte
    }{
        {
            "标准序列化",
            func() []byte {
                data, _ := fxjson.Marshal(users)
                return data
            },
        },
        {
            "快速序列化",
            func() []byte {
                return fxjson.FastMarshal(users)
            },
        },
        {
            "标准库",
            func() []byte {
                data, _ := json.Marshal(users)
                return data
            },
        },
    }

    fmt.Println("=== 序列化性能对比 ===")
    fmt.Printf("%-15s %-15s %-15s %-15s\n", "方法", "耗时", "速度", "大小")
    fmt.Println(strings.Repeat("-", 65))

    for _, method := range methods {
        var m1, m2 runtime.MemStats
        runtime.ReadMemStats(&m1)

        start := time.Now()
        result := method.fn()
        duration := time.Since(start)

        runtime.ReadMemStats(&m2)
        memUsed := m2.TotalAlloc - m1.TotalAlloc

        speed := float64(len(result)) / duration.Seconds() / 1024 / 1024

        fmt.Printf("%-15s %-15v %-15.2f %-15d\n",
            method.name, duration, speed, len(result))
    }
}
```

### 4. 内存分配分析

```go
func analyzeMemoryAllocation() {
    fmt.Println("=== 内存分配分析 ===")

    operations := []struct {
        name string
        fn   func()
    }{
        {
            "解析操作",
            func() {
                for i := 0; i < 100; i++ {
                    jsonData := `{"id": 1, "name": "test", "value": 100}`
                    fxjson.FromString(jsonData)
                }
            },
        },
        {
            "访问操作",
            func() {
                node := fxjson.FromString(`{"user": {"name": "test", "age": 25}}`)
                for i := 0; i < 100; i++ {
                    _ = node.GetPath("user.name").StringOr("")
                }
            },
        },
        {
            "遍历操作",
            func() {
                node := fxjson.FromString(generateTestJSON(100))
                for i := 0; i < 10; i++ {
                    node.Get("users").ArrayForEach(func(idx int, user fxjson.Node) bool {
                        _ = user.Get("name").StringOr("")
                        return true
                    })
                }
            },
        },
    }

    fmt.Printf("%-15s %-15s %-15s %-15s\n",
        "操作", "分配次数", "分配大小", "平均每次")
    fmt.Println(strings.Repeat("-", 65))

    for _, op := range operations {
        var m1, m2 runtime.MemStats
        runtime.ReadMemStats(&m1)

        op.fn()

        runtime.ReadMemStats(&m2)

        allocCount := m2.Mallocs - m1.Mallocs
        allocBytes := m2.TotalAlloc - m1.TotalAlloc
        avgAlloc := int64(0)
        if allocCount > 0 {
            avgAlloc = int64(allocBytes) / int64(allocCount)
        }

        fmt.Printf("%-15s %-15d %-15d %-15d\n",
            op.name, allocCount, allocBytes, avgAlloc)
    }
}
```

---

## 性能优化建议

### 1. 自动性能分析

```go
func autoPerformanceAnalysis(jsonData []byte) {
    // 使用调试模式获取性能提示
    node, debugInfo := fxjson.FromBytesWithDebug(jsonData)

    fmt.Println("=== 自动性能分析 ===")
    fmt.Printf("数据大小: %d 字节\n", debugInfo.DataSize)
    fmt.Printf("解析耗时: %v\n", debugInfo.ParseTime)
    fmt.Printf("节点数量: %d\n", debugInfo.NodeCount)
    fmt.Printf("最大深度: %d\n", debugInfo.MaxDepth)

    if len(debugInfo.PerformanceHints) > 0 {
        fmt.Println("\n性能优化建议:")
        for i, hint := range debugInfo.PerformanceHints {
            fmt.Printf("%d. %s\n", i+1, hint)
        }
    }

    // 分析结构复杂度
    fmt.Println("\n结构复杂度:")
    fmt.Printf("  对象: %d\n", debugInfo.ObjectCount)
    fmt.Printf("  数组: %d\n", debugInfo.ArrayCount)
    fmt.Printf("  字符串: %d\n", debugInfo.StringCount)
    fmt.Printf("  数字: %d\n", debugInfo.NumberCount)

    // 提供具体建议
    if debugInfo.MaxDepth > 10 {
        fmt.Println("\n⚠️ 嵌套深度较深，建议:")
        fmt.Println("   - 考虑扁平化数据结构")
        fmt.Println("   - 使用 GetPath 而非多次 Get")
    }

    if debugInfo.ArrayCount > 100 {
        fmt.Println("\n⚠️ 数组较多，建议:")
        fmt.Println("   - 使用 ArrayForEach 而非 Index 遍历")
        fmt.Println("   - 考虑启用缓存优化")
    }

    // 确保 node 被使用
    _ = node.Exists()
}
```

### 2. 热点路径优化

```go
func optimizeHotPath() {
    jsonData := generateTestJSON(10000)

    fmt.Println("=== 热点路径优化对比 ===")

    // 优化前: 重复路径访问
    fmt.Println("\n优化前:")
    start := time.Now()
    node := fxjson.FromString(jsonData)
    for i := 0; i < 1000; i++ {
        _ = node.GetPath("users.0.name").StringOr("")
        _ = node.GetPath("users.0.email").StringOr("")
        _ = node.GetPath("users.0.age").IntOr(0)
    }
    beforeTime := time.Since(start)
    fmt.Printf("  耗时: %v\n", beforeTime)

    // 优化后: 缓存节点
    fmt.Println("\n优化后:")
    start = time.Now()
    node = fxjson.FromString(jsonData)
    firstUser := node.GetPath("users.0")  // 缓存父节点
    for i := 0; i < 1000; i++ {
        _ = firstUser.Get("name").StringOr("")
        _ = firstUser.Get("email").StringOr("")
        _ = firstUser.Get("age").IntOr(0)
    }
    afterTime := time.Since(start)
    fmt.Printf("  耗时: %v\n", afterTime)

    improvement := float64(beforeTime) / float64(afterTime)
    fmt.Printf("\n性能提升: %.2fx\n", improvement)
}
```

### 3. 批量操作优化

```go
func optimizeBatchOperations() {
    users := make([]map[string]interface{}, 1000)
    for i := 0; i < 1000; i++ {
        users[i] = map[string]interface{}{
            "id":    i,
            "name":  fmt.Sprintf("用户%d", i),
            "email": fmt.Sprintf("user%d@example.com", i),
        }
    }

    fmt.Println("=== 批量操作优化对比 ===")

    // 优化前: 逐个序列化
    fmt.Println("\n优化前 (逐个序列化):")
    start := time.Now()
    for _, user := range users {
        _ = fxjson.FastMarshal(user)
    }
    beforeTime := time.Since(start)
    fmt.Printf("  耗时: %v\n", beforeTime)

    // 优化后: 批量序列化
    fmt.Println("\n优化后 (批量序列化):")
    start = time.Now()
    _ = fxjson.FastMarshal(users)
    afterTime := time.Since(start)
    fmt.Printf("  耗时: %v\n", afterTime)

    improvement := float64(beforeTime) / float64(afterTime)
    fmt.Printf("\n性能提升: %.2fx\n", improvement)
}
```

---

## 基准测试

### 标准基准测试

```go
func BenchmarkParse(b *testing.B) {
    jsonData := []byte(generateTestJSON(100))

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        node := fxjson.FromBytes(jsonData)
        _ = node.Exists()
    }
}

func BenchmarkAccess(b *testing.B) {
    jsonData := generateTestJSON(100)
    node := fxjson.FromString(jsonData)

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _ = node.GetPath("users.0.name").StringOr("")
    }
}

func BenchmarkArrayForEach(b *testing.B) {
    jsonData := generateTestJSON(100)
    node := fxjson.FromString(jsonData)
    users := node.Get("users")

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        users.ArrayForEach(func(idx int, user fxjson.Node) bool {
            _ = user.Get("name").StringOr("")
            return true
        })
    }
}

func BenchmarkMarshal(b *testing.B) {
    type User struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }

    user := User{ID: 1, Name: "test", Email: "test@example.com"}

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _ = fxjson.FastMarshal(user)
    }
}
```

### 对比基准测试

```go
func BenchmarkComparison(b *testing.B) {
    type User struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
        Age   int    `json:"age"`
    }

    user := User{ID: 1, Name: "张三", Email: "zhang@example.com", Age: 28}
    jsonData, _ := json.Marshal(user)

    b.Run("FxJSON_Parse", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            node := fxjson.FromBytes(jsonData)
            _ = node.Exists()
        }
    })

    b.Run("StdLib_Parse", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            var u User
            _ = json.Unmarshal(jsonData, &u)
        }
    })

    b.Run("FxJSON_Marshal", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            _ = fxjson.FastMarshal(user)
        }
    })

    b.Run("StdLib_Marshal", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            _, _ = json.Marshal(user)
        }
    })
}
```

---

## 生产环境监控

### 监控指标导出

```go
type MetricsCollector struct {
    mu              sync.RWMutex
    parseCount      int64
    parseDuration   time.Duration
    errorCount      int64
    cacheHits       int64
    cacheMisses     int64
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{}
}

func (mc *MetricsCollector) RecordParse(duration time.Duration, success bool) {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    mc.parseCount++
    mc.parseDuration += duration
    if !success {
        mc.errorCount++
    }
}

func (mc *MetricsCollector) RecordCacheAccess(hit bool) {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    if hit {
        mc.cacheHits++
    } else {
        mc.cacheMisses++
    }
}

func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
    mc.mu.RLock()
    defer mc.mu.RUnlock()

    avgDuration := time.Duration(0)
    if mc.parseCount > 0 {
        avgDuration = mc.parseDuration / time.Duration(mc.parseCount)
    }

    cacheHitRate := 0.0
    totalAccess := mc.cacheHits + mc.cacheMisses
    if totalAccess > 0 {
        cacheHitRate = float64(mc.cacheHits) / float64(totalAccess) * 100
    }

    return map[string]interface{}{
        "parse_count":      mc.parseCount,
        "avg_parse_time":   avgDuration,
        "error_count":      mc.errorCount,
        "error_rate":       float64(mc.errorCount) / float64(mc.parseCount) * 100,
        "cache_hits":       mc.cacheHits,
        "cache_misses":     mc.cacheMisses,
        "cache_hit_rate":   cacheHitRate,
    }
}

func (mc *MetricsCollector) PrintMetrics() {
    metrics := mc.GetMetrics()

    fmt.Println("=== 运行时指标 ===")
    fmt.Printf("解析次数: %v\n", metrics["parse_count"])
    fmt.Printf("平均耗时: %v\n", metrics["avg_parse_time"])
    fmt.Printf("错误次数: %v\n", metrics["error_count"])
    fmt.Printf("错误率: %.2f%%\n", metrics["error_rate"])
    fmt.Printf("缓存命中: %v\n", metrics["cache_hits"])
    fmt.Printf("缓存未命中: %v\n", metrics["cache_misses"])
    fmt.Printf("缓存命中率: %.2f%%\n", metrics["cache_hit_rate"])
}
```

### 实时监控示例

```go
func productionMonitoring() {
    collector := NewMetricsCollector()

    // 模拟生产环境负载
    go func() {
        for i := 0; i < 10000; i++ {
            jsonData := fmt.Sprintf(`{"id": %d, "value": %d}`, i, i*100)

            start := time.Now()
            node := fxjson.FromString(jsonData)
            duration := time.Since(start)

            success := node.Exists()
            collector.RecordParse(duration, success)

            // 模拟缓存访问
            hit := i%3 == 0  // 33% 命中率
            collector.RecordCacheAccess(hit)

            time.Sleep(time.Millisecond)
        }
    }()

    // 定期输出指标
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    timeout := time.After(30 * time.Second)

    for {
        select {
        case <-ticker.C:
            fmt.Println("\n" + time.Now().Format("15:04:05"))
            collector.PrintMetrics()

        case <-timeout:
            fmt.Println("\n=== 最终统计 ===")
            collector.PrintMetrics()
            return
        }
    }
}
```

### Prometheus 集成示例

```go
// 假设使用 Prometheus client
/*
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    parseCounter = promauto.NewCounter(prometheus.CounterOpts{
        Name: "fxjson_parse_total",
        Help: "Total number of JSON parse operations",
    })

    parseDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "fxjson_parse_duration_seconds",
        Help:    "Duration of JSON parse operations",
        Buckets: prometheus.DefBuckets,
    })

    cacheHitCounter = promauto.NewCounter(prometheus.CounterOpts{
        Name: "fxjson_cache_hits_total",
        Help: "Total number of cache hits",
    })

    cacheMissCounter = promauto.NewCounter(prometheus.CounterOpts{
        Name: "fxjson_cache_misses_total",
        Help: "Total number of cache misses",
    })
)

func parseWithMetrics(jsonData []byte) fxjson.Node {
    start := time.Now()
    node := fxjson.FromBytes(jsonData)
    duration := time.Since(start)

    parseCounter.Inc()
    parseDuration.Observe(duration.Seconds())

    return node
}
*/
```

---

## 最佳实践

### 1. 持续监控

```go
// 在应用启动时初始化监控
func init() {
    // 定期收集性能统计
    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()

        for range ticker.C {
            stats := fxjson.GetPerformanceStats()
            logMetrics(stats)
        }
    }()
}

func logMetrics(stats map[string]interface{}) {
    log.Printf("FxJSON 性能指标: %+v", stats)
}
```

### 2. 性能回归测试

```go
// 在 CI/CD 中运行性能测试
func TestPerformanceRegression(t *testing.T) {
    threshold := 100 * time.Microsecond

    jsonData := generateTestJSON(100)

    start := time.Now()
    for i := 0; i < 100; i++ {
        node := fxjson.FromString(jsonData)
        _ = node.GetPath("users.0.name").StringOr("")
    }
    avgTime := time.Since(start) / 100

    if avgTime > threshold {
        t.Errorf("性能回归: 平均耗时 %v 超过阈值 %v", avgTime, threshold)
    }
}
```

### 3. 生产环境告警

```go
func setupAlerts(collector *MetricsCollector) {
    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()

        for range ticker.C {
            metrics := collector.GetMetrics()

            // 错误率告警
            if errorRate := metrics["error_rate"].(float64); errorRate > 5.0 {
                alert("错误率过高: %.2f%%", errorRate)
            }

            // 缓存命中率告警
            if hitRate := metrics["cache_hit_rate"].(float64); hitRate < 50.0 {
                alert("缓存命中率过低: %.2f%%", hitRate)
            }

            // 性能告警
            if avgTime := metrics["avg_parse_time"].(time.Duration); avgTime > 100*time.Microsecond {
                alert("平均解析时间过长: %v", avgTime)
            }
        }
    }()
}

func alert(format string, args ...interface{}) {
    message := fmt.Sprintf(format, args...)
    log.Printf("⚠️ 告警: %s", message)
    // 发送到告警系统 (PagerDuty, 钉钉等)
}
```

通过完善的性能监控体系，您可以持续优化 FxJSON 的使用，确保应用始终保持最佳性能状态。
