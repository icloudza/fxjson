# 基础概念

本节介绍 FxJSON 的核心概念和设计理念，帮助您更好地理解和使用这个高性能 JSON 解析库。

## 核心概念

### Node（节点）

`Node` 是 FxJSON 的核心数据结构，代表 JSON 数据中的任意一个节点。无论是对象、数组、字符串、数字、布尔值还是 null，都用 Node 来表示。

```go
type Node struct {
    // 内部实现细节（用户无需关心）
}
```

#### Node 的特点

1. **不可变性**: Node 创建后内容不会改变，保证线程安全
2. **轻量级**: Node 本身只是对原始数据的引用，不复制数据
3. **零分配**: 大多数操作都是零内存分配的
4. **类型安全**: 提供类型检查和安全的类型转换

### 节点类型

FxJSON 支持所有标准 JSON 数据类型：

```go
type NodeType byte

const (
    TypeInvalid NodeType = 0    // 无效节点
    TypeObject  NodeType = 'o'  // JSON 对象 {}
    TypeArray   NodeType = 'a'  // JSON 数组 []
    TypeString  NodeType = 's'  // 字符串 "text"
    TypeNumber  NodeType = 'n'  // 数字 123 或 123.45
    TypeBool    NodeType = 'b'  // 布尔值 true/false
    TypeNull    NodeType = 'l'  // null 值
)
```

#### 类型检查示例

```go
node := fxjson.FromBytes([]byte(`{"name": "张三", "age": 30, "tags": ["Go", "JSON"]}`))

// 检查根节点类型
if node.IsObject() {
    fmt.Println("根节点是对象")
}

// 检查字段类型
if node.Get("name").IsString() {
    fmt.Println("name 是字符串")
}

if node.Get("age").IsNumber() {
    fmt.Println("age 是数字")
}

if node.Get("tags").IsArray() {
    fmt.Println("tags 是数组")
}
```

## 设计理念

### 1. 性能优先

FxJSON 的设计以性能为首要目标：

- **零分配操作**: 核心访问操作不产生内存分配
- **原地解析**: 直接在原始字节数据上操作，避免数据复制
- **智能缓存**: 自动缓存重复访问的路径，提升性能
- **优化算法**: 使用高效的解析和查找算法

### 2. 安全性

- **边界检查**: 所有操作都有完备的边界检查
- **类型安全**: 强类型接口，避免运行时类型错误
- **默认值机制**: 提供安全的默认值，避免程序崩溃
- **错误处理**: 两种错误处理模式满足不同需求

### 3. 易用性

- **链式调用**: 支持链式调用，代码更简洁
- **直观 API**: API 设计直观，学习成本低
- **丰富功能**: 内置验证、查询、聚合等高级功能

## 数据访问模式

### 安全模式（推荐）

使用 `Or` 系列方法，自动处理错误并提供默认值：

```go
node := fxjson.FromBytes([]byte(`{"user": {"name": "张三", "age": 30}}`))

// 安全访问，即使字段不存在也不会出错
name := node.GetPath("user.name").StringOr("未知用户")
age := node.GetPath("user.age").IntOr(0)
email := node.GetPath("user.email").StringOr("未提供") // 字段不存在，返回默认值

fmt.Printf("用户: %s, 年龄: %d, 邮箱: %s\n", name, age, email)
// 输出: 用户: 张三, 年龄: 30, 邮箱: 未提供
```

### 严格模式

使用无后缀方法，需要手动处理错误：

```go
name, err := node.GetPath("user.name").String()
if err != nil {
    fmt.Printf("获取用户名失败: %v\n", err)
    return
}

age, err := node.GetPath("user.age").Int()
if err != nil {
    fmt.Printf("获取年龄失败: %v\n", err)
    return
}
```

## 路径访问

FxJSON 支持点号分隔的路径访问，简化嵌套数据的获取：

```go
jsonData := `{
    "company": {
        "name": "科技公司",
        "address": {
            "city": "北京",
            "district": "朝阳区"
        },
        "employees": [
            {"name": "张三", "position": "工程师"},
            {"name": "李四", "position": "设计师"}
        ]
    }
}`

node := fxjson.FromBytes([]byte(jsonData))

// 路径访问
companyName := node.GetPath("company.name").StringOr("")
city := node.GetPath("company.address.city").StringOr("")
firstEmployee := node.GetPath("company.employees.0.name").StringOr("")

fmt.Printf("公司: %s, 城市: %s, 第一个员工: %s\n", 
    companyName, city, firstEmployee)
```

### 路径语法

- `.` : 对象属性分隔符
- `数字` : 数组索引（从0开始）
- 路径示例:
  - `user.name` : 获取 user 对象的 name 属性
  - `users.0.email` : 获取 users 数组第一个元素的 email 属性
  - `config.database.host` : 多层嵌套访问

## 遍历机制

### 零分配遍历

FxJSON 的遍历操作是零分配的，性能极高：

