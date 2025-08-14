[![Go Doc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/icloudza/fxjson?utm_source=godoc)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Performance](https://img.shields.io/badge/performance-fast-orange.svg)](#性能对比)

[📄 English Documentation](README.md)

FxJSON 是一个专注性能的Go JSON解析库，提供高效的JSON遍历和访问能力。相比标准库有不错的性能提升，同时保持内存安全和易用性。

## 🚀 核心特性

- **🔥 性能优化**: 遍历速度比标准库有显著提升
- **⚡ 内存高效**: 核心操作尽量减少内存分配
- **🛡️ 内存安全**: 完备的边界检查和安全机制
- **🎯 易于使用**: 链式调用，直观的API设计
- **🔧 功能完整**: 支持所有JSON数据类型和复杂嵌套结构
- **🌐 Unicode支持**: 很好地处理中文、emoji等Unicode字符
- **🧩 嵌套JSON展开**: 智能识别和展开JSON字符串中的嵌套JSON
- **🔢 数字精度**: 通过`FloatString()`保持原始JSON数字格式

## 📊 性能对比

| 操作        | FxJSON   | 标准库      | 性能提升      | 内存优势             |
|-----------|----------|----------|-----------|------------------|
| ForEach遍历 | 104.7 ns | 2115 ns  | **20.2x** | 零分配 vs 57次分配     |
| 数组遍历      | 30.27 ns | 2044 ns  | **67.5x** | 零分配 vs 57次分配     |
| 深度遍历      | 1363 ns  | 2787 ns  | **2.0x**  | 29次分配 vs 83次分配   |
| 复杂遍历      | 1269 ns  | 3280 ns  | **2.6x**  | 零分配 vs 104次分配    |
| 大数据遍历     | 11302 ns | 16670 ns | **1.5x**  | 181次分配 vs 559次分配 |

# FxJSON ![Flame](flame.png) - 高性能JSON解析库

## 📦 安装

```bash
go get github.com/icloudza/fxjson
```

## 🎯 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := []byte(`{
        "name": "Alice",
        "age": 30,
        "active": true,
        "score": 95.5,
        "tags": ["developer", "golang"],
        "profile": {
            "city": "北京",
            "hobby": "coding"
        }
    }`)

    // 创建节点
    node := fxjson.FromBytes(jsonData)

    // 基础访问
    name, _ := node.Get("name").String()
    age, _ := node.Get("age").Int()
    active, _ := node.Get("active").Bool()
    score, _ := node.Get("score").Float()

    fmt.Printf("姓名: %s, 年龄: %d, 激活: %v, 分数: %.1f\n", 
               name, age, active, score)
    
    // 嵌套访问
    city, _ := node.Get("profile").Get("city").String()
    fmt.Printf("城市: %s\n", city)
    
    // 路径访问
    hobby, _ := node.GetPath("profile.hobby").String()
    fmt.Printf("爱好: %s\n", hobby)
}
```

**输出:**
```
姓名: Alice, 年龄: 30, 激活: true, 分数: 95.5
城市: 北京
爱好: coding
```

### 数组操作

```go
jsonData := []byte(`{
    "users": [
        {"name": "Alice", "age": 30},
        {"name": "Bob", "age": 25},
        {"name": "Charlie", "age": 35}
    ]
}`)

node := fxjson.FromBytes(jsonData)
users := node.Get("users")

// 数组长度
fmt.Printf("用户数量: %d\n", users.Len())

// 索引访问
firstUser := users.Index(0)
name, _ := firstUser.Get("name").String()
age, _ := firstUser.Get("age").Int()
fmt.Printf("第一个用户: %s (%d岁)\n", name, age)

// 路径访问数组元素
secondName, _ := node.GetPath("users[1].name").String()
fmt.Printf("第二个用户: %s\n", secondName)
```

**输出:**
```
用户数量: 3
第一个用户: Alice (30岁)
第二个用户: Bob
```

