# 批量操作指南

FxJSON 提供了强大的批量操作功能，可以高效处理大量 JSON 数据。通过并发处理、流式操作和智能缓存，批量操作性能可以提升 10-50 倍。

## 目录

- [批量序列化](#批量序列化)
- [批量解析](#批量解析)
- [批处理器](#批处理器)
- [并发批量操作](#并发批量操作)
- [批量转换](#批量转换)
- [批量验证](#批量验证)
- [性能优化](#性能优化)
- [实际应用场景](#实际应用场景)

## 批量序列化

### BatchMarshalStructs()

一次性批量序列化多个结构体。

```go
func BatchMarshalStructs(structs []interface{}) ([][]byte, error)
```

**基础用法**：
```go
package main

import (
    "fmt"
    "time"
    "github.com/icloudza/fxjson"
)

type User struct {
    ID       int      `json:"id"`
    Name     string   `json:"name"`
    Email    string   `json:"email"`
    Active   bool     `json:"active"`
    Tags     []string `json:"tags,omitempty"`
}

func main() {
    // 准备批量数据
    items := []interface{}{
        User{ID: 1, Name: "张三", Email: "zhangsan@example.com", Active: true},
        User{ID: 2, Name: "李四", Email: "lisi@example.com", Active: false},
        User{ID: 3, Name: "王五", Email: "wangwu@example.com", Active: true},
        map[string]interface{}{
            "type":    "config",
            "version": "1.0.0",
            "debug":   true,
        },
        "简单字���串",
        12345,
    }

    // 批量序列化
    start := time.Now()
    results, err := fxjson.BatchMarshalStructs(items)
    if err != nil {
        panic(err)
    }
    duration := time.Since(start)

    fmt.Printf("批量序列化完成，耗时: %v\n", duration)
    fmt.Printf("处理了 %d 个项目\n\n", len(results))

    // 输出结果
    for i, data := range results {
        fmt.Printf("Item %d: %s\n", i+1, data)
    }
}
```

### BatchMarshalStructsWithOptions()

带选项的批量序列化。

```go
func BatchMarshalStructsWithOptions(structs []interface{}, opts SerializeOptions) ([][]byte, error)
```

```go
func batchWithOptions() {
    items := []interface{}{
        User{ID: 1, Name: "张三", Email: "zhangsan@example.com", Active: true},
        User{ID: 2, Name: "李四", Email: "lisi@example.com", Active: false},
    }

    // 使用美化选项
    options := fxjson.SerializeOptions{
        Indent:       2,  // 2个空格缩进
        EscapeHTML:   true,
        SortKeys:     true,
        OmitEmpty:    true,
        Precision:    2,
    }

    results, err := fxjson.BatchMarshalStructsWithOptions(items, options)
    if err != nil {
        panic(err)
    }

    for i, data := range results {
        fmt.Printf("美化输出 %d:\n%s\n", i+1, data)
    }
}
```

## 批量解析

### 批量解析 JSON 字符串

```go
func batchParseJSON() {
    // 准备多个 JSON 字符串
    jsonStrings := []string{
        `{"id": 1, "name": "张三", "age": 30}`,
        `{"id": 2, "name": "李四", "age": 25}`,
        `{"id": 3, "name": "王五", "age": 35}`,
        `{"users": [{"name": "A"}, {"name": "B"}]}`,
        `{"config": {"debug": true, "port": 8080}}`,
    }

    // 批量解析
    nodes := make([]fxjson.Node, len(jsonStrings))
    start := time.Now()

    for i, jsonStr := range jsonStrings {
        nodes[i] = fxjson.FromString(jsonStr)
    }

    duration := time.Since(start)
    fmt.Printf("批量解析 %d 个 JSON 耗时: %v\n", len(jsonStrings), duration)

    // 批量提取数据
    for i, node := range nodes {
        name := node.Get("name").StringOr("未知")
        age := node.Get("age").IntOr(0)
        fmt.Printf("Node %d: name=%s, age=%d\n", i+1, name, age)
    }
}
```

### 并发批量解析

```go
import "sync"

func concurrentBatchParse() {
    jsonStrings := make([]string, 10000)
    for i := 0; i < 10000; i++ {
        jsonStrings[i] = fmt.Sprintf(`{"id": %d, "name": "User%d", "active": %t}`,
            i+1, i+1, i%2 == 0)
    }

    var nodes []fxjson.Node
    var mu sync.Mutex
    var wg sync.WaitGroup

    // 使用工作池模式并发解析
    workerCount := 8
    batchSize := len(jsonStrings) / workerCount

    start := time.Now()

    for i := 0; i < workerCount; i++ {
        start := i * batchSize
        end := start + batchSize
        if i == workerCount-1 {
            end = len(jsonStrings)
        }

        wg.Add(1)
        go func(batch []string) {
            defer wg.Done()
            localNodes := make([]fxjson.Node, len(batch))

            for j, jsonStr := range batch {
                localNodes[j] = fxjson.FromString(jsonStr)
            }

            mu.Lock()
            nodes = append(nodes, localNodes...)
            mu.Unlock()
        }(jsonStrings[start:end])
    }

    wg.Wait()
    duration := time.Since(start)

    fmt.Printf("并发批量解析完成:\n")
    fmt.Printf("- 处理数量: %d\n", len(nodes))
    fmt.Printf("- 耗时: %v\n", duration)
    fmt.Printf("- 平均速度: %.0f items/sec\n", float64(len(nodes))/duration.Seconds())
}
```

## 批处理器

### NewBatchProcessor()

创建批处理器，自动处理批量数据。

```go
func NewBatchProcessor(batchSize int, processor func([]Node) error) *BatchProcessor
```

```go
func batchProcessorExample() {
    // 创建批处理器
    batchSize := 100
    processor := func(batch []fxjson.Node) error {
        // 处理批量数据
        fmt.Printf("处理批次: %d 个项目\n", len(batch))

        // 批量插入数据库
        for _, node := range batch {
            name := node.Get("name").StringOr("")
            // 插入操作...
            _ = name
        }

        return nil
    }

    bp := fxjson.NewBatchProcessor(batchSize, processor)

    // 添加数据到批处理器
    for i := 0; i < 350; i++ {
        jsonData := fmt.Sprintf(`{"id": %d, "name": "User%d"}`, i+1, i+1)
        node := fxjson.FromString(jsonData)

        err := bp.Add(node)
        if err != nil {
            fmt.Printf("添加失败: %v\n", err)
        }
    }

    // 刷新剩余数据
    err := bp.Flush()
    if err != nil {
        fmt.Printf("刷新失败: %v\n", err)
    }

    fmt.Println("批处理完成")
}
```

### 自定义批处理器

```go
type UserBatchProcessor struct {
    db *sql.DB
    batchSize int
    buffer    []User
}

func NewUserBatchProcessor(db *sql.DB, batchSize int) *UserBatchProcessor {
    return &UserBatchProcessor{
        db:        db,
        batchSize: batchSize,
        buffer:    make([]User, 0, batchSize),
    }
}

func (p *UserBatchProcessor) Add(user User) error {
    p.buffer = append(p.buffer, user)

    if len(p.buffer) >= p.batchSize {
        return p.Flush()
    }

    return nil
}

func (p *UserBatchProcessor) Flush() error {
    if len(p.buffer) == 0 {
        return nil
    }

    // 批量插入
    query := `INSERT INTO users (id, name, email, active) VALUES `
    values := make([]string, len(p.buffer))
    args := make([]interface{}, len(p.buffer)*4)

    for i, user := range p.buffer {
        values[i] = "(?, ?, ?, ?)"
        args[i*4] = user.ID
        args[i*4+1] = user.Name
        args[i*4+2] = user.Email
        args[i*4+3] = user.Active
    }

    query += strings.Join(values, ",")
    _, err := p.db.Exec(query, args...)

    if err == nil {
        p.buffer = p.buffer[:0] // 清空缓冲区
    }

    return err
}

func customBatchProcessor() {
    // 模拟数据库连接
    // db, _ := sql.Open("mysql", "dsn")
    // processor := NewUserBatchProcessor(db, 500)

    // for i := 0; i < 1500; i++ {
    //     user := User{
    //         ID:     i + 1,
    //         Name:   fmt.Sprintf("User%d", i+1),
    //         Email:  fmt.Sprintf("user%d@example.com", i+1),
    //         Active: true,
    //     }
    //     processor.Add(user)
    // }
    // processor.Flush()
}
```

## 并发批量操作

### BatchMarshalStructsConcurrent()

并发批量序列化，充分利用多核 CPU。

```go
func BatchMarshalStructsConcurrent(structs []interface{}, workers int) ([][]byte, error)
```

```go
func concurrentBatchMarshal() {
    // 准备大量数据
    var users []interface{}
    for i := 0; i < 100000; i++ {
        users = append(users, User{
            ID:     i + 1,
            Name:   fmt.Sprintf("User%d", i+1),
            Email:  fmt.Sprintf("user%d@example.com", i+1),
            Active: i%2 == 0,
            Tags:   []string{"tag1", "tag2", "tag3"},
        })
    }

    // 测试不同并发数的性能
    workerCounts := []int{1, 2, 4, 8, 16}

    for _, workers := range workerCounts {
        start := time.Now()
        results, err := fxjson.BatchMarshalStructsConcurrent(users, workers)
        if err != nil {
            panic(err)
        }
        duration := time.Since(start)

        fmt.Printf("Workers=%d: 处理 %d 项，耗时 %v，速度 %.0f items/sec\n",
            workers, len(results), duration,
            float64(len(results))/duration.Seconds())
    }
}
```

### NewBatchMarshaler()

创建可重用的批量序列化器。

```go
func NewBatchMarshaler(opts SerializeOptions, workers int) *BatchMarshaler
```

```go
func reusableBatchMarshaler() {
    // 创建批量序列化器
    options := fxjson.SerializeOptions{
        Indent:     0,  // 紧凑模式
        EscapeHTML: false,
        Precision:  -1,
    }

    batcher := fxjson.NewBatchMarshaler(options, 4) // 4个工作协程
    defer batcher.Close()

    // 批量处理多个批次
    for batch := 0; batch < 10; batch++ {
        // 准备批次数据
        var batchData []User
        for i := 0; i < 1000; i++ {
            batchData = append(batchData, User{
                ID:     batch*1000 + i + 1,
                Name:   fmt.Sprintf("User%d", batch*1000+i+1),
                Active: true,
            })
        }

        // 序列化批次
        start := time.Now()
        var results [][]byte
        for _, user := range batchData {
            data, err := batcher.MarshalSlice(user)
            if err != nil {
                log.Printf("序列化失败: %v", err)
                continue
            }
            results = append(results, data)
        }
        duration := time.Since(start)

        fmt.Printf("批次 %d: 序列化 %d 项，耗时 %v\n",
            batch+1, len(results), duration)
    }

    fmt.Println("所有批次处理完成")
}
```

### 并发批量解析

```go
func concurrentBatchParsing() {
    // 大量JSON数据
    jsonDocuments := generateJSONDocuments(10000)
    
    // 设置并发数
    concurrency := runtime.NumCPU()
    semaphore := make(chan struct{}, concurrency)
    
    var wg sync.WaitGroup
    resultChan := make(chan User, len(jsonDocuments))
    
    start := time.Now()
    
    for i, jsonData := range jsonDocuments {
        wg.Add(1)
        go func(index int, data []byte) {
            defer wg.Done()
            
            // 限制并发数
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            // 解析JSON
            node := fxjson.FromBytes(data)
            var user User
            if err := node.Decode(&user); err == nil {
                resultChan <- user
            }
        }(i, jsonData)
    }
    
    // 等待所有goroutine完成
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    // 收集结果
    var users []User
    for user := range resultChan {
        users = append(users, user)
    }
    
    duration := time.Since(start)
    fmt.Printf("并发解析 %d 个文档，耗时: %v\n", len(users), duration)
}

func generateJSONDocuments(count int) [][]byte {
    documents := make([][]byte, count)
    for i := 0; i < count; i++ {
        documents[i] = []byte(fmt.Sprintf(
            `{"id": %d, "name": "用户%d", "email": "user%d@example.com", "age": %d}`,
            i+1, i+1, i+1, 20+i%50,
        ))
    }
    return documents
}
```

## 批量序列化

### 基础批量序列化

```go
func basicBatchSerialization() {
    // 创建大量数据
    users := make([]User, 5000)
    for i := 0; i < 5000; i++ {
        users[i] = User{
            ID:    int64(i + 1),
            Name:  fmt.Sprintf("用户%d", i+1),
            Email: fmt.Sprintf("user%d@example.com", i+1),
            Age:   20 + i%50,
        }
    }

    start := time.Now()
    
    // 标准序列化
    standardJSON, err := fxjson.Marshal(users)
    if err != nil {
        panic(err)
    }
    standardDuration := time.Since(start)

    // 快速序列化
    start = time.Now()
    fastJSON := fxjson.FastMarshal(users)
    fastDuration := time.Since(start)

    fmt.Printf("标准序列化: %v, 大小: %d bytes\n", standardDuration, len(standardJSON))
    fmt.Printf("快速序列化: %v, 大小: %d bytes\n", fastDuration, len(fastJSON))
    fmt.Printf("性能提升: %.2fx\n", float64(standardDuration)/float64(fastDuration))
}
```

### 分批序列化优化

```go
func batchMarshalOptimized() {
    users := generateUsers(10000)
    
    // 分批处理避免内存峰值
    batchSize := 1000
    var allResults [][]byte
    
    start := time.Now()
    
    for i := 0; i < len(users); i += batchSize {
        end := i + batchSize
        if end > len(users) {
            end = len(users)
        }
        
        batch := users[i:end]
        batchJSON := fxjson.FastMarshal(batch)
        allResults = append(allResults, batchJSON)
    }
    
    duration := time.Since(start)
    
    totalSize := 0
    for _, result := range allResults {
        totalSize += len(result)
    }
    
    fmt.Printf("分批序列化完成\n")
    fmt.Printf("数据量: %d\n", len(users))
    fmt.Printf("批次数: %d\n", len(allResults))
    fmt.Printf("耗时: %v\n", duration)
    fmt.Printf("总大小: %d bytes\n", totalSize)
}

func generateUsers(count int) []User {
    users := make([]User, count)
    for i := 0; i < count; i++ {
        users[i] = User{
            ID:    int64(i + 1),
            Name:  fmt.Sprintf("用户%d", i+1),
            Email: fmt.Sprintf("user%d@example.com", i+1),
            Age:   20 + i%50,
        }
    }
    return users
}
```

## 批量转换

### 数据格式转换

```go
func batchTransformation() {
    // 原始数据格式
    rawData := []byte(`[
        {"user_name": "zhang_san", "user_email": "ZHANG@EXAMPLE.COM", "user_age": "28"},
        {"user_name": "li_si", "user_email": "LI@EXAMPLE.COM", "user_age": "25"},
        {"user_name": "wang_wu", "user_email": "WANG@EXAMPLE.COM", "user_age": "32"}
    ]`)

    node := fxjson.FromBytes(rawData)
    
    // 批量转换
    var transformedUsers []map[string]interface{}
    
    node.ArrayForEach(func(index int, userNode fxjson.Node) bool {
        nameStr := userNode.Get("user_name").StringOr("")
        emailStr := userNode.Get("user_email").StringOr("")
        ageStr := userNode.Get("user_age").StringOr("0")
        age, _ := strconv.Atoi(ageStr)
        
        transformed := map[string]interface{}{
            "id":       index + 1,
            "name":     normalizeUsername(nameStr),
            "email":    strings.ToLower(emailStr),
            "age":      age,
            "isActive": true,
            "profile": map[string]interface{}{
                "createdAt": time.Now().Format(time.RFC3339),
                "updatedAt": time.Now().Format(time.RFC3339),
            },
        }
        
        transformedUsers = append(transformedUsers, transformed)
        return true
    })

    // 序列化转换结果
    result, _ := fxjson.MarshalIndent(transformedUsers, "", "  ")
    
    fmt.Printf("转换结果:\n%s\n", result)
}

func normalizeUsername(username string) string {
    // 将下划线格式转换为驼峰格式
    parts := strings.Split(username, "_")
    if len(parts) <= 1 {
        return username
    }
    
    normalized := parts[0]
    for i := 1; i < len(parts); i++ {
        if len(parts[i]) > 0 {
            normalized += strings.ToUpper(parts[i][:1]) + parts[i][1:]
        }
    }
    return normalized
}
```

## 批量数据处理

### 大数据集遍历

```go
func batchDataProcessing() {
    // 生成大量销售数据
    salesData := generateLargeSalesData(50000)
    node := fxjson.FromBytes(salesData)
    
    sales := node.Get("sales")
    
    start := time.Now()
    
    // 使用 ArrayForEach 进行批量处理
    var highValueSales []map[string]interface{}
    var totalAmount float64
    var count int
    
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        amount := sale.Get("amount").FloatOr(0)
        totalAmount += amount
        count++
        
        // 筛选高价值销售
        if amount > 5000 {
            saleData := map[string]interface{}{
                "id":       index,
                "product":  sale.Get("product").StringOr(""),
                "amount":   amount,
                "category": sale.Get("category").StringOr(""),
            }
            highValueSales = append(highValueSales, saleData)
        }
        
        return true
    })
    
    duration := time.Since(start)
    
    fmt.Printf("处理完成，耗时: %v\n", duration)
    fmt.Printf("总记录数: %d\n", count)
    fmt.Printf("总销售额: %.2f\n", totalAmount)
    fmt.Printf("平均销售额: %.2f\n", totalAmount/float64(count))
    fmt.Printf("高价值销售数: %d\n", len(highValueSales))
}

func generateLargeSalesData(count int) []byte {
    categories := []string{"电子产品", "服装", "图书", "食品", "家具"}
    regions := []string{"北京", "上海", "广州", "深圳", "杭州"}
    
    var sales []string
    for i := 0; i < count; i++ {
        sale := fmt.Sprintf(`{
            "id": %d,
            "product": "商品%d",
            "category": "%s",
            "amount": %d,
            "quantity": %d,
            "region": "%s",
            "date": "2024-%02d-%02d"
        }`, 
            i+1,
            i+1,
            categories[i%len(categories)],
            100+(i%9900),
            1+(i%50),
            regions[i%len(regions)],
            (i%12)+1,
            (i%28)+1,
        )
        sales = append(sales, sale)
    }
    
    return []byte(fmt.Sprintf(`{"sales": [%s]}`, strings.Join(sales, ",")))
}
```

### 深度遍历批量处理

```go
func deepTraversalBatch() {
    // 复杂嵌套数据
    complexData := []byte(`{
        "company": {
            "departments": [
                {
                    "name": "研发部",
                    "teams": [
                        {
                            "name": "后端组",
                            "members": [
                                {"name": "张三", "role": "工程师"},
                                {"name": "李四", "role": "高级工程师"}
                            ]
                        },
                        {
                            "name": "前端组",
                            "members": [
                                {"name": "王五", "role": "工程师"},
                                {"name": "赵六", "role": "架构师"}
                            ]
                        }
                    ]
                },
                {
                    "name": "产品部",
                    "teams": [
                        {
                            "name": "产品设计组",
                            "members": [
                                {"name": "陈七", "role": "产品经理"},
                                {"name": "周八", "role": "UI设计师"}
                            ]
                        }
                    ]
                }
            ]
        }
    }`)
    
    node := fxjson.FromBytes(complexData)
    
    // 使用 Walk 进行深度遍历
    var allMembers []string
    var roleCount = make(map[string]int)
    
    node.Walk(func(path string, n fxjson.Node) bool {
        // 查找所有的 members
        if strings.HasSuffix(path, ".name") && strings.Contains(path, "members") {
            if name := n.StringOr(""); name != "" {
                allMembers = append(allMembers, name)
            }
        }
        
        // 统计角色
        if strings.HasSuffix(path, ".role") {
            if role := n.StringOr(""); role != "" {
                roleCount[role]++
            }
        }
        
        return true
    })
    
    fmt.Printf("所有成员: %v\n", allMembers)
    fmt.Printf("角色统计:\n")
    for role, count := range roleCount {
        fmt.Printf("  %s: %d人\n", role, count)
    }
}
```

## 性能优化策略

### 内存管理

```go
func memoryOptimizedBatchProcessing() {
    // 对于超大数据集，使用分块处理避免内存溢出
    chunkSize := 1000
    processedCount := 0
    
    // 模拟流式处理
    for chunk := 0; chunk < 100; chunk++ { // 假设100个块
        chunkData := generateChunkData(chunkSize, chunk)
        node := fxjson.FromBytes(chunkData)
        
        // 处理当前块
        processChunk(node, chunk)
        processedCount += chunkSize
        
        // 定期触发GC，释放内存
        if chunk%10 == 0 {
            runtime.GC()
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            fmt.Printf("处理进度: %d/%d, 内存使用: %d KB\n", 
                processedCount, 100*chunkSize, m.Alloc/1024)
        }
    }
    
    fmt.Printf("流式处理完成，总共处理 %d 条记录\n", processedCount)
}

func generateChunkData(size, chunkIndex int) []byte {
    var items []string
    for i := 0; i < size; i++ {
        id := chunkIndex*size + i + 1
        item := fmt.Sprintf(`{"id": %d, "name": "项目%d", "value": %d}`, 
            id, id, id*100)
        items = append(items, item)
    }
    return []byte(fmt.Sprintf(`{"items": [%s]}`, strings.Join(items, ",")))
}

func processChunk(node fxjson.Node, chunkIndex int) {
    count := 0
    totalValue := 0.0
    
    node.Get("items").ArrayForEach(func(index int, item fxjson.Node) bool {
        count++
        totalValue += item.Get("value").FloatOr(0)
        return true
    })
    
    fmt.Printf("块 %d: 处理 %d 项，总值 %.0f\n", chunkIndex, count, totalValue)
}
```

### 并发处理优化

```go
func concurrentBatchOptimization() {
    // 生成大量数据进行并发处理测试
    data := generateLargeSalesData(20000)
    node := fxjson.FromBytes(data)
    sales := node.Get("sales")
    
    // 串行处理
    start := time.Now()
    serialResult := processSerially(sales)
    serialDuration := time.Since(start)
    
    // 并发处理
    start = time.Now()
    concurrentResult := processConcurrently(sales)
    concurrentDuration := time.Since(start)
    
    fmt.Printf("串行处理: %v, 结果: %d\n", serialDuration, serialResult)
    fmt.Printf("并发处理: %v, 结果: %d\n", concurrentDuration, concurrentResult)
    fmt.Printf("性能提升: %.2fx\n", float64(serialDuration)/float64(concurrentDuration))
}

func processSerially(sales fxjson.Node) int {
    count := 0
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        amount := sale.Get("amount").FloatOr(0)
        if amount > 5000 {
            count++
        }
        return true
    })
    return count
}

func processConcurrently(sales fxjson.Node) int {
    numWorkers := runtime.NumCPU()
    totalCount := sales.Len()
    chunkSize := totalCount / numWorkers
    
    if chunkSize == 0 {
        chunkSize = 1
    }
    
    var wg sync.WaitGroup
    resultChan := make(chan int, numWorkers)
    
    for i := 0; i < numWorkers; i++ {
        start := i * chunkSize
        end := start + chunkSize
        if i == numWorkers-1 {
            end = totalCount // 最后一个worker处理剩余的所有数据
        }
        
        wg.Add(1)
        go func(startIdx, endIdx int) {
            defer wg.Done()
            
            localCount := 0
            for j := startIdx; j < endIdx; j++ {
                sale := sales.Index(j)
                amount := sale.Get("amount").FloatOr(0)
                if amount > 5000 {
                    localCount++
                }
            }
            resultChan <- localCount
        }(start, end)
    }
    
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    totalResult := 0
    for result := range resultChan {
        totalResult += result
    }
    
    return totalResult
}
```

### 性能监控工具

```go
func performanceMonitoring() {
    // 性能监控示例
    start := time.Now()
    
    // 生成测试数据
    data := generateLargeSalesData(10000)
    parseTime := time.Since(start)
    
    // 解析
    start = time.Now()
    node := fxjson.FromBytes(data)
    decodeTime := time.Since(start)
    
    // 处理
    start = time.Now()
    count := 0
    totalAmount := 0.0
    
    node.Get("sales").ArrayForEach(func(index int, sale fxjson.Node) bool {
        count++
        totalAmount += sale.Get("amount").FloatOr(0)
        return true
    })
    processTime := time.Since(start)
    
    // 输出性能报告
    fmt.Println("性能监控报告:")
    fmt.Printf("数据生成: %v\n", parseTime)
    fmt.Printf("JSON解析: %v\n", decodeTime)
    fmt.Printf("数据处理: %v\n", processTime)
    fmt.Printf("总耗时: %v\n", parseTime+decodeTime+processTime)
    fmt.Printf("处理速度: %.0f 记录/秒\n", float64(count)/processTime.Seconds())
    
    // 内存使用情况
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("内存使用: %d KB\n", m.Alloc/1024)
    fmt.Printf("GC次数: %d\n", m.NumGC)
}
```

批量操作功能让 FxJSON 能够高效处理大规模数据，通过合理的并发控制、内存管理和性能监控，您可以构建高性能的数据处理应用程序。