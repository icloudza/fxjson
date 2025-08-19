# API 参考

FxJSON 提供了完整的 API 用于 JSON 解析、序列化和操作。API 分为两类：
- **包级函数**：通过 `fxjson.Xxx` 直接调用
- **Node 方法**：通过 `node.Xxx` 调用（node 是 `fxjson.Node` 类型的实例）

## 核心类型

### Node

`Node` 是 FxJSON 的核心数据结构，表示 JSON 文档中的一个值。

```go
type Node struct {
    // 内部字段（不导出）
}
```

### NodeType

```go
type NodeType byte

const (
    TypeInvalid NodeType = 0
    TypeObject  NodeType = 'o'
    TypeArray   NodeType = 'a'
    TypeString  NodeType = 's'
    TypeNumber  NodeType = 'n'
    TypeBool    NodeType = 'b'
    TypeNull    NodeType = 'l'
)
```

## 包级函数（fxjson.Xxx）

以下函数可以直接通过 `fxjson.` 调用：

### 解析函数

#### FromBytes
```go
func FromBytes(b []byte) Node
```
从字节数组解析 JSON，返回根节点。这是最常用的解析函数。

**示例：**
```go
data := []byte(`{"name":"Alice","age":30}`)
node := fxjson.FromBytes(data)
```

#### FromBytesWithOptions
```go
func FromBytesWithOptions(b []byte, opts ParseOptions) Node
```
使用指定选项解析 JSON，可以设置安全限制。

**示例：**
```go
opts := fxjson.ParseOptions{
    MaxDepth: 100,
    MaxStringLen: 10000,
}
node := fxjson.FromBytesWithOptions(data, opts)
```


### 序列化函数

#### Marshal
```go
func Marshal(v interface{}) ([]byte, error)
```
将 Go 值序列化为 JSON 字节切片。

**示例：**
```go
data := map[string]interface{}{
    "name": "Alice",
    "age": 30,
}
jsonBytes, err := fxjson.Marshal(data)
```

#### MarshalIndent
```go
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)
```
序列化为格式化的 JSON，带缩进。

**示例：**
```go
jsonBytes, err := fxjson.MarshalIndent(data, "", "  ")
```

#### FastMarshal
```go
func FastMarshal(v interface{}) []byte
```
极速序列化，跳过错误检查，追求极致性能。

**示例：**
```go
jsonBytes := fxjson.FastMarshal(data)
```

#### MarshalWithOptions
```go
func MarshalWithOptions(v interface{}, opts SerializeOptions) ([]byte, error)
```
使用自定义选项序列化。

**示例：**
```go
opts := fxjson.SerializeOptions{
    Indent: "  ",
    SortKeys: true,
}
jsonBytes, err := fxjson.MarshalWithOptions(data, opts)
```

#### MarshalToString
```go
func MarshalToString(v interface{}) (string, error)
```
序列化为字符串。

**示例：**
```go
jsonStr, err := fxjson.MarshalToString(data)
```

#### MarshalToStringWithOptions
```go
func MarshalToStringWithOptions(v interface{}, opts SerializeOptions) (string, error)
```
使用选项序列化为字符串。


### 解码函数

#### DecodeStruct
```go
func DecodeStruct(data []byte, v any) error
```
将 JSON 字节切片直接解码到结构体，避免 Node 创建开销。

**示例：**
```go
var user User
err := fxjson.DecodeStruct(jsonData, &user)
```

#### DecodeStructFast
```go
func DecodeStructFast(data []byte, v any) error
```
极致优化的结构体解码函数。

**示例：**
```go
var user User
err := fxjson.DecodeStructFast(jsonData, &user)
```



## Node 方法（node.Xxx）

以下方法需要在 `fxjson.Node` 实例上调用：

### 访问方法

#### Get
```go
func (n Node) Get(key string) Node
```
获取对象的字段。

**示例：**
```go
node := fxjson.FromBytes(data)
nameNode := node.Get("name")  // 注意：是 node.Get，不是 fxjson.Get
```

#### GetPath
```go
func (n Node) GetPath(path string) Node
```
通过路径获取嵌套值，支持点号和数组索引。

**示例：**
```go
node := fxjson.FromBytes(data)
cityNode := node.GetPath("address.city")
firstItem := node.GetPath("items[0]")
```


#### Index
```go
func (n Node) Index(i int) Node
```
获取数组的第 i 个元素（O(1) 复杂度）。