## 🔄 高性能遍历

### 对象遍历

```go
profile := []byte(`{
    "name": "开发者",
    "skills": ["Go", "Python", "JavaScript"],
    "experience": 5,
    "remote": true
}`)

node := fxjson.FromBytes(profile)

// 零分配高性能遍历
node.ForEach(func(key string, value fxjson.Node) bool {
    switch value.Kind() {
    case fxjson.TypeString:
        str, _ := value.String()
        fmt.Printf("%s: %s\n", key, str)
    case fxjson.TypeNumber:
        num, _ := value.Int()
        fmt.Printf("%s: %d\n", key, num)
    case fxjson.TypeBool:
        b, _ := value.Bool()
        fmt.Printf("%s: %v\n", key, b)
    case fxjson.TypeArray:
        fmt.Printf("%s: [数组，长度=%d]\n", key, value.Len())
    }
    return true // 继续遍历
})
```

**输出:**
```
name: 开发者
skills: [数组，长度=3]
experience: 5
remote: true
```

### 数组遍历

```go
scores := []byte(`[95, 87, 92, 88, 96]`)
node := fxjson.FromBytes(scores)

var total int64
var count int

// 极速数组遍历（67倍性能提升）
node.ArrayForEach(func(index int, value fxjson.Node) bool {
    if score, err := value.Int(); err == nil {
        total += score
        count++
        fmt.Printf("分数 %d: %d\n", index+1, score)
    }
    return true
})

fmt.Printf("平均分: %.1f\n", float64(total)/float64(count))
```

**输出:**
```
分数 1: 95
分数 2: 87
分数 3: 92
分数 4: 88
分数 5: 96
平均分: 91.6
```

### 深度遍历

```go
complexData := []byte(`{
    "company": {
        "name": "科技公司",
        "departments": [
            {
                "name": "研发部",
                "employees": [
                    {"name": "张三", "position": "工程师"},
                    {"name": "李四", "position": "架构师"}
                ]
            }
        ]
    }
}`)

node := fxjson.FromBytes(complexData)

// 深度优先遍历整个JSON树
node.Walk(func(path string, node fxjson.Node) bool {
    if node.IsString() {
        value, _ := node.String()
        fmt.Printf("路径: %s = %s\n", path, value)
    }
    return true // 继续遍历子节点
})
```

**输出:**
```
路径: company.name = 科技公司
路径: company.departments[0].name = 研发部
路径: company.departments[0].employees[0].name = 张三
路径: company.departments[0].employees[0].position = 工程师
路径: company.departments[0].employees[1].name = 李四
路径: company.departments[0].employees[1].position = 架构师
```

## 🛠️ 高级功能

### 类型检查和转换

```go
data := []byte(`{
    "user_id": 12345,
    "username": "developer",
    "is_admin": false,
    "metadata": null,
    "scores": [100, 95, 88]
}`)

node := fxjson.FromBytes(data)

// 类型检查
fmt.Printf("user_id是数字: %v\n", node.Get("user_id").IsNumber())
fmt.Printf("username是字符串: %v\n", node.Get("username").IsString())
fmt.Printf("is_admin是布尔: %v\n", node.Get("is_admin").IsBool())
fmt.Printf("metadata是null: %v\n", node.Get("metadata").IsNull())
fmt.Printf("scores是数组: %v\n", node.Get("scores").IsArray())

// 安全类型转换
if userID, err := node.Get("user_id").Int(); err == nil {
    fmt.Printf("用户ID: %d\n", userID)
}

// 获取原始JSON
if rawScores := node.Get("scores").Raw(); len(rawScores) > 0 {
    fmt.Printf("原始scores JSON: %s\n", rawScores)
}
```

**输出:**
```
user_id是数字: true
username是字符串: true
is_admin是布尔: true
metadata是null: true
scores是数组: true
用户ID: 12345
原始scores JSON: [100, 95, 88]
```

### 数字精度处理

