# 性能对比

FxJSON 专为高性能场景设计，在各种操作中都表现出显著的性能优势。本节详细分析 FxJSON 相对于 Go 标准库的性能提升。

## 核心操作性能

### 数据访问操作

| 操作 | FxJSON | 标准库 | 性能提升 | 内存优势 |
|------|--------|--------|----------|----------|
| Get | 24.88 ns | 2012 ns | **快80.8倍** | 0 vs 1984 B |
| GetPath | 111.5 ns | 2055 ns | **快18.4倍** | 0 vs 1984 B |
| Int转换 | 16.70 ns | 2026 ns | **快121.3倍** | 0 vs 1984 B |
| Float转换 | 7.688 ns | 2051 ns | **快266.7倍** | 0 vs 1984 B |
| Bool转换 | 3.684 ns | 2149 ns | **快583.2倍** | 0 vs 1984 B |
| String访问 | 5.402 ns | 2083 ns | **快385.6倍** | 0 vs 1984 B |
| 数组长度 | 20.70 ns | 2152 ns | **快103.9倍** | 0 vs 1984 B |
| 数组索引 | 18.42 ns | 2134 ns | **快115.9倍** | 0 vs 1984 B |
| 键存在检查 | 0.2454 ns | 2110 ns | **快8598倍** | 0 vs 1984 B |

### 遍历操作性能

| 操作 | FxJSON | 标准库 | 性能提升 | 内存优势 |
|------|--------|--------|----------|----------|
| 对象遍历 | 108.9 ns | 2142 ns | **快19.7倍** | 0 vs 1984 B |
| 数组遍历 | 30.21 ns | 2119 ns | **快70.2倍** | 0 vs 1984 B |
| 深度遍历 | 1536 ns | 2891 ns | **快1.9倍** | 3056 vs 2289 B |
| 复杂遍历 | 1310 ns | 3505 ns | **快2.7倍** | 0 vs 4136 B |
| 大数据遍历 | 12.8 µs | 17.4 µs | **快1.4倍** | 19136 vs 14698 B |

## 性能亮点

### 1. 零分配核心操作

FxJSON 的核心优势是在大多数操作中实现了零内存分配：

```go
// 标准库方式 - 每次都会分配内存
var result interface{}
json.Unmarshal(data, &result)
value := result.(map[string]interface{})["name"].(string)

// FxJSON - 零分配直接访问
node := fxjson.FromBytes(data)
name := node.Get("name").StringOr("") // 0 分配
```

### 2. 缓存加速

重复访问同一数据时，FxJSON 提供显著的性能提升：

| 操作类型 | 首次访问 | 缓存访问 | 性能提升 |
|----------|----------|----------|----------|
| 基础解析 | 5,290 ns | **641.8 ns** | **8.2倍** |
| 内存使用 | 6,448 B | **20 B** | **99.7%减少** |
| 分配次数 | 45 allocs | **2 allocs** | **95.6%减少** |

### 3. 高效的数组遍历

数组遍历是 FxJSON 最突出的性能优势：

```go
// 性能测试代码示例
func BenchmarkArrayTraversal(b *testing.B) {
    largeArray := generateLargeArray(10000) // 10k元素数组
    
    // FxJSON 遍历
    b.Run("FxJSON", func(b *testing.B) {
        node := fxjson.FromBytes(largeArray)
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            node.ArrayForEach(func(index int, value fxjson.Node) bool {
                _ = value.Get("name").StringOr("")
                return true
            })
        }
    })
    
    // 标准库遍历
    b.Run("Standard", func(b *testing.B) {
        var data []map[string]interface{}
        json.Unmarshal(largeArray, &data)
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            for _, item := range data {
                _ = item["name"].(string)
            }
        }
    })
}
```

**结果：FxJSON 比标准库快 67.5 倍**

## 高级功能性能

### 查询和过滤

| 功能特性 | 操作耗时 | 内存使用 | 分配次数 | 说明 |
|----------|----------|----------|----------|------|
| 简单查询 | 3,386 ns | 640 B | 14 allocs | 基础过滤 |
| 复杂查询 | 4,986 ns | 1,720 B | 52 allocs | 多条件查询和排序 |
| 数据聚合 | 4,804 ns | 2,640 B | 32 allocs | 统计计算 |
| 数据变换 | 478.7 ns | 368 B | 5 allocs | 字段映射和类型转换 |
| 数据验证 | 216.6 ns | 360 B | 4 allocs | 基于规则的验证 |
| 流式处理 | 3,250 ns | 0 B | 0 allocs | 零分配流式数据处理 |

### 序列化性能

| 操作 | 时间 | 内存 | 分配次数 | 性能说明 |
|------|------|------|----------|----------|
| Marshal | 652.1 ns | 424 B | 9 allocs | 标准序列化 |
| **FastMarshal** | **226.7 ns** | **136 B** | **2 allocs** | **高性能序列化，快2.9倍** |
| StructMarshal | 267.1 ns | 136 B | 2 allocs | 直接结构体序列化 |

## 实际应用场景性能

### Web API 响应处理