**示例：**
```go
arrayNode := node.Get("items")
firstItem := arrayNode.Index(0)
```

### 类型转换方法

这些方法用于将 Node 转换为 Go 的基本类型：

#### String
```go
func (n Node) String() (string, error)
```
获取字符串值。

**示例：**
```go
node := fxjson.FromBytes(data)
name, err := node.Get("name").String()
```

#### Int
```go
func (n Node) Int() (int64, error)
```
获取整数值。

**示例：**
```go
age, err := node.Get("age").Int()
```

#### Uint
```go
func (n Node) Uint() (uint64, error)
```
获取无符号整数值。

**示例：**
```go
id, err := node.Get("id").Uint()
```

#### Float
```go
func (n Node) Float() (float64, error)
```
获取浮点数值。

**示例：**
```go
price, err := node.Get("price").Float()
```

#### Bool
```go
func (n Node) Bool() (bool, error)
```
获取布尔值。

**示例：**
```go
isActive, err := node.Get("isActive").Bool()
```

### 便捷方法（带默认值）

#### StringOr
```go
func (n Node) StringOr(defaultValue string) string
```
获取字符串值，如果失败返回默认值。

**示例：**
```go
name := node.Get("name").StringOr("未知")
```

#### IntOr
```go
func (n Node) IntOr(defaultValue int64) int64
```
获取整数值，如果失败返回默认值。

**示例：**
```go
age := node.Get("age").IntOr(0)
```

#### UintOr
```go
func (n Node) UintOr(defaultValue uint64) uint64
```
获取无符号整数值，如果失败返回默认值。

**示例：**
```go
id := node.Get("id").UintOr(0)
```

#### FloatOr
```go
func (n Node) FloatOr(defaultValue float64) float64
```
获取浮点数值，如果失败返回默认值。

**示例：**
```go
price := node.Get("price").FloatOr(0.0)
```

#### BoolOr
```go
func (n Node) BoolOr(defaultValue bool) bool
```
获取布尔值，如果失败返回默认值。

**示例：**
```go
isActive := node.Get("isActive").BoolOr(false)
```

#### NumStr
```go
func (n Node) NumStr() (string, error)
```
获取数字的原始字符串表示。

**示例：**
```go
numStr, err := node.Get("amount").NumStr()
```

### 类型检查方法

这些方法用于检查 Node 的类型：

#### Exists
```go
func (n Node) Exists() bool
```
检查节点是否存在且有效。

**示例：**
```go
if node.Get("email").Exists() {
    // 字段存在
}
```

#### IsObject
```go
func (n Node) IsObject() bool
```
检查是否为对象类型。

#### IsArray
```go
func (n Node) IsArray() bool
```
检查是否为数组类型。

#### IsString
```go
func (n Node) IsString() bool
```
检查是否为字符串类型。

#### IsNumber
```go
func (n Node) IsNumber() bool
```
检查是否为数字类型。

#### IsBool
```go
func (n Node) IsBool() bool
```
检查是否为布尔类型。

#### IsNull
```go
func (n Node) IsNull() bool
```
检查是否为 null 值。

#### IsScalar
```go
func (n Node) IsScalar() bool
```
检查是否为标量类型（字符串、数字、布尔或 null）。

#### IsContainer
```go
func (n Node) IsContainer() bool
```
检查是否为容器类型（对象或数组）。

### 遍历方法

#### ForEach
```go
func (n Node) ForEach(fn func(key string, value Node) bool)
```
遍历对象的所有键值对。

**示例：**
```go
node.ForEach(func(key string, value Node) bool {
    fmt.Printf("%s: %v\n", key, value.Raw())
    return true // 继续遍历
})
```

#### ArrayForEach
```go
func (n Node) ArrayForEach(fn func(index int, value Node) bool)
```
遍历数组的所有元素。

**示例：**
```go
arrayNode.ArrayForEach(func(index int, item Node) bool {
    fmt.Printf("[%d]: %v\n", index, item.Raw())
    return true
})
```

#### Walk
```go
func (n Node) Walk(fn func(path string, node Node) bool)
```
深度优先遍历整个 JSON 树。

**示例：**
```go
node.Walk(func(path string, n Node) bool {
    fmt.Printf("%s: %v\n", path, n.Raw())
    return true
})
```

### 数据转换方法

#### ToStringSlice
```go
func (n Node) ToStringSlice() ([]string, error)
```
将数组转换为字符串切片。