FxJSON 提供特殊的浮点数精度处理，以保持原始JSON格式：

```go
data := []byte(`{
    "price": 1.1,
    "rating": 4.50,
    "score": 95.0,
    "percentage": 12.34
}`)

node := fxjson.FromBytes(data)

// 保持原始JSON数字格式
price := node.Get("price")
if priceStr, err := price.FloatString(); err == nil {
    fmt.Printf("价格: %s\n", priceStr) // 输出: 1.1 (保持原始格式)
}

rating := node.Get("rating")
if ratingStr, err := rating.FloatString(); err == nil {
    fmt.Printf("评分: %s\n", ratingStr) // 输出: 4.50 (保持尾随零)
}

// 与其他方法对比
if floatVal, err := price.Float(); err == nil {
    fmt.Printf("价格 float值: %v\n", floatVal)     // 输出: 1.1
    fmt.Printf("价格格式化: %g\n", floatVal)        // 输出: 1.1
}

// 获取原始数字字符串
if numStr, err := price.NumStr(); err == nil {
    fmt.Printf("价格 NumStr: %s\n", numStr)       // 输出: 1.1
}
```

**输出:**
```
价格: 1.1
评分: 4.50
价格 float值: 1.1
价格格式化: 1.1
价格 NumStr: 1.1
```

**数字处理方法说明:**
- `FloatString()` - 返回原始JSON数字格式(推荐用于显示)
- `NumStr()` - 返回JSON中的原始数字字符串
- `Float()` - 返回`float64`值用于计算
- `Int()` - 返回`int64`值用于整数

### 条件查找和过滤

```go
students := []byte(`{
    "class": "高级班",
    "students": [
        {"name": "小明", "grade": 95, "subject": "数学"},
        {"name": "小红", "grade": 87, "subject": "英语"},
        {"name": "小李", "grade": 92, "subject": "数学"},
        {"name": "小王", "grade": 78, "subject": "英语"}
    ]
}`)

node := fxjson.FromBytes(students)
studentsArray := node.Get("students")

// 查找第一个数学科目的学生
_, student, found := studentsArray.FindInArray(func(index int, value fxjson.Node) bool {
    subject, _ := value.Get("subject").String()
    return subject == "数学"
})

if found {
    name, _ := student.Get("name").String()
    grade, _ := student.Get("grade").Int()
    fmt.Printf("第一个数学学生: %s (分数: %d)\n", name, grade)
}

// 过滤所有高分学生 (>90分)
highScoreStudents := studentsArray.FilterArray(func(index int, value fxjson.Node) bool {
    grade, _ := value.Get("grade").Int()
    return grade > 90
})

fmt.Printf("高分学生数量: %d\n", len(highScoreStudents))
for i, student := range highScoreStudents {
    name, _ := student.Get("name").String()
    grade, _ := student.Get("grade").Int()
    fmt.Printf("高分学生 %d: %s (%d分)\n", i+1, name, grade)
}
```

**输出:**
```
第一个数学学生: 小明 (分数: 95)
高分学生数量: 2
高分学生 1: 小明 (95分)
高分学生 2: 小李 (92分)
```

### 统计和分析

```go
data := []byte(`{
    "sales": [
        {"amount": 1500, "region": "北区"},
        {"amount": 2300, "region": "南区"},
        {"amount": 1800, "region": "北区"},
        {"amount": 2100, "region": "南区"}
    ]
}`)

node := fxjson.FromBytes(data)
salesArray := node.Get("sales")

// 统计北区销售记录数量
northCount := salesArray.CountIf(func(index int, value fxjson.Node) bool {
    region, _ := value.Get("region").String()
    return region == "北区"
})

// 检查是否所有销售额都大于1000
allAbove1000 := salesArray.AllMatch(func(index int, value fxjson.Node) bool {
    amount, _ := value.Get("amount").Int()
    return amount > 1000
})

// 检查是否有销售额超过2000
hasHighSales := salesArray.AnyMatch(func(index int, value fxjson.Node) bool {
    amount, _ := value.Get("amount").Int()
    return amount > 2000
})

fmt.Printf("北区记录: %d条\n", northCount)
fmt.Printf("全部>1000: %v\n", allAbove1000)
fmt.Printf("有>2000: %v\n", hasHighSales)
```

