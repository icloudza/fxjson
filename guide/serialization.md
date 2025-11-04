# 序列化功能指南

FxJSON 提供了强大而灵活的序列化功能，支持将 Go 数据结构转换为 JSON 字符串。不仅支持基础的序列化，还包括批量序列化、流式序列化、性能优化等高级特性。

## 目录

- [基础序列化](#基础序列化)
- [快速序列化](#快速序列化)
- [类型专用序列化](#类型专用序列化)
- [批量序列化](#批量序列化)
- [流式序列化](#流式序列化)
- [序列化选项](#序列化选项)
- [���能优化](#性能优化)
- [实际应用场景](#实际应用场景)

## 基础序列化

### Marshal()

最基础的序列化方法，支持所有 Go 数据类型。

```go
func Marshal(v interface{}) ([]byte, error)
```

**基本用法**：
```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

type User struct {
    ID       int      `json:"id"`
    Name     string   `json:"name"`
    Email    string   `json:"email"`
    Tags     []string `json:"tags,omitempty"`
    Active   bool     `json:"active"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func main() {
    user := User{
        ID:     1001,
        Name:   "张三",
        Email:  "zhangsan@example.com",
        Tags:   []string{"golang", "developer"},
        Active: true,
        Metadata: map[string]interface{}{
            "level": "senior",
            "score": 95.5,
        },
    }

    // 序列化为 JSON
    data, err := fxjson.Marshal(user)
    if err != nil {
        panic(err)
    }

    fmt.Printf("JSON: %s\n", data)
    // 输出: {"id":1001,"name":"张三","email":"zhangsan@example.com","tags":["golang","developer"],"active":true,"metadata":{"level":"senior","score":95.5}}
}
```

## 快速序列化

### MarshalStructFast()

高性能的结构体序列化，比标准方法快 3-5 倍。

```go
func MarshalStructFast(v interface{}) []byte
```

**特点**：
- 返回字节数组而非错误（内部处理错误）
- 性能优化，减少内存分配
- 自动处理空值和默认值

```go
func main() {
    user := User{
        ID:     1001,
        Name:   "张三",
        Email:  "zhangsan@example.com",
        Active: true,
    }

    // 快速序列化（无需错误处理）
    data := fxjson.MarshalStructFast(user)
    fmt.Printf("Fast JSON: %s\n", data)

    // 如果需要错误信息，可以使用 MarshalStruct
    data, err := fxjson.MarshalStruct(user)
    if err != nil {
        log.Printf("序列化错误: %v", err)
    }
}
```

## 类型专用序列化

### MarshalSlice()

专门用于切片/数组序列化的优化方法。

```go
func MarshalSlice(v interface{}) ([]byte, error)
```

```go
func main() {
    // 字符串切片
    names := []string{"张三", "李四", "王五"}
    data, _ := fxjson.MarshalSlice(names)
    fmt.Printf("Names: %s\n", data)

    // 结构体切片
    users := []User{
        {ID: 1, Name: "张三", Active: true},
        {ID: 2, Name: "李四", Active: false},
    }
    data, _ = fxjson.MarshalSlice(users)
    fmt.Printf("Users: %s\n", data)

    // 数字切片
    scores := []float64{95.5, 87.0, 92.5}
    data, _ = fxjson.MarshalSlice(scores)
    fmt.Printf("Scores: %s\n", data)
}
```

### MarshalMap()

专门用于 Map 序列化的优化方法。

```go
func MarshalMap(v interface{}) ([]byte, error)
```

```go
func main() {
    // 简单 Map
    config := map[string]string{
        "host": "localhost",
        "port": "8080",
        "env":  "production",
    }
    data, _ := fxjson.MarshalMap(config)
    fmt.Printf("Config: %s\n", data)

    // 复杂 Map
    dataMap := map[string]interface{}{
        "count":    100,
        "active":   true,
        "users":    []string{"user1", "user2"},
        "settings": map[string]interface{}{
            "cache":  true,
            "debug":  false,
            "limit":  1000,
        },
    }
    data, _ = fxjson.MarshalMap(dataMap)
    fmt.Printf("Data: %s\n", data)
}
```

### 序列化选项

FxJSON 提供了丰富的序列化选项：

```go
func advancedSerialization() {
    user := User{
        ID:    1,
        Name:  "张三",
        Email: "zhangsan@example.com",
        Age:   30,
    }

    // 使用默认选项（压缩模式）
    compact, _ := fxjson.MarshalWithOptions(user, fxjson.DefaultSerializeOptions)
    fmt.Printf("紧凑模式: %s\n", compact)

    // 使用美化选项
    pretty, _ := fxjson.MarshalWithOptions(user, fxjson.PrettySerializeOptions)
    fmt.Printf("美化模式:\n%s\n", pretty)

    // 自定义选项
    customOpts := fxjson.SerializeOptions{
        Indent:          "    ", // 4个空格缩进
        EscapeHTML:      true,   // 转义HTML字符
        SortKeys:        true,   // 键排序
        OmitEmpty:       true,   // 忽略空值
        FloatPrecision:  2,      // 浮点数保留2位小数
        UseNumberString: false,  // 数字不用字符串
    }
    
    custom, _ := fxjson.MarshalWithOptions(user, customOpts)
    fmt.Printf("自定义格式:\n%s\n", custom)
}
```

### 嵌套结构体序列化

```go
type Address struct {
    Street   string `json:"street"`
    City     string `json:"city"`
    Country  string `json:"country"`
    PostCode string `json:"post_code,omitempty"`
}

type Company struct {
    Name    string  `json:"name"`
    Website string  `json:"website,omitempty"`
    Address Address `json:"address"`
}

type Employee struct {
    ID          int64    `json:"id"`
    Name        string   `json:"name"`
    Email       string   `json:"email"`
    Department  string   `json:"department"`
    Salary      float64  `json:"salary"`
    Skills      []string `json:"skills"`
    Company     Company  `json:"company"`
    IsActive    bool     `json:"is_active"`
    JoinedAt    string   `json:"joined_at"`
}

func nestedSerialization() {
    employee := Employee{
        ID:         1001,
        Name:       "张三",
        Email:      "zhangsan@company.com",
        Department: "工程部",
        Salary:     150000.50,
        Skills:     []string{"Go", "Python", "Docker", "Kubernetes"},
        Company: Company{
            Name:    "科技公司",
            Website: "https://tech-company.com",
            Address: Address{
                Street:   "中关村大街1号",
                City:     "北京",
                Country:  "中国",
                PostCode: "100080",
            },
        },
        IsActive: true,
        JoinedAt: "2023-01-15T09:00:00Z",
    }

    // 序列化嵌套结构
    jsonData, err := fxjson.FastMarshal(employee)
    if err != nil {
        panic(err)
    }

    // 美化输出
    prettyData, _ := fxjson.MarshalWithOptions(employee, fxjson.PrettySerializeOptions)
    fmt.Printf("员工信息 JSON:\n%s\n", prettyData)
}
```

## 序列化到字符串

### MarshalToString()

直接序列化为字符串，无需类型转换。

```go
func MarshalToString(v interface{}) (string, error)
func MarshalToStringWithOptions(v interface{}, opts SerializeOptions) (string, error)
```

```go
func stringSerializationExample() {
    user := User{
        ID:    1,
        Name:  "张三",
        Email: "zhang@example.com",
        Age:   30,
    }

    // 直接序列化为字符串
    jsonStr, err := fxjson.MarshalToString(user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("JSON 字符串: %s\n", jsonStr)

    // 带选项序列化为字符串
    opts := fxjson.SerializeOptions{
        Indent:     "  ",
        SortKeys:   true,
        EscapeHTML: false,
    }
    prettyStr, _ := fxjson.MarshalToStringWithOptions(user, opts)
    fmt.Printf("美化字符串:\n%s\n", prettyStr)
}
```

## 流式序列化到 Writer

### MarshalToWriter()

直接序列化到 Writer，避免中间缓冲区，节省内存。

```go
func MarshalToWriter(v interface{}, writer func([]byte) error) error
func MarshalToWriterWithOptions(v interface{}, writer func([]byte) error, opts SerializeOptions) error
```

```go
func writerSerializationExample() {
    users := []User{
        {ID: 1, Name: "张三", Email: "zhang@example.com", Age: 30},
        {ID: 2, Name: "李四", Email: "li@example.com", Age: 25},
        {ID: 3, Name: "王五", Email: "wang@example.com", Age: 35},
    }

    // 写入文件
    file, err := os.Create("users.json")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // 流式写入
    err = fxjson.MarshalToWriter(users, func(data []byte) error {
        _, err := file.Write(data)
        return err
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("数据已写入 users.json")
}

func httpResponseExample() {
    http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        users := []User{
            {ID: 1, Name: "张三", Email: "zhang@example.com", Age: 30},
            {ID: 2, Name: "李四", Email: "li@example.com", Age: 25},
        }

        w.Header().Set("Content-Type", "application/json")

        // 直接流式写入 HTTP 响应
        err := fxjson.MarshalToWriter(users, func(data []byte) error {
            _, err := w.Write(data)
            return err
        })

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })
}
```

## 特殊类型序列化

### 时间序列化

```go
func MarshalTime(t time.Time) []byte
func MarshalTimeUnix(t time.Time) []byte
```

```go
func timeSerializationExample() {
    now := time.Now()

    // 标准 RFC3339 格式
    timeJSON := fxjson.MarshalTime(now)
    fmt.Printf("时间(RFC3339): %s\n", timeJSON)
    // 输出: "2024-01-15T10:30:00Z"

    // Unix 时间戳
    unixJSON := fxjson.MarshalTimeUnix(now)
    fmt.Printf("时间(Unix): %s\n", unixJSON)
    // 输出: 1705315800

    // 在结构体中使用
    type Event struct {
        ID        int       `json:"id"`
        Name      string    `json:"name"`
        CreatedAt time.Time `json:"created_at"`
    }

    event := Event{
        ID:        1,
        Name:      "技术分享会",
        CreatedAt: now,
    }

    eventJSON := fxjson.FastMarshal(event)
    fmt.Printf("事件: %s\n", eventJSON)
}
```

### 时间段序列化

```go
func MarshalDuration(d time.Duration) []byte
```

```go
func durationSerializationExample() {
    duration := 2*time.Hour + 30*time.Minute

    // 序列化为纳秒数
    durationJSON := fxjson.MarshalDuration(duration)
    fmt.Printf("时长: %s\n", durationJSON)
    // 输出: 9000000000000

    // 在结构体中使用
    type Task struct {
        Name     string        `json:"name"`
        Duration time.Duration `json:"duration_ns"`
    }

    task := Task{
        Name:     "数据处理",
        Duration: duration,
    }

    taskJSON := fxjson.FastMarshal(task)
    fmt.Printf("任务: %s\n", taskJSON)
}
```

### 二进制数据序列化

```go
func MarshalBinary(data []byte) []byte
```

```go
func binarySerializationExample() {
    // 原始二进制数据
    binaryData := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}  // "Hello"

    // 序列化为 Base64 编码的 JSON 字符串
    binaryJSON := fxjson.MarshalBinary(binaryData)
    fmt.Printf("二进制数据: %s\n", binaryJSON)
    // 输出: "SGVsbG8="

    // 在结构体中使用
    type FileData struct {
        Name    string `json:"name"`
        Content []byte `json:"content"`
    }

    fileData := FileData{
        Name:    "document.pdf",
        Content: binaryData,
    }

    fileJSON := fxjson.FastMarshal(fileData)
    fmt.Printf("文件数据: %s\n", fileJSON)
}
```

## 批量序列化

### 数组序列化

```go
func arraySerialization() {
    users := []User{
        {ID: 1, Name: "张三", Email: "zhang@example.com", Age: 30},
        {ID: 2, Name: "李四", Email: "li@example.com", Age: 25},
        {ID: 3, Name: "王五", Email: "wang@example.com", Age: 35},
    }

    // 标准数组序列化
    jsonData := fxjson.FastMarshal(users)
    fmt.Printf("JSON 数据: %s\n", jsonData)

    // 美化输出
    prettyData, _ := fxjson.MarshalIndent(users, "", "  ")
    fmt.Printf("用户列表:\n%s\n", prettyData)

    // 序列化为字符串
    jsonStr, _ := fxjson.MarshalToString(users)
    fmt.Printf("字符串格式: %s\n", jsonStr)
}
```

### 批量序列化器

对于大量数据，使用批量序列化器获得更好的性能：

```go
func batchSerialization() {
    // 创建大量用户数据
    users := make([]User, 10000)
    for i := 0; i < 10000; i++ {
        users[i] = User{
            ID:    int64(i + 1),
            Name:  fmt.Sprintf("用户%d", i+1),
            Email: fmt.Sprintf("user%d@example.com", i+1),
            Age:   20 + i%50,
        }
    }

    // 使用高性能序列化
    start := time.Now()
    jsonData := fxjson.FastMarshal(users)
    duration := time.Since(start)

    fmt.Printf("批量序列化 %d 个用户耗时: %v\n", len(users), duration)
    fmt.Printf("JSON 大小: %d bytes\n", len(jsonData))
}
```

## 从 JSON 反序列化

### 基础反序列化

```go
func basicDeserialization() {
    jsonData := `{
        "id": 1,
        "name": "张三",
        "email": "zhangsan@example.com",
        "age": 30
    }`

    node := fxjson.FromBytes([]byte(jsonData))

    // 反序列化到结构体
    var user User
    err := node.DecodeStruct(&user)
    if err != nil {
        panic(err)
    }

    fmt.Printf("用户: %+v\n", user)

    // 也可以使用 Decode 方法
    var fastUser User
    err = node.Decode(&fastUser)
    if err != nil {
        panic(err)
    }

    fmt.Printf("解码用户: %+v\n", fastUser)
}
```

### 数组反序列化

```go
func arrayDeserialization() {
    jsonData := `[
        {"id": 1, "name": "张三", "email": "zhang@example.com", "age": 30},
        {"id": 2, "name": "李四", "email": "li@example.com", "age": 25},
        {"id": 3, "name": "王五", "email": "wang@example.com", "age": 35}
    ]`

    node := fxjson.FromBytes([]byte(jsonData))

    // 反序列化到数组
    var users []User
    err := node.DecodeStruct(&users)
    if err != nil {
        panic(err)
    }

    fmt.Printf("用户列表: %+v\n", users)

    // 逐个解码（适用于大数组）
    var streamUsers []User
    node.ArrayForEach(func(index int, userNode fxjson.Node) bool {
        var user User
        if err := userNode.Decode(&user); err == nil {
            streamUsers = append(streamUsers, user)
        }
        return true
    })

    fmt.Printf("流式解码用户数: %d\n", len(streamUsers))
}
```

## 自定义序列化

### 标签支持

FxJSON 支持标准的 JSON 标签：

```go
type Product struct {
    ID          int64   `json:"id"`
    Name        string  `json:"name"`
    Description string  `json:"description,omitempty"`
    Price       float64 `json:"price"`
    Stock       int     `json:"stock"`
    Category    string  `json:"category"`
    Tags        []string `json:"tags,omitempty"`
    CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
    
    // 忽略字段
    InternalID string `json:"-"`
}

func customTagSerialization() {
    product := Product{
        ID:          1001,
        Name:        "智能手机",
        Description: "", // 空值，会被 omitempty 忽略
        Price:       2999.99,
        Stock:       100,
        Category:    "电子产品",
        Tags:        []string{"智能", "通讯", "科技"},
        CreatedAt:   "2024-01-01T00:00:00Z",
        UpdatedAt:   "2024-01-15T12:00:00Z",
        InternalID:  "INTERNAL-123", // 会被忽略
    }

    jsonData, _ := fxjson.MarshalIndent(product, "", "  ")
    fmt.Printf("商品信息:\n%s\n", jsonData)
}
```

### 自定义字段映射

```go
type APIResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

func customMarshal() {
    // 使用接口类型处理动态数据
    response := APIResponse{
        Code:    200,
        Message: "成功",
        Data: map[string]interface{}{
            "users": []User{
                {ID: 1, Name: "张三", Email: "zhang@example.com", Age: 30},
                {ID: 2, Name: "李四", Email: "li@example.com", Age: 25},
            },
            "total": 2,
            "page":  1,
        },
    }

    jsonData, _ := fxjson.MarshalIndent(response, "", "  ")
    fmt.Printf("API 响应:\n%s\n", jsonData)
}
```

## 时间处理

```go
import "time"

type Event struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
    Duration    int64     `json:"duration_seconds"`
}

func timeSerialization() {
    now := time.Now()
    event := Event{
        ID:          1,
        Name:        "技术分享会",
        Description: "Go 语言高性能编程",
        StartTime:   now,
        EndTime:     now.Add(2 * time.Hour),
        Duration:    7200, // 2小时 = 7200秒
    }

    jsonData, _ := fxjson.MarshalIndent(event, "", "  ")
    fmt.Printf("活动信息:\n%s\n", jsonData)

    // 从 JSON 解析时间
    node := fxjson.FromBytes(jsonData)
    var parsedEvent Event
    node.Decode(&parsedEvent)
    
    fmt.Printf("解析后的开始时间: %v\n", parsedEvent.StartTime)
}
```

## 性能优化技巧

### 1. 选择合适的序列化方法

```go
func performanceComparison() {
    user := User{
        ID:    1,
        Name:  "张三",
        Email: "zhangsan@example.com",
        Age:   30,
    }

    // 标准序列化
    start := time.Now()
    for i := 0; i < 10000; i++ {
        _, _ = fxjson.Marshal(user)
    }
    standardDuration := time.Since(start)

    // 高性能序列化
    start = time.Now()
    for i := 0; i < 10000; i++ {
        _, _ = fxjson.FastMarshal(user)
    }
    fastDuration := time.Since(start)

    fmt.Printf("标准序列化 10k 次: %v\n", standardDuration)
    fmt.Printf("快速序列化 10k 次: %v\n", fastDuration)
    fmt.Printf("性能提升: %.2fx\n", 
        float64(standardDuration)/float64(fastDuration))
}
```

### 2. 重用序列化器

```go
func reuseSerializer() {
    users := make([]User, 1000)
    for i := 0; i < 1000; i++ {
        users[i] = User{
            ID:    int64(i + 1),
            Name:  fmt.Sprintf("用户%d", i+1),
            Email: fmt.Sprintf("user%d@example.com", i+1),
            Age:   20 + i%50,
        }
    }

    // 多次序列化
    for round := 0; round < 10; round++ {
        _ = fxjson.FastMarshal(users)
    }

    fmt.Println("重复序列化完成")
}
```

### 3. 内存池使用

```go
func memoryPoolUsage() {
    // FxJSON 内部使用内存池，用户无需管理
    // 但可以通过配置减少内存分配
    
    opts := fxjson.SerializeOptions{
        Indent:          "", // 紧凑模式减少内存使用
        SortKeys:        false, // 不排序减少计算
        OmitEmpty:       false, // 不检查空值减少判断
        FloatPrecision:  -1, // 使用默认精度
        UseNumberString: false, // 不转换为字符串
    }

    user := User{
        ID:    1,
        Name:  "张三",
        Email: "zhangsan@example.com",
        Age:   30,
    }

    // 使用优化选项
    jsonData, _ := fxjson.MarshalWithOptions(user, opts)
    fmt.Printf("优化序列化: %s\n", jsonData)
}
```

## 错误处理

```go
func errorHandling() {
    // 处理循环引用
    type Node struct {
        Value string `json:"value"`
        Child *Node  `json:"child,omitempty"`
    }

    root := &Node{Value: "root"}
    child := &Node{Value: "child"}
    root.Child = child
    // child.Parent = root // 这会导致循环引用

    jsonData, err := fxjson.FastMarshal(root)
    if err != nil {
        fmt.Printf("序列化错误: %v\n", err)
        return
    }

    fmt.Printf("成功序列化: %s\n", jsonData)

    // 处理类型错误
    invalidData := `{"id": "not_a_number", "name": "张三"}`
    node := fxjson.FromBytes([]byte(invalidData))

    var user User
    err = node.Decode(&user)
    if err != nil {
        fmt.Printf("反序列化错误: %v\n", err)
        
        // 使用安全方式获取数据
        user.ID = node.Get("id").IntOr(0) // 转换失败返回0
        user.Name = node.Get("name").StringOr("未知")
        
        fmt.Printf("安全解析结果: %+v\n", user)
    }
}
```

## 最佳实践

### 1. 结构体设计

```go
// 好的设计
type GoodUser struct {
    ID    int64  `json:"id"`              // 明确的类型
    Name  string `json:"name"`            // 必需字段
    Email string `json:"email,omitempty"` // 可选字段使用 omitempty
    Age   int    `json:"age"`             // 简单类型
}

// 需要注意的设计
type ProblematicUser struct {
    ID   interface{} `json:"id"`   // 避免使用 interface{}
    Data map[string]interface{} `json:"data"` // 复杂嵌套结构
}
```

### 2. 性能优化

```go
func bestPractices() {
    // 1. 对于大量数据使用 FastMarshal
    largeData := make([]User, 10000)
    jsonData, _ := fxjson.FastMarshal(largeData)
    
    // 2. 选择合适的解析方式
    node := fxjson.FromBytes(jsonData)
    
    // 小数据量：直接解码
    var smallUsers []User
    node.Decode(&smallUsers)
    
    // 大数据量：流式处理
    count := 0
    node.ArrayForEach(func(index int, userNode fxjson.Node) bool {
        // 只处理前100个
        if count >= 100 {
            return false
        }
        count++
        return true
    })
}
```

序列化功能是 FxJSON 的重要组成部分，通过合理使用这些功能，您可以在保证性能的同时，灵活处理各种数据格式转换需求。