**示例：**
```go
tags, err := node.Get("tags").ToStringSlice()
```

#### ToIntSlice
```go
func (n Node) ToIntSlice() ([]int64, error)
```
将数组转换为整数切片。

**示例：**
```go
numbers, err := node.Get("numbers").ToIntSlice()
```

#### ToFloatSlice
```go
func (n Node) ToFloatSlice() ([]float64, error)
```
将数组转换为浮点数切片。

**示例：**
```go
prices, err := node.Get("prices").ToFloatSlice()
```

#### ToBoolSlice
```go
func (n Node) ToBoolSlice() ([]bool, error)
```
将数组转换为布尔值切片。

**示例：**
```go
flags, err := node.Get("flags").ToBoolSlice()
```

### 数据验证方法

#### IsValidEmail
```go
func (n Node) IsValidEmail() bool
```
检查字符串是否为有效的电子邮件地址。

**示例：**
```go
if node.Get("email").IsValidEmail() {
    fmt.Println("邮箱格式正确")
}
```

#### IsValidURL
```go
func (n Node) IsValidURL() bool
```
检查字符串是否为有效的URL。

**示例：**
```go
if node.Get("website").IsValidURL() {
    fmt.Println("URL格式正确")
}
```

#### IsValidIP
```go
func (n Node) IsValidIP() bool
```
检查字符串是否为有效的IP地址（IPv4或IPv6）。

**示例：**
```go
if node.Get("ip").IsValidIP() {
    fmt.Println("IP地址格式正确")
}
```

#### IsValidIPv4
```go
func (n Node) IsValidIPv4() bool
```
检查字符串是否为有效的IPv4地址。

**示例：**
```go
if node.Get("ip").IsValidIPv4() {
    fmt.Println("这是IPv4地址")
}
```

#### IsValidIPv6
```go
func (n Node) IsValidIPv6() bool
```
检查字符串是否为有效的IPv6地址。

**示例：**
```go
if node.Get("ip").IsValidIPv6() {
    fmt.Println("这是IPv6地址")
}
```

#### IsValidPhone
```go
func (n Node) IsValidPhone() bool
```
检查字符串是否为有效的电话号码（E.164格式）。

**示例：**
```go
if node.Get("phone").IsValidPhone() {
    fmt.Println("电话格式正确")
}
```

#### IsValidUUID
```go
func (n Node) IsValidUUID() bool
```
检查字符串是否为有效的UUID。

**示例：**
```go
if node.Get("uuid").IsValidUUID() {
    fmt.Println("UUID格式正确")
}
```

### 类型信息方法

#### Kind
```go
func (n Node) Kind() NodeType
```
获取节点的类型。

**示例：**
```go
switch node.Kind() {
case fxjson.TypeString:
    fmt.Println("这是字符串")
case fxjson.TypeNumber:
    fmt.Println("这是数字")
}
```

#### Type
```go
func (n Node) Type() byte
```
获取节点的内部类型字节。

**示例：**
```go
if node.Type() == 's' {
    fmt.Println("这是字符串")
}
```

### 原始数据方法

#### RawString
```go
func (n Node) RawString() (string, error)
```
获取节点的原始JSON字符串形式。

**示例：**
```go
rawStr, err := node.RawString()
if err == nil {
    fmt.Printf("原始JSON: %s\n", rawStr)
}
```

#### Json
```go
func (n Node) Json() (string, error)
```
获取对象或数组节点的JSON表示。

**示例：**
```go
jsonStr, err := node.Get("config").Json()
if err == nil {
    fmt.Printf("配置JSON: %s\n", jsonStr)
}
```

#### FloatString
```go
func (n Node) FloatString() (string, error)
```
获取数字节点的字符串表示，保持原始精度。

**示例：**
```go
floatStr, err := node.Get("price").FloatString()
if err == nil {
    fmt.Printf("价格: %s\n", floatStr)
}
```

### JSON序列化方法

#### ToJSON
```go
func (n Node) ToJSON() (string, error)
```
将节点序列化为JSON字符串（压缩模式）。

**示例：**
```go
jsonStr, err := node.ToJSON()
if err == nil {
    fmt.Printf("JSON: %s\n", jsonStr)
}
```

#### ToJSONIndent
```go
func (n Node) ToJSONIndent(prefix, indent string) (string, error)
```
将节点序列化为格式化的JSON字符串。