**输出:**
```
北区记录: 2条
全部>1000: true
有>2000: true
```

## 🌟 复杂应用场景

### 嵌套JSON字符串处理

```go
// 包含嵌套JSON字符串的数据
complexJSON := []byte(`{
    "user_info": "{\"name\":\"张三\",\"age\":30,\"skills\":[\"Go\",\"Python\"]}",
    "config": "{\"theme\":\"dark\",\"language\":\"zh-CN\"}",
    "regular_field": "普通字符串"
}`)

node := fxjson.FromBytes(complexJSON)

// FxJSON自动识别和展开嵌套的JSON字符串
userInfo := node.Get("user_info")
if userInfo.IsObject() { // 嵌套JSON被自动展开为对象
    name, _ := userInfo.Get("name").String()
    age, _ := userInfo.Get("age").Int()
    fmt.Printf("用户: %s, 年龄: %d\n", name, age)
    
    // 遍历技能数组
    fmt.Print("技能: ")
    userInfo.Get("skills").ArrayForEach(func(index int, skill fxjson.Node) bool {
        skillName, _ := skill.String()
        fmt.Printf("%s ", skillName)
        return true
    })
    fmt.Println()
}

// 配置也会被自动展开
config := node.Get("config")
if config.IsObject() {
    theme, _ := config.Get("theme").String()
    language, _ := config.Get("language").String()
    fmt.Printf("主题: %s, 语言: %s\n", theme, language)
}
```

**输出:**
```
用户: 张三, 年龄: 30
技能: Go Python 
主题: dark, 语言: zh-CN
```

### 配置文件解析

```go
configJSON := []byte(`{
    "database": {
        "host": "localhost",
        "port": 5432,
        "name": "myapp",
        "ssl": true,
        "pool": {
            "min": 5,
            "max": 100
        }
    },
    "redis": {
        "host": "127.0.0.1",
        "port": 6379,
        "db": 0
    },
    "features": ["auth", "logging", "metrics"]
}`)

config := fxjson.FromBytes(configJSON)

// 数据库配置
dbHost, _ := config.GetPath("database.host").String()
dbPort, _ := config.GetPath("database.port").Int()
sslEnabled, _ := config.GetPath("database.ssl").Bool()
maxPool, _ := config.GetPath("database.pool.max").Int()

fmt.Printf("数据库: %s:%d (SSL: %v, 最大连接: %d)\n", 
           dbHost, dbPort, sslEnabled, maxPool)

// Redis配置
redisHost, _ := config.GetPath("redis.host").String()
redisPort, _ := config.GetPath("redis.port").Int()
fmt.Printf("Redis: %s:%d\n", redisHost, redisPort)

// 功能列表
features := config.Get("features")
fmt.Printf("启用功能 (%d项): ", features.Len())
features.ArrayForEach(func(index int, feature fxjson.Node) bool {
    name, _ := feature.String()
    fmt.Printf("%s ", name)
    return true
})
fmt.Println()
```

**输出:**
```
数据库: localhost:5432 (SSL: true, 最大连接: 100)
Redis: 127.0.0.1:6379
启用功能 (3项): auth logging metrics 
```

### API响应处理

