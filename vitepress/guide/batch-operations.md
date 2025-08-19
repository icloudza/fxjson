# 批量操作

FxJSON 提供了高效的批量操作功能，让您能够处理大量 JSON 数据。本节介绍各种批量处理技术和优化方法。

## 批量解析

### 基础批量解析

```go
func basicBatchParsing() {
    // 模拟多个JSON文档
    jsonDocuments := [][]byte{
        []byte(`{"id": 1, "name": "用户1", "email": "user1@example.com"}`),
        []byte(`{"id": 2, "name": "用户2", "email": "user2@example.com"}`),
        []byte(`{"id": 3, "name": "用户3", "email": "user3@example.com"}`),
        // ... 更多文档
    }

    // 批量解析
    nodes := make([]fxjson.Node, len(jsonDocuments))
    for i, jsonData := range jsonDocuments {
        nodes[i] = fxjson.FromBytes(jsonData)
    }

    // 批量处理
    var users []User
    for _, node := range nodes {
        var user User
        if err := node.Decode(&user); err == nil {
            users = append(users, user)
        }
    }

    fmt.Printf("成功解析 %d 个用户\n", len(users))
}

type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age,omitempty"`
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