**示例：**
```go
jsonStr, err := node.ToJSONIndent("", "  ")
if err == nil {
    fmt.Printf("格式化JSON:\n%s\n", jsonStr)
}
```

#### ToJSONWithOptions
```go
func (n Node) ToJSONWithOptions(opts SerializeOptions) (string, error)
```
使用自定义选项序列化节点。

**示例：**
```go
opts := fxjson.SerializeOptions{
    Indent: "  ",
    SortKeys: true,
}
jsonStr, err := node.ToJSONWithOptions(opts)
```

#### ToJSONBytes
```go
func (n Node) ToJSONBytes() ([]byte, error)
```
将节点序列化为JSON字节切片。

**示例：**
```go
jsonBytes, err := node.ToJSONBytes()
if err == nil {
    fmt.Printf("JSON字节: %v\n", jsonBytes)
}
```

#### ToJSONBytesWithOptions
```go
func (n Node) ToJSONBytesWithOptions(opts SerializeOptions) ([]byte, error)
```
使用自定义选项序列化为字节切片。

**示例：**
```go
opts := fxjson.SerializeOptions{EscapeHTML: true}
jsonBytes, err := node.ToJSONBytesWithOptions(opts)
```

#### ToJSONFast
```go
func (n Node) ToJSONFast() string
```
快速序列化节点为JSON字符串（最小开销）。

**示例：**
```go
jsonStr := node.ToJSONFast()
fmt.Printf("快速序列化: %s\n", jsonStr)
```

### 字符串操作方法

#### Contains
```go
func (n Node) Contains(substr string) bool
```
检查字符串是否包含子串。

**示例：**
```go
if node.Get("description").Contains("important") {
    fmt.Println("描述包含important")
}
```

#### StartsWith
```go
func (n Node) StartsWith(prefix string) bool
```
检查字符串是否以指定前缀开始。

**示例：**
```go
if node.Get("url").StartsWith("https://") {
    fmt.Println("这是HTTPS链接")
}
```

#### EndsWith
```go
func (n Node) EndsWith(suffix string) bool
```
检查字符串是否以指定后缀结束。

**示例：**
```go
if node.Get("filename").EndsWith(".json") {
    fmt.Println("这是JSON文件")
}
```

#### ToLower
```go
func (n Node) ToLower() (string, error)
```
将字符串转换为小写。

**示例：**
```go
lowerStr, err := node.Get("name").ToLower()
if err == nil {
    fmt.Printf("小写: %s\n", lowerStr)
}
```

#### ToUpper
```go
func (n Node) ToUpper() (string, error)
```
将字符串转换为大写。

**示例：**
```go
upperStr, err := node.Get("name").ToUpper()
if err == nil {
    fmt.Printf("大写: %s\n", upperStr)
}
```

#### Trim
```go
func (n Node) Trim() (string, error)
```
去除字符串两端的空白字符。

**示例：**
```go
trimmedStr, err := node.Get("input").Trim()
if err == nil {
    fmt.Printf("去空白: '%s'\n", trimmedStr)
}
```

### 数组操作方法

#### First
```go
func (n Node) First() Node
```
获取数组的第一个元素。

**示例：**
```go
firstItem := node.Get("items").First()
if firstItem.Exists() {
    fmt.Printf("第一个元素: %v\n", firstItem.Raw())
}
```

#### Last
```go
func (n Node) Last() Node
```
获取数组的最后一个元素。

**示例：**
```go
lastItem := node.Get("items").Last()
if lastItem.Exists() {
    fmt.Printf("最后一个元素: %v\n", lastItem.Raw())
}
```

#### Slice
```go
func (n Node) Slice(start, end int) []Node
```
获取数组的切片（包含start，不包含end）。

**示例：**
```go
items := node.Get("items").Slice(1, 3)
for i, item := range items {
    fmt.Printf("Item %d: %v\n", i, item.Raw())
}
```

#### Reverse
```go
func (n Node) Reverse() []Node
```
返回反转后的数组节点切片。

**示例：**
```go
reversedItems := node.Get("items").Reverse()
for i, item := range reversedItems {
    fmt.Printf("Reversed %d: %v\n", i, item.Raw())
}
```

#### GetAllValues
```go
func (n Node) GetAllValues() []Node
```
获取数组的所有元素节点。

**示例：**
```go
allItems := node.Get("items").GetAllValues()
fmt.Printf("数组长度: %d\n", len(allItems))
```