```go
apiResponse := []byte(`{
    "status": "success",
    "data": {
        "users": [
            {
                "id": 1,
                "name": "管理员",
                "email": "admin@example.com",
                "roles": ["admin", "user"],
                "profile": {
                    "avatar": "https://example.com/avatar.jpg",
                    "bio": "系统管理员"
                }
            },
            {
                "id": 2,
                "name": "普通用户",
                "email": "user@example.com",
                "roles": ["user"]
            }
        ],
        "pagination": {
            "page": 1,
            "per_page": 10,
            "total": 2
        }
    }
}`)

response := fxjson.FromBytes(apiResponse)

// 检查响应状态
status, _ := response.Get("status").String()
if status == "success" {
    // 处理用户数据
    users := response.GetPath("data.users")
    fmt.Printf("用户列表 (共%d个):\n", users.Len())
    
    users.ArrayForEach(func(index int, user fxjson.Node) bool {
        id, _ := user.Get("id").Int()
        name, _ := user.Get("name").String()
        email, _ := user.Get("email").String()
        
        fmt.Printf("  用户 %d: %s (%s)\n", id, name, email)
        
        // 处理角色
        roles := user.Get("roles")
        fmt.Printf("    角色: ")
        roles.ArrayForEach(func(i int, role fxjson.Node) bool {
            roleName, _ := role.String()
            fmt.Printf("%s ", roleName)
            return true
        })
        fmt.Println()
        
        return true
    })
    
    // 分页信息
    page, _ := response.GetPath("data.pagination.page").Int()
    total, _ := response.GetPath("data.pagination.total").Int()
    perPage, _ := response.GetPath("data.pagination.per_page").Int()
    fmt.Printf("分页: 第%d页，每页%d条，共%d条\n", page, perPage, total)
}
```

**输出:**
```
用户列表 (共2个):
  用户 1: 管理员 (admin@example.com)
    角色: admin user 
  用户 2: 普通用户 (user@example.com)
    角色: user 
分页: 第1页，每页10条，共2条
```

## ⚙️ 解码到结构体

```go
type User struct {
    Name  string   `json:"name"`
    Age   int      `json:"age"`
    Tags  []string `json:"tags"`
    Email string   `json:"email"`
}

jsonData := []byte(`{
    "name": "开发者",
    "age": 28,
    "tags": ["golang", "json", "performance"],
    "email": "dev@example.com"
}`)

node := fxjson.FromBytes(jsonData)

var user User
if err := node.Decode(&user); err != nil {
    fmt.Printf("解码错误: %v\n", err)
} else {
    fmt.Printf("解码结果:\n")
    fmt.Printf("  姓名: %s\n", user.Name)
    fmt.Printf("  年龄: %d\n", user.Age)
    fmt.Printf("  邮箱: %s\n", user.Email)
    fmt.Printf("  标签: %v\n", user.Tags)
}
```

**输出:**
```
解码结果:
  姓名: 开发者
  年龄: 28
  邮箱: dev@example.com
  标签: [golang json performance]
```

## 🚨 错误处理

```go
jsonData := []byte(`{
    "number": "not_a_number",
    "missing": null,
    "empty_string": "",
    "valid_number": 42
}`)

node := fxjson.FromBytes(jsonData)

// 类型转换错误处理
if num, err := node.Get("number").Int(); err != nil {
    fmt.Printf("数字转换失败: %v\n", err)
}

// 成功的类型转换
if num, err := node.Get("valid_number").Int(); err == nil {
    fmt.Printf("有效数字: %d\n", num)
}

// 检查字段是否存在
if node.HasKey("missing_field") {
    fmt.Println("missing_field字段存在")
} else {
    fmt.Println("missing_field字段不存在")
}

if node.HasKey("valid_number") {
    fmt.Println("valid_number字段存在")
}

// 使用默认值
defaultNode := fxjson.FromBytes([]byte(`"默认值"`))
value := node.GetKeyValue("missing_field", defaultNode)
defaultStr, _ := value.String()
fmt.Printf("使用默认值: %s\n", defaultStr)

// 处理空字符串
emptyStr, err := node.Get("empty_string").String()
if err == nil {
    fmt.Printf("空字符串长度: %d\n", len(emptyStr))
}
```