```go
users := node.Get("users")

// 零分配数组遍历
users.ArrayForEach(func(index int, user fxjson.Node) bool {
    name := user.Get("name").StringOr("")
    fmt.Printf("用户 %d: %s\n", index, name)
    return true // 返回 true 继续遍历，false 停止
})

// 零分配对象遍历
config := node.Get("config")
config.ForEach(func(key string, value fxjson.Node) bool {
    fmt.Printf("%s = %s\n", key, value.StringOr(""))
    return true
})
```

### 深度遍历

```go
// 遍历所有节点
node.Walk(func(path string, value fxjson.Node) bool {
    if value.IsString() {
        fmt.Printf("路径 %s: %s\n", path, value.StringOr(""))
    }
    return true
})
```

## 缓存机制

### 自动缓存

FxJSON 会自动缓存数组索引，提升重复访问性能：

```go
users := node.Get("users")

// 第一次访问，建立缓存
user1 := users.Index(0)

// 后续访问使用缓存，速度更快
user2 := users.Index(1)
user3 := users.Index(2)
```

### 显式缓存

对于需要大量重复访问的数据，可以显式启用缓存：

```go
// 获取节点进行重复访问
cachedNode := node

// 重复访问同一路径时，性能提升显著
for i := 0; i < 1000; i++ {
    _ = cachedNode.GetPath("config.app.features.auth.enabled").BoolOr(false)
}
```

## 内存模型

### 零拷贝设计

FxJSON 采用零拷贝设计，不复制原始 JSON 数据：

```go
// 原始数据
jsonBytes := []byte(`{"key": "value"}`)

// 解析时不复制数据，只是创建引用
node := fxjson.FromBytes(jsonBytes)

// 所有操作都在原始数据上进行
value := node.Get("key").StringOr("")
```

### 内存安全

尽管是零拷贝，FxJSON 仍然保证内存安全：

- 边界检查防止越界访问
- 引用计数确保数据有效性
- 类型检查避免错误转换

## 错误处理策略

### 分层错误处理

FxJSON 提供分层的错误处理策略：

1. **节点级别**: 节点本身是否有效
2. **类型级别**: 类型转换是否成功
3. **值级别**: 值是否符合预期

```go
node := fxjson.FromBytes([]byte(`{"age": "not_a_number"}`))

// 1. 检查节点是否存在
if !node.Get("age").Exists() {
    fmt.Println("age 字段不存在")
}

// 2. 检查类型
if !node.Get("age").IsNumber() {
    fmt.Println("age 不是数字类型")
}

// 3. 安全转换
age := node.Get("age").IntOr(0) // 转换失败时返回默认值 0
```

### 错误类型

```go
var (
    ErrInvalidJSON     = errors.New("invalid JSON")
    ErrKeyNotFound     = errors.New("key not found")
    ErrTypeMismatch    = errors.New("type mismatch")
    ErrIndexOutOfRange = errors.New("index out of range")
    ErrMaxDepthExceeded = errors.New("max depth exceeded")
)
```

## 线程安全

### 只读操作

FxJSON 的所有读取操作都是线程安全的：

```go
node := fxjson.FromBytes([]byte(`{"counter": 0}`))

// 多个 goroutine 同时读取是安全的
go func() {
    for i := 0; i < 1000; i++ {
        _ = node.Get("counter").IntOr(0)
    }
}()

go func() {
    for i := 0; i < 1000; i++ {
        _ = node.Get("counter").IntOr(0)
    }
}()
```

### 缓存并发

缓存操作使用了线程安全的实现：

```go
// 多个 goroutine 同时访问是安全的
cached := node

go func() {
    for i := 0; i < 1000; i++ {
        _ = cached.GetPath("deep.nested.value").StringOr("")
    }
}()
```

## 性能特征

### 时间复杂度

- **键查找**: O(1) - 基于哈希或直接索引
- **数组访问**: O(1) - 直接索引访问
- **路径访问**: O(k) - k 为路径深度
- **遍历**: O(n) - n 为元素数量

### 空间复杂度

- **解析**: O(1) - 零拷贝，不分配额外空间
- **缓存**: O(m) - m 为缓存的路径数量
- **索引**: O(n) - n 为数组元素数量

## 最佳实践

### 1. 选择合适的访问模式

```go
// 对于可能不存在的字段，使用安全模式
email := user.Get("email").StringOr("未提供")

// 对于必须存在的字段，可以使用严格模式
id, err := user.Get("id").Int()
if err != nil {
    return fmt.Errorf("用户ID无效: %w", err)
}
```

### 2. 合理使用缓存

```go
// 对于一次性访问，不需要缓存
quickValue := node.Get("status").StringOr("")

// 对于重复访问，直接使用节点
if needRepeatedAccess {
    cached := node
    // 使用 cached 进行后续操作
}
```

### 3. 优先使用遍历而非索引

```go
// 推荐：零分配遍历
items.ArrayForEach(func(index int, item fxjson.Node) bool {
    processItem(item)
    return true
})

// 不推荐：重复索引访问
for i := 0; i < items.Len(); i++ {
    item := items.Index(i)
    processItem(item)
}
```

理解这些基础概念将帮助您更有效地使用 FxJSON，充分发挥其高性能的优势。