#### ToSlice
```go
func (n Node) ToSlice() []Node
```
将数组节点转换为Node切片。

**示例：**
```go
nodeSlice := arrayNode.ToSlice()
fmt.Printf("切片长度: %d\n", len(nodeSlice))
```

### 对象操作方法

#### GetAllKeys
```go
func (n Node) GetAllKeys() []string
```
获取对象的所有键名（字符串形式）。

**示例：**
```go
keys := node.GetAllKeys()
for _, key := range keys {
    fmt.Printf("键: %s\n", key)
}
```

#### ToMap
```go
func (n Node) ToMap() map[string]Node
```
将对象节点转换为map[string]Node。

**示例：**
```go
nodeMap := objectNode.ToMap()
for key, value := range nodeMap {
    fmt.Printf("%s: %v\n", key, value.Raw())
}
```

#### Merge
```go
func (n Node) Merge(other Node) map[string]Node
```
合并两个对象节点（浅合并）。

**示例：**
```go
merged := node1.Merge(node2)
for key, value := range merged {
    fmt.Printf("%s: %v\n", key, value.Raw())
}
```

#### Pick
```go
func (n Node) Pick(keys ...string) map[string]Node
```
从对象中选择指定的键。

**示例：**
```go
selected := node.Pick("name", "age", "email")
for key, value := range selected {
    fmt.Printf("%s: %v\n", key, value.Raw())
}
```

#### Omit
```go
func (n Node) Omit(keys ...string) map[string]Node
```
从对象中排除指定的键。

**示例：**
```go
filtered := node.Omit("password", "secret")
for key, value := range filtered {
    fmt.Printf("%s: %v\n", key, value.Raw())
}
```

#### HasKey
```go
func (n Node) HasKey(key string) bool
```
检查对象是否包含指定键。

**示例：**
```go
if node.HasKey("email") {
    fmt.Println("对象包含email字段")
}
```

#### GetKeyValue
```go
func (n Node) GetKeyValue(key string, defaultValue Node) Node
```
获取对象中指定键的值，如果不存在返回默认值。

**示例：**
```go
defaultNode := fxjson.FromBytes([]byte(`"default"`))
value := node.GetKeyValue("optional_field", defaultNode)
```

### 批量获取方法

#### GetMultiple
```go
func (n Node) GetMultiple(paths ...string) []Node
```
同时获取多个路径的值。

**示例：**
```go
nodes := node.GetMultiple("name", "age", "email")
for i, n := range nodes {
    fmt.Printf("Path %d: %v\n", i, n.Raw())
}
```

#### HasAnyPath
```go
func (n Node) HasAnyPath(paths ...string) bool
```
检查是否存在任意一个路径。

**示例：**
```go
if node.HasAnyPath("email", "phone", "contact") {
    fmt.Println("至少有一种联系方式")
}
```

#### HasAllPaths
```go
func (n Node) HasAllPaths(paths ...string) bool
```
检查是否存在所有路径。

**示例：**
```go
if node.HasAllPaths("name", "age", "email") {
    fmt.Println("所有必需字段都存在")
}
```

### 查找和过滤方法

#### FindInObject
```go
func (n Node) FindInObject(predicate func(key string, value Node) bool) (string, Node, bool)
```
在对象中查找满足条件的第一个键值对。

**示例：**
```go
key, value, found := node.FindInObject(func(k string, v Node) bool {
    return v.IsString() && v.StringOr("") == "target"
})
if found {
    fmt.Printf("找到: %s = %v\n", key, value.Raw())
}
```

#### FindInArray
```go
func (n Node) FindInArray(predicate func(index int, value Node) bool) (int, Node, bool)
```
在数组中查找满足条件的第一个元素。

**示例：**
```go
index, value, found := arrayNode.FindInArray(func(i int, v Node) bool {
    return v.Get("id").IntOr(0) == 123
})
if found {
    fmt.Printf("找到在索引 %d: %v\n", index, value.Raw())
}
```

#### FilterArray
```go
func (n Node) FilterArray(predicate func(index int, value Node) bool) []Node
```
过滤数组元素，返回满足条件的所有元素。

**示例：**
```go
filtered := arrayNode.FilterArray(func(i int, v Node) bool {
    return v.Get("active").BoolOr(false)
})
fmt.Printf("筛选出 %d 个活跃元素\n", len(filtered))
```

#### FindByPath
```go
func (n Node) FindByPath(path string) Node
```
根据路径查找节点（等同于GetPath）。