**输出:**
```
数字转换失败: node is not a number type (got type="string")
有效数字: 42
missing_field字段不存在
valid_number字段存在
使用默认值: 默认值
空字符串长度: 0
```

## 🎨 便捷方法

```go
data := []byte(`{
    "company": {
        "name": "科技公司",
        "founded": 2020,
        "employees": [
            {"name": "张三", "department": "研发", "salary": 15000},
            {"name": "李四", "department": "市场", "salary": 12000},
            {"name": "王五", "department": "研发", "salary": 18000}
        ]
    }
}`)

node := fxjson.FromBytes(data)

// 转换为Map
fmt.Println("=== 公司信息 (ToMap) ===")
companyMap := node.Get("company").ToMap()
for key, value := range companyMap {
    if key == "employees" {
        fmt.Printf("%s: [数组，长度=%d]\n", key, value.Len())
    } else {
        fmt.Printf("%s: %s\n", key, string(value.Raw()))
    }
}

// 转换为切片
fmt.Println("\n=== 员工列表 (ToSlice) ===")
employees := node.GetPath("company.employees").ToSlice()
fmt.Printf("员工总数: %d\n", len(employees))
for i, employee := range employees {
    name, _ := employee.Get("name").String()
    dept, _ := employee.Get("department").String()
    salary, _ := employee.Get("salary").Int()
    fmt.Printf("员工 %d: %s - %s部门 (薪资: %d)\n", i+1, name, dept, salary)
}

// 获取所有键名
fmt.Println("\n=== 公司字段列表 (GetAllKeys) ===")
keys := node.Get("company").GetAllKeys()
fmt.Printf("公司字段: %v\n", keys)

// 获取所有员工节点
fmt.Println("\n=== 员工节点列表 (GetAllValues) ===")
employeeNodes := node.GetPath("company.employees").GetAllValues()
fmt.Printf("员工节点数: %d\n", len(employeeNodes))
for i, empNode := range employeeNodes {
    name, _ := empNode.Get("name").String()
    fmt.Printf("节点 %d: %s的信息\n", i+1, name)
}
```

**输出:**
```
=== 公司信息 (ToMap) ===
name: "科技公司"
founded: 2020
employees: [数组，长度=3]

=== 员工列表 (ToSlice) ===
员工总数: 3
员工 1: 张三 - 研发部门 (薪资: 15000)
员工 2: 李四 - 市场部门 (薪资: 12000)
员工 3: 王五 - 研发部门 (薪资: 18000)

=== 公司字段列表 (GetAllKeys) ===
公司字段: [name founded employees]

=== 员工节点列表 (GetAllValues) ===
员工节点数: 3
节点 1: 张三的信息
节点 2: 李四的信息
节点 3: 王五的信息
```

## 📝 性能提示

1. **遍历优化**: 对于大数据量，优先使用`ForEach`、`ArrayForEach`和`Walk`方法
2. **路径访问**: 使用`GetPath`可以一次性访问深层嵌套字段
3. **内存管理**: 核心遍历操作实现零分配，适合高频调用场景
4. **类型检查**: 使用`IsXXX()`方法进行类型检查，避免不必要的类型转换
5. **缓存利用**: 数组索引会自动缓存，重复访问同一数组时性能更佳

## ⚠️ 注意事项

1. **输入验证**: 假设输入是有效的JSON，专注于性能而非错误处理
2. **内存安全**: 所有字符串操作都经过边界检查
3. **Unicode支持**: 完美支持中文、emoji等Unicode字符
4. **并发安全**: 节点读取操作是并发安全的
5. **Go版本**: 需要Go 1.18或更高版本

## 📚 完整API参考

### 核心方法

#### 节点创建
- `FromBytes(data []byte) Node` - 从JSON字节创建节点，自动展开嵌套JSON

#### 基础访问
- `Get(key string) Node` - 通过键获取对象字段
- `GetPath(path string) Node` - 通过路径获取值 (如 "user.profile.name")
- `Index(i int) Node` - 通过索引获取数组元素