```go
// 典型的 API 响应处理场景
func processAPIResponse(responseData []byte) {
    // FxJSON 处理
    start := time.Now()
    node := fxjson.FromBytes(responseData)
    users := node.GetPath("data.users")
    
    activeUsers := 0
    users.ArrayForEach(func(index int, user fxjson.Node) bool {
        if user.Get("active").BoolOr(false) {
            activeUsers++
        }
        return true
    })
    fxjsonTime := time.Since(start)
    
    // 标准库处理
    start = time.Now()
    var response map[string]interface{}
    json.Unmarshal(responseData, &response)
    
    data := response["data"].(map[string]interface{})
    usersList := data["users"].([]interface{})
    
    activeUsersStd := 0
    for _, userInterface := range usersList {
        user := userInterface.(map[string]interface{})
        if active, ok := user["active"].(bool); ok && active {
            activeUsersStd++
        }
    }
    stdTime := time.Since(start)
    
    fmt.Printf("FxJSON: %v, 标准库: %v, 性能提升: %.1fx\n", 
        fxjsonTime, stdTime, float64(stdTime)/float64(fxjsonTime))
}
```

**典型结果：FxJSON 比标准库快 15-25 倍**

### 配置文件解析

```go
// 大型配置文件解析场景
func parseConfig(configData []byte) {
    // 测试多次访问配置项的性能
    iterations := 1000
    
    // FxJSON 测试
    start := time.Now()
    node := fxjson.FromBytes(configData).EnableCache()
    for i := 0; i < iterations; i++ {
        _ = node.GetPath("server.database.host").StringOr("")
        _ = node.GetPath("server.database.port").IntOr(5432)
        _ = node.GetPath("server.redis.enabled").BoolOr(false)
    }
    fxjsonTime := time.Since(start)
    
    // 标准库测试
    start = time.Now()
    var config map[string]interface{}
    json.Unmarshal(configData, &config)
    for i := 0; i < iterations; i++ {
        server := config["server"].(map[string]interface{})
        database := server["database"].(map[string]interface{})
        _ = database["host"].(string)
        _ = int(database["port"].(float64))
        redis := server["redis"].(map[string]interface{})
        _ = redis["enabled"].(bool)
    }
    stdTime := time.Since(start)
    
    fmt.Printf("配置访问 - FxJSON: %v, 标准库: %v, 提升: %.1fx\n", 
        fxjsonTime, stdTime, float64(stdTime)/float64(fxjsonTime))
}
```

**典型结果：启用缓存后，FxJSON 比标准库快 30-50 倍**

## 内存使用分析

### 内存分配对比

FxJSON 的内存优势主要体现在：

1. **零分配操作**：大多数查询和访问操作不产生内存分配
2. **智能缓存**：缓存机制大幅减少重复解析的内存开销
3. **原地操作**：直接在原始数据上操作，避免数据复制

```go
// 内存使用对比示例
func memoryComparison() {
    data := generateLargeJSON(100000) // 100k 条记录
    
    // FxJSON 内存使用
    var m1 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    node := fxjson.FromBytes(data)
    node.ArrayForEach(func(index int, item fxjson.Node) bool {
        _ = item.Get("name").StringOr("")
        return index < 1000 // 只处理前1000条
    })
    
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)
    fxjsonMemory := m2.TotalAlloc - m1.TotalAlloc
    
    // 标准库内存使用
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    var items []map[string]interface{}
    json.Unmarshal(data, &items)
    for i, item := range items {
        if i >= 1000 {
            break
        }
        _ = item["name"].(string)
    }
    
    runtime.ReadMemStats(&m2)
    stdMemory := m2.TotalAlloc - m1.TotalAlloc
    
    fmt.Printf("内存使用 - FxJSON: %d bytes, 标准库: %d bytes\n", 
        fxjsonMemory, stdMemory)
    fmt.Printf("内存节省: %.1f%%\n", 
        (1.0-float64(fxjsonMemory)/float64(stdMemory))*100)
}
```

## 性能优化建议

### 1. 启用缓存

对于需要重复访问的数据，启用缓存可以获得显著性能提升：

```go
// 启用缓存
node := fxjson.FromBytes(data).EnableCache()

// 重复访问会更快
for i := 0; i < 1000; i++ {
    _ = node.GetPath("deep.nested.value").StringOr("")
}
```

### 2. 使用 ArrayForEach 而不是索引访问

```go
// 推荐：零分配遍历
users.ArrayForEach(func(index int, user fxjson.Node) bool {
    processUser(user)
    return true
})

// 不推荐：多次索引访问
for i := 0; i < users.ArrayLen(); i++ {
    user := users.Index(i) // 每次都有开销
    processUser(user)
}
```

### 3. 使用 Or 系列方法

```go
// 推荐：安全且高效
name := node.Get("name").StringOr("默认")

// 不推荐：需要错误处理
name, err := node.Get("name").String()
if err != nil {
    name = "默认"
}
```

### 4. 批量操作

```go
// 推荐：批量解码
var users []User
err := node.Get("users").DecodeStructFast(&users)

// 不推荐：逐个解码
users := make([]User, 0)
node.Get("users").ArrayForEach(func(index int, userNode fxjson.Node) bool {
    var user User
    userNode.DecodeStruct(&user)
    users = append(users, user)
    return true
})
```

## 基准测试环境

所有性能测试结果基于以下环境：

- **硬件**: Apple M4 Pro
- **Go版本**: Go 1.24.6
- **测试方法**: go test -bench=. -benchmem
- **数据集**: 包含各种复杂度的真实 JSON 数据

测试代码和详细基准测试可以在 [GitHub仓库](https://github.com/icloudza/fxjson) 中找到。