**示例：**
```go
foundNode := node.FindByPath("user.profile.name")
if foundNode.Exists() {
    fmt.Printf("找到: %v\n", foundNode.Raw())
}
```

### 统计和分析方法

#### CountIf
```go
func (n Node) CountIf(predicate func(index int, value Node) bool) int
```
统计数组中满足条件的元素个数。

**示例：**
```go
count := arrayNode.CountIf(func(i int, v Node) bool {
    return v.Get("score").FloatOr(0) > 80
})
fmt.Printf("高分学生数量: %d\n", count)
```

#### AllMatch
```go
func (n Node) AllMatch(predicate func(index int, value Node) bool) bool
```
检查数组中是否所有元素都满足条件。

**示例：**
```go
allPassed := arrayNode.AllMatch(func(i int, v Node) bool {
    return v.Get("score").FloatOr(0) >= 60
})
if allPassed {
    fmt.Println("所有学生都及格了")
}
```

#### AnyMatch
```go
func (n Node) AnyMatch(predicate func(index int, value Node) bool) bool
```
检查数组中是否有任何元素满足条件。

**示例：**
```go
hasExcellent := arrayNode.AnyMatch(func(i int, v Node) bool {
    return v.Get("score").FloatOr(0) >= 95
})
if hasExcellent {
    fmt.Println("有优秀学生")
}
```

### 比较和状态检查方法

#### Equals
```go
func (n Node) Equals(other Node) bool
```
检查两个节点是否相等。

**示例：**
```go
if node1.Equals(node2) {
    fmt.Println("两个节点相等")
}
```

#### IsEmpty
```go
func (n Node) IsEmpty() bool
```
检查节点是否为空（空字符串、空数组、空对象、null）。

**示例：**
```go
if node.Get("description").IsEmpty() {
    fmt.Println("描述为空")
}
```

### 数字操作方法

#### IsPositive
```go
func (n Node) IsPositive() bool
```
检查数字是否为正数。

**示例：**
```go
if node.Get("balance").IsPositive() {
    fmt.Println("余额为正")
}
```

#### IsNegative
```go
func (n Node) IsNegative() bool
```
检查数字是否为负数。

**示例：**
```go
if node.Get("delta").IsNegative() {
    fmt.Println("变化为负")
}
```

#### IsZero
```go
func (n Node) IsZero() bool
```
检查数字是否为零。

**示例：**
```go
if node.Get("count").IsZero() {
    fmt.Println("计数为零")
}
```

#### IsInteger
```go
func (n Node) IsInteger() bool
```
检查数字是否为整数。

**示例：**
```go
if node.Get("age").IsInteger() {
    fmt.Println("年龄是整数")
}
```

#### InRange
```go
func (n Node) InRange(min, max float64) bool
```
检查数字是否在指定范围内（包含边界）。

**示例：**
```go
if node.Get("score").InRange(0, 100) {
    fmt.Println("分数在有效范围内")
}
```

### 高级查询方法

#### Query
```go
func (n Node) Query() *QueryBuilder
```
创建查询构建器，用于复杂的数组查询。

**示例：**
```go
results, err := node.Get("users").Query().
    Where("age", ">", 18).
    Where("active", "=", true).
    SortBy("name", "asc").
    Limit(10).
    ToSlice()
```

#### Aggregate
```go
func (n Node) Aggregate() *Aggregator
```
创建聚合器，用于数据统计。

**示例：**
```go
result, err := node.Get("sales").Aggregate().
    Sum("amount", "total_sales").
    Avg("amount", "avg_sale").
    Count("order_count").
    Execute(node.Get("sales"))
```

#### Transform
```go
func (n Node) Transform(mapper FieldMapper) (map[string]interface{}, error)
```
数据变换，应用字段映射规则。

**示例：**
```go
mapper := fxjson.FieldMapper{
    Rules: map[string]string{
        "user_name": "name",
        "user_age": "age",
    },
}
transformed, err := node.Transform(mapper)
```

#### Validate
```go
func (n Node) Validate(validator *DataValidator) (map[string]interface{}, []error)
```
数据验证，应用验证规则。

**示例：**
```go
validator := &fxjson.DataValidator{
    Rules: map[string]fxjson.ValidationRule{
        "name": {Required: true, Type: "string"},
        "age": {Required: true, Type: "number", Min: 0, Max: 150},
    },
}
result, errors := node.Validate(validator)
```