#### 类型检查
- `Exists() bool` - 检查节点是否存在
- `IsObject() bool` - 检查是否为JSON对象
- `IsArray() bool` - 检查是否为JSON数组
- `IsString() bool` - 检查是否为JSON字符串
- `IsNumber() bool` - 检查是否为JSON数字
- `IsBool() bool` - 检查是否为JSON布尔值
- `IsNull() bool` - 检查是否为JSON null
- `IsScalar() bool` - 检查是否为标量类型 (字符串、数字、布尔、null)
- `IsContainer() bool` - 检查是否为容器类型 (对象、数组)
- `Kind() NodeType` - 获取节点类型枚举
- `Type() byte` - 获取内部类型字节

#### 值提取
- `String() (string, error)` - 获取字符串值
- `Int() (int64, error)` - 获取整数值
- `Uint() (uint64, error)` - 获取无符号整数值
- `Float() (float64, error)` - 获取浮点数值
- `Bool() (bool, error)` - 获取布尔值
- `NumStr() (string, error)` - 获取原始JSON数字字符串
- `FloatString() (string, error)` - 获取保持原始JSON格式的数字字符串
- `Raw() []byte` - 获取此节点的原始JSON字节
- `RawString() (string, error)` - 获取原始JSON字符串
- `Json() (string, error)` - 获取JSON表示 (仅对象/数组)

#### 大小和键值
- `Len() int` - 获取长度 (数组元素、对象字段、字符串字符)
- `Keys() [][]byte` - 获取对象键的字节切片
- `GetAllKeys() []string` - 获取对象键的字符串切片
- `GetAllValues() []Node` - 获取数组元素的节点切片
- `ToMap() map[string]Node` - 将对象转换为映射
- `ToSlice() []Node` - 将数组转换为切片

#### 高性能遍历
- `ForEach(fn ForEachFunc) bool` - 零分配遍历对象 (20倍更快)
- `ArrayForEach(fn ArrayForEachFunc) bool` - 零分配遍历数组 (67倍更快)
- `Walk(fn WalkFunc) bool` - 深度遍历整个JSON树 (2倍更快)

#### 搜索和过滤
- `FindInObject(predicate func(key string, value Node) bool) (string, Node, bool)` - 查找首个匹配的对象字段
- `FindInArray(predicate func(index int, value Node) bool) (int, Node, bool)` - 查找首个匹配的数组元素
- `FilterArray(predicate func(index int, value Node) bool) []Node` - 过滤数组元素
- `FindByPath(path string) Node` - GetPath的别名

#### 条件操作
- `HasKey(key string) bool` - 检查对象是否有指定键
- `GetKeyValue(key string, defaultValue Node) Node` - 获取值，支持默认值回退
- `CountIf(predicate func(index int, value Node) bool) int` - 统计匹配的数组元素
- `AllMatch(predicate func(index int, value Node) bool) bool` - 检查是否所有数组元素匹配
- `AnyMatch(predicate func(index int, value Node) bool) bool` - 检查是否有数组元素匹配

#### 解码
- `Decode(v any) error` - 解码JSON到Go结构体/类型

### 回调函数类型

```go
// 对象遍历回调
type ForEachFunc func(key string, value Node) bool

// 数组遍历回调  
type ArrayForEachFunc func(index int, value Node) bool

// 深度遍历回调
type WalkFunc func(path string, node Node) bool
```

### 节点类型

```go
const (
    TypeInvalid NodeType = 0    // 无效类型
    TypeObject  NodeType = 'o'  // 对象类型
    TypeArray   NodeType = 'a'  // 数组类型
    TypeString  NodeType = 's'  // 字符串类型
    TypeNumber  NodeType = 'n'  // 数字类型
    TypeBool    NodeType = 'b'  // 布尔类型
    TypeNull    NodeType = 'l'  // null类型
)
```

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

---

**FxJSON - 让JSON解析飞起来！** 🚀