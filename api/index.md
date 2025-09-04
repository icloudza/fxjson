# API 参考手册

FxJSON 提供了简洁而强大的 API，让您能够高效处理 JSON 数据。本文档将按照使用频率和学习难度组织，帮助您快速找到需要的方法。

## 快速索引

- [核心解析函数](#核心解析函数) - 如何开始使用 FxJSON
- [基础数据访问](#基础数据访问) - 获取 JSON 中的值
- [类型转换方法](#类型转换方法) - 安全地转换数据类型
- [数组操作](#数组操作) - 处理 JSON 数组
- [对象操作](#对象操作) - 处理 JSON 对象  
- [高级功能](#高级功能) - 数据验证、序列化等高级特性

---

## 核心解析函数

这些函数是使用 FxJSON 的起点，用于将 JSON 字符串或字节数组转换为可操作的 Node 对象。

### FromString()
```go
func FromString(s string) Node
```

**最常用的解析函数**，从 JSON 字符串创建 Node。

**使用场景**：处理配置文件、API 响应等字符串格式的 JSON

```go
// 基本用法
jsonStr := `{"name": "张三", "age": 30}`
node := fxjson.FromString(jsonStr)
name := node.Get("name").StringOr("")  // "张三"
```

### FromBytes()
```go
func FromBytes(b []byte) Node
```

从字节数组解析 JSON，性能略好于字符串版本。

**使用场景**：处理网络请求、文件读取等字节数据

```go
// 处理 HTTP 响应
response, _ := http.Get("https://api.example.com/user")
body, _ := io.ReadAll(response.Body)
node := fxjson.FromBytes(body)
```

### FromStringWithOptions() / FromBytesWithOptions()
```go
func FromStringWithOptions(s string, opts ParseOptions) Node
func FromBytesWithOptions(b []byte, opts ParseOptions) Node
```

带自定义解析选项的版本，用于处理特殊需求。

**使用场景**：需要限制解析深度、字符串长度等安全场景

```go
// 自定义解析选项
opts := fxjson.ParseOptions{
    MaxDepth: 100,        // 最大嵌套深度
    MaxStringLen: 10000,  // 最大字符串长度
    StrictMode: true,     // 严格模式
}
node := fxjson.FromStringWithOptions(jsonStr, opts)
```

---

## 基础数据访问

### Get()
```go
func (n Node) Get(key string) Node
```

**核心访问方法**，根据键名获取对象中的字段。

```go
json := `{"user": {"name": "张三", "profile": {"city": "北京"}}}`
node := fxjson.FromString(json)

// 链式访问
name := node.Get("user").Get("name").StringOr("")     // "张三"
city := node.Get("user").Get("profile").Get("city").StringOr("")  // "北京"
```

### GetPath()
```go
func (n Node) GetPath(path string) Node
```

**路径访问方法**，使用点号分隔的路径字符串访问嵌套数据，比链式调用更简洁。

```go
// 等价于上面的链式调用，但更简洁
name := node.GetPath("user.name").StringOr("")         // "张三"
city := node.GetPath("user.profile.city").StringOr("") // "北京"
```

### Index()
```go
func (n Node) Index(i int) Node
```

根据索引访问数组元素。

```go
json := `{"users": ["张三", "李四", "王五"]}`
node := fxjson.FromString(json)

users := node.Get("users")
first := users.Index(0).StringOr("")   // "张三"
second := users.Index(1).StringOr("")  // "李四"
```

---

## 类型转换方法

FxJSON 提供两套类型转换方法：**安全方法**（推荐）和**严格方法**。

### 安全转换方法（推荐）

这些方法自动处理错误，提供默认值，是日常使用的首选。

#### StringOr()
```go
func (n Node) StringOr(defaultValue string) string
```

获取字符串值，失败时返回默认值。

```go
name := node.Get("name").StringOr("未知用户")
email := node.Get("email").StringOr("未提供邮箱")
```

#### IntOr()
```go
func (n Node) IntOr(defaultValue int64) int64
```

获取整数值，失败时返回默认值。

```go
age := node.Get("age").IntOr(0)
score := node.Get("score").IntOr(-1)
```

#### FloatOr()
```go
func (n Node) FloatOr(defaultValue float64) float64
```

获取浮点数值，失败时返回默认值。

```go
price := node.Get("price").FloatOr(0.0)
rating := node.Get("rating").FloatOr(5.0)
```

#### BoolOr()
```go
func (n Node) BoolOr(defaultValue bool) bool
```

获取布尔值，失败时返回默认值。

```go
active := node.Get("active").BoolOr(false)
verified := node.Get("verified").BoolOr(false)
```

### 严格转换方法

这些方法返回错误，适合需要明确错误处理的场景。

#### String() / Int() / Float() / Bool()
```go
func (n Node) String() (string, error)
func (n Node) Int() (int64, error) 
func (n Node) Float() (float64, error)
func (n Node) Bool() (bool, error)
```

```go
// 严格模式示例
name, err := node.Get("name").String()
if err != nil {
    log.Printf("获取姓名失败: %v", err)
    return
}

age, err := node.Get("age").Int()
if err != nil {
    log.Printf("获取年龄失败: %v", err) 
    return
}
```

---

## 数组操作

### Len()
```go
func (n Node) Len() int
```

获取数组或对象的长度。

```go
json := `{"users": ["张三", "李四"], "count": 2}`
node := fxjson.FromString(json)

arrayLen := node.Get("users").Len()   // 2
objectLen := node.Len()               // 2（对象有2个字段）
```

### ArrayForEach()
```go
func (n Node) ArrayForEach(fn func(index int, item Node) bool)
```

**高性能数组遍历**，比标准库快 67 倍，零内存分配。

```go
json := `{"scores": [95, 87, 92, 88]}`
node := fxjson.FromString(json)

total := int64(0)
count := 0

node.Get("scores").ArrayForEach(func(index int, item fxjson.Node) bool {
    score := item.IntOr(0)
    total += score
    count++
    fmt.Printf("第%d个分数: %d\n", index+1, score)
    return true  // 返回 true 继续遍历，false 停止
})

average := float64(total) / float64(count)
fmt.Printf("平均分: %.1f\n", average)
```

### ToSlice系列方法

将 JSON 数组转换为 Go 切片。

#### ToStringSlice()
```go
func (n Node) ToStringSlice() ([]string, error)
```

#### ToIntSlice() 
```go
func (n Node) ToIntSlice() ([]int64, error)
```

#### ToFloatSlice()
```go
func (n Node) ToFloatSlice() ([]float64, error)
```

```go
json := `{
    "names": ["张三", "李四", "王五"],
    "ages": [25, 30, 35],
    "scores": [95.5, 87.0, 92.5]
}`
node := fxjson.FromString(json)

// 转换为字符串切片
names, err := node.Get("names").ToStringSlice()
if err == nil {
    fmt.Printf("姓名列表: %v\n", names)
}

// 转换为整数切片  
ages, err := node.Get("ages").ToIntSlice()
if err == nil {
    fmt.Printf("年龄列表: %v\n", ages)
}

// 转换为浮点数切片
scores, err := node.Get("scores").ToFloatSlice()
if err == nil {
    fmt.Printf("分数列表: %v\n", scores)
}
```

---

## 对象操作

### ForEach()
```go
func (n Node) ForEach(fn func(key string, value Node) bool)
```

**高性能对象遍历**，比标准库快 20 倍，零内存分配。

```go
json := `{
    "users": {
        "admin": {"name": "管理员", "level": 5},
        "guest": {"name": "访客", "level": 1}
    }
}`
node := fxjson.FromString(json)

node.Get("users").ForEach(func(userType string, userInfo fxjson.Node) bool {
    name := userInfo.Get("name").StringOr("")
    level := userInfo.Get("level").IntOr(0)
    fmt.Printf("%s: %s (等级 %d)\n", userType, name, level)
    return true
})
```

### Keys()
```go
func (n Node) Keys() []string
```

获取对象的所有键名。

```go
json := `{"name": "张三", "age": 30, "city": "北京"}`
node := fxjson.FromString(json)

keys := node.Keys()
fmt.Printf("对象的键: %v\n", keys)  // ["name", "age", "city"]
```

---

## 类型检查方法

### 基础类型检查
```go
func (n Node) IsString() bool    // 是否为字符串
func (n Node) IsNumber() bool    // 是否为数字
func (n Node) IsBool() bool      // 是否为布尔值
func (n Node) IsNull() bool      // 是否为 null
func (n Node) IsArray() bool     // 是否为数组
func (n Node) IsObject() bool    // 是否为对象
```

```go
json := `{
    "name": "张三",
    "age": 30,
    "active": true,
    "address": null,
    "hobbies": ["阅读"],
    "profile": {"city": "北京"}
}`
node := fxjson.FromString(json)

fmt.Printf("name 是字符串: %t\n", node.Get("name").IsString())     // true
fmt.Printf("age 是数字: %t\n", node.Get("age").IsNumber())         // true  
fmt.Printf("active 是布尔值: %t\n", node.Get("active").IsBool())   // true
fmt.Printf("address 是 null: %t\n", node.Get("address").IsNull()) // true
fmt.Printf("hobbies 是数组: %t\n", node.Get("hobbies").IsArray()) // true
fmt.Printf("profile 是对象: %t\n", node.Get("profile").IsObject())// true
```

### 分组类型检查
```go
func (n Node) IsScalar() bool     // 是否为标量（字符串、数字、布尔值、null）
func (n Node) IsContainer() bool  // 是否为容器（数组或对象）
```

### Exists()
```go
func (n Node) Exists() bool
```

检查字段是否存在（非常实用的方法）。

```go
if node.Get("optional_field").Exists() {
    // 字段存在时才处理
    value := node.Get("optional_field").StringOr("")
    fmt.Printf("可选字段的值: %s\n", value)
}
```

---

## 高级功能

### 数据验证

FxJSON 内置多种验证方法，方便进行数据校验。

#### 格式验证
```go
func (n Node) IsValidEmail() bool       // 验证邮箱格式
func (n Node) IsValidURL() bool         // 验证 URL 格式  
func (n Node) IsValidIP() bool          // 验证 IP 地址格式
func (n Node) IsValidJSON() bool        // 验证 JSON 格式
```

```go
json := `{
    "email": "user@example.com",
    "website": "https://example.com", 
    "server": "192.168.1.1"
}`
node := fxjson.FromString(json)

email := node.Get("email")
if email.IsValidEmail() {
    fmt.Printf("邮箱 %s 格式正确\n", email.StringOr(""))
}

url := node.Get("website")  
if url.IsValidURL() {
    fmt.Printf("网址 %s 格式正确\n", url.StringOr(""))
}

ip := node.Get("server")
if ip.IsValidIP() {
    fmt.Printf("IP地址 %s 格式正确\n", ip.StringOr(""))
}
```

#### 数值范围验证
```go
func (n Node) InRange(min, max float64) bool    // 检查数值是否在指定范围内
```

```go
json := `{"age": 25, "score": 95.5}`
node := fxjson.FromString(json)

age := node.Get("age")
if age.InRange(18, 65) {
    fmt.Printf("年龄 %d 在合法范围内\n", age.IntOr(0))
}

score := node.Get("score") 
if score.InRange(0, 100) {
    fmt.Printf("分数 %.1f 在有效范围内\n", score.FloatOr(0))
}
```

### 结构体操作

#### Marshal() / FastMarshal()
```go
func Marshal(v interface{}) ([]byte, error)          // 标准序列化
func FastMarshal(v interface{}) ([]byte, error)     // 高性能序列化
```

#### DecodeStruct()
```go  
func DecodeStruct(data []byte, v interface{}) error  // 解码到结构体
```

```go
// 定义结构体
type User struct {
    Name   string `json:"name"`
    Age    int    `json:"age"`
    Active bool   `json:"active"`
}

// 序列化
user := User{Name: "张三", Age: 30, Active: true}
jsonBytes, err := fxjson.FastMarshal(user)
if err == nil {
    fmt.Printf("JSON: %s\n", jsonBytes)
}

// 反序列化
var newUser User
err = fxjson.DecodeStruct(jsonBytes, &newUser)
if err == nil {
    fmt.Printf("用户: %+v\n", newUser)
}
```

### 深度遍历

#### Walk()
```go
func (n Node) Walk(fn func(path string, node Node) bool)
```

递归遍历 JSON 的所有节点，适合复杂数据结构的分析。

```go
json := `{
    "company": "Tech Corp",
    "departments": {
        "engineering": {
            "count": 25,
            "teams": ["backend", "frontend"]
        }
    }
}`
node := fxjson.FromString(json)

// 深度遍历所有节点
node.Walk(func(path string, n fxjson.Node) bool {
    if n.IsString() {
        fmt.Printf("%s = %s (字符串)\n", path, n.StringOr(""))
    } else if n.IsNumber() {
        fmt.Printf("%s = %d (数字)\n", path, n.IntOr(0))
    } else if n.IsArray() {
        fmt.Printf("%s = 数组(长度: %d)\n", path, n.Len())
    }
    return true  // 继续遍历
})
```

---

## 配置选项

### ParseOptions
```go
type ParseOptions struct {
    MaxDepth      int  // 最大嵌套深度，0 表示无限制
    MaxStringLen  int  // 最大字符串长度，0 表示无限制  
    MaxObjectKeys int  // 最大对象键数量，0 表示无限制
    MaxArrayItems int  // 最大数组项数量，0 表示无限制
    StrictMode    bool // 严格模式：拒绝格式错误的 JSON
}
```

**默认配置**：
```go
var DefaultParseOptions = ParseOptions{
    MaxDepth:      1000,        // 最大1000层嵌套
    MaxStringLen:  1024 * 1024, // 最大1MB字符串
    MaxObjectKeys: 10000,       // 最大10000个键
    MaxArrayItems: 100000,      // 最大100000个数组项
    StrictMode:    false,       // 非严格模式
}
```

### JsonParam（序列化选项）
```go
type JsonParam struct {
    Indent     int  // 缩进空格数；0 表示紧凑模式
    EscapeHTML bool // 是否转义 HTML 符号
    Precision  int  // 浮点数精度；-1 表示原样输出
}
```

---

## 性能特性

### 零分配操作
以下操作在 FxJSON 中是零内存分配的：
- `Get()` / `GetPath()` 访问
- `ArrayForEach()` / `ForEach()` 遍历
- `StringOr()` / `IntOr()` / `FloatOr()` / `BoolOr()` 转换
- `IsString()` / `IsNumber()` 等类型检查
- `Len()` / `Index()` / `Exists()` 等基础操作

### 缓存机制
FxJSON 内置智能缓存：
- 数组索引访问会被缓存，重复访问速度提升 4 倍
- 缓存是线程安全的，无锁设计
- 基于指针+范围的键，内存效率高

```go
json := `{"users": [{"name": "张三"}, {"name": "李四"}]}`
node := fxjson.FromString(json)

// 第一次访问，建立缓存
name1 := node.GetPath("users.0.name").StringOr("")  

// 第二次访问，使用缓存，速度快 4 倍
name1Again := node.GetPath("users.0.name").StringOr("")
```

---

## 错误处理指南

### 推荐模式：使用 `Or` 方法
```go
// 优雅的错误处理，无需 if err != nil
name := node.Get("name").StringOr("默认用户")
age := node.Get("age").IntOr(0)
active := node.Get("active").BoolOr(false)
```

### 严格模式：手动错误处理
```go
// 需要明确错误信息时使用
name, err := node.Get("name").String()
if err != nil {
    switch err := err.(type) {
    case *fxjson.FxJSONError:
        fmt.Printf("JSON错误: %s (位置: %d行%d列)\n", 
            err.Error(), err.Position.Line, err.Position.Column)
    default:
        fmt.Printf("未知错误: %v\n", err)
    }
    return
}
```

### 字段存在性检查
```go
// 检查字段是否存在
if node.Get("optional").Exists() {
    value := node.Get("optional").StringOr("")
    // 处理存在的字段
}

// 检查并获取值的组合用法
email := node.Get("email")
if email.Exists() && email.IsValidEmail() {
    fmt.Printf("有效邮箱: %s\n", email.StringOr(""))
}
```

---

## 最佳实践

### 1. 优先使用安全方法
```go
// 推荐：简洁且安全
name := node.Get("user").Get("name").StringOr("匿名")

// 不推荐：代码冗余
userNode := node.Get("user")
nameNode := userNode.Get("name") 
name, err := nameNode.String()
if err != nil {
    name = "匿名"
}
```

### 2. 合理使用路径访问
```go
// 深层嵌套时使用路径访问
city := node.GetPath("user.profile.address.city").StringOr("")

// 浅层访问时链式调用更清晰
name := node.Get("user").Get("name").StringOr("")
```

### 3. 高效的遍历方式
```go
// 推荐：零分配高性能遍历
users.ArrayForEach(func(i int, user fxjson.Node) bool {
    processUser(user)
    return true
})

// 不推荐：传统索引遍历
for i := 0; i < users.Len(); i++ {
    user := users.Index(i)
    processUser(user)
}
```

### 4. 适当的类型检查
```go
// 处理未知数据时先检查类型
value := node.Get("dynamic_field")
if value.IsString() {
    text := value.StringOr("")
    // 处理字符串
} else if value.IsNumber() {
    num := value.IntOr(0)
    // 处理数字
}
```

这就是 FxJSON 的完整 API 参考。从基础的解析和访问，到高级的验证和序列化，FxJSON 都为您提供了简洁高效的解决方案。建议从基础的 `Get()` 和 `StringOr()` 方法开始，逐步探索更多功能。