#### Stream
```go
func (n Node) Stream(processor func(Node, int) bool) error
```
流式处理数组元素。

**示例：**
```go
err := arrayNode.Stream(func(item Node, index int) bool {
    fmt.Printf("处理第 %d 个元素: %v\n", index, item.Raw())
    return true // 继续处理
})
```

### 其他方法

#### Len
```go
func (n Node) Len() int
```
获取数组或对象的长度。

**示例：**
```go
length := arrayNode.Len()
```

#### Keys
```go
func (n Node) Keys() [][]byte
```
获取对象的所有键（字节切片形式）。

**示例：**
```go
keys := objectNode.Keys()
```

#### Raw
```go
func (n Node) Raw() []byte
```
获取节点的原始 JSON 字节。

**示例：**
```go
rawJSON := node.Raw()
```

#### Decode
```go
func (n Node) Decode(v any) error
```
将节点解码到 Go 值。

**示例：**
```go
var user User
err := node.Decode(&user)
```

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

### SerializeOptions

```go
type SerializeOptions struct {
    Indent          string // 缩进字符串，空字符串表示压缩模式
    EscapeHTML      bool   // 是否转义HTML字符 (<, >, &)
    SortKeys        bool   // 是否对对象键进行排序
    OmitEmpty       bool   // 是否忽略空值
    FloatPrecision  int    // 浮点数精度，-1表示默认
    UseNumberString bool   // 大数字是否用字符串表示
}
```

## 完整使用示例

```go
package main

import (
   "fmt"
   "github.com/icloudza/fxjson"
)

func main() {
   // JSON 数据
   jsonData := []byte(`{
        "user": {
            "name": "Alice",
            "age": 30,
            "emails": ["alice@example.com", "alice@work.com"]
        }
    }`)

   // 1. 解析 JSON（包级函数）
   node := fxjson.FromBytes(jsonData)

   // 2. 访问数据（Node 方法）
   userNode := node.Get("user")     // 获取 user 对象
   nameNode := userNode.Get("name") // 获取 name 字段

   // 3. 类型转换（Node 方法）
   name, _ := nameNode.String()        // 转换为字符串
   age, _ := userNode.Get("age").Int() // 转换为整数

   // 4. 数组操作（Node 方法）
   emails := userNode.Get("emails")
   _, _ = emails.Index(0).String()

   // 5. 遍历数组（Node 方法）
   emails.ArrayForEach(func(index int, email fxjson.Node) bool {
      emailStr, _ := email.String()
      fmt.Printf("Email %d: %s\n", index, emailStr)
      return true
   })

   // 6. 序列化（包级函数）
   newData := map[string]interface{}{
      "name": name,
      "age":  age,
   }
   jsonBytes, _ := fxjson.Marshal(newData)
   fmt.Printf("序列化结果: %s\n", jsonBytes)

   // 7. 美化输出（包级函数）
   pretty := fxjson.PrettyJSON(jsonBytes)
   fmt.Printf("美化输出:\n%s\n", pretty)
}

```

## 性能相关

### 基准测试结果

在 Apple M4 Pro 上的性能表现：

| 操作                | 耗时          | 内存分配   | 分配次数        |
|-------------------|-------------|--------|-------------|
| Get               | 22.44 ns/op | 0 B/op | 0 allocs/op |
| GetPath/GetByPath | 92.77 ns/op | 0 B/op | 0 allocs/op |
| Index             | 18.57 ns/op | 0 B/op | 0 allocs/op |
| Len               | 18.57 ns/op | 0 B/op | 0 allocs/op |
| ForEach           | 150.3 ns/op | 0 B/op | 0 allocs/op |

## 使用建议

1. **明确区分包级函数和 Node 方法**
   - 包级函数：`fxjson.FromBytes()`, `fxjson.Marshal()` 等
   - Node 方法：`node.Get()`, `node.String()` 等

2. **优先使用 FromBytes**：相比字符串，字节数组避免了额外的内存分配

3. **使用 FastMarshal 提升性能**：在确保数据有效的情况下，使用快速序列化函数

4. **使用 DecodeStruct/DecodeStructFast**：直接解码到结构体，避免创建 Node

5. **错误处理**：始终检查 Exists() 或处理返回的错误

6. **使用 GetPath**：访问深层嵌套数据时，比链式 Get 更高效

7. **零分配遍历**：使用 ForEach 和 ArrayForEach 进行零分配遍历