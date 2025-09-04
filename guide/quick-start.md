# 5分钟快速上手

欢迎使用 FxJSON！本指南将帮助您快速掌握 FxJSON 的核心功能，从安装到实际应用，让您在 5 分钟内开始使用这个高性能的 JSON 解析库。

## 第一步：安装 FxJSON

### 环境要求
- Go 1.18 或更高版本
- 支持 Linux、macOS、Windows 等主流操作系统

### 安装命令
```bash
go get github.com/icloudza/fxjson
```

### 验证安装
创建一个简单的测试文件验证安装：

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    node := fxjson.FromString(`{"message": "安装成功"}`)
    fmt.Println(node.Get("message").StringOr("安装失败"))
}
```

运行 `go run main.go`，看到输出 "安装成功" 表示安装完成。

## 第二步：核心概念理解

### 什么是 Node？
在 FxJSON 中，`Node` 是核心概念，代表 JSON 中的任意值（对象、数组、字符串、数字等）。您可以把它想象成一个智能指针，指向 JSON 数据的某个部分。

### 三种解析方式
```go
// 1. 从字符串解析（最常用）
node := fxjson.FromString(`{"name": "张三"}`)

// 2. 从字节数组解析（高性能）
data := []byte(`{"name": "张三"}`)
node := fxjson.FromBytes(data)

// 3. 带选项解析（自定义配置）
node := fxjson.FromStringWithOptions(`{"name": "张三"}`, fxjson.DefaultParseOptions)
```

## 第三步：数据访问

### 两种访问方式

#### 1. 链式访问（推荐新手）
```go
jsonData := `{
    "user": {
        "name": "张三",
        "age": 30,
        "address": {
            "city": "北京"
        }
    }
}`

node := fxjson.FromString(jsonData)

// 逐级访问
name := node.Get("user").Get("name").StringOr("未知")
age := node.Get("user").Get("age").IntOr(0)
city := node.Get("user").Get("address").Get("city").StringOr("")

fmt.Printf("姓名: %s, 年龄: %d, 城市: %s\n", name, age, city)
```

#### 2. 路径访问（更简洁）
```go
// 使用点号分隔路径
name := node.GetPath("user.name").StringOr("未知")
age := node.GetPath("user.age").IntOr(0) 
city := node.GetPath("user.address.city").StringOr("")

fmt.Printf("姓名: %s, 年龄: %d, 城市: %s\n", name, age, city)
```

### 为什么使用 `StringOr()` ？

**传统方式**（需要错误处理）：
```go
name, err := node.Get("name").String()
if err != nil {
    name = "默认值"  // 手动处理错误
}
```

**FxJSON 方式**（自动处理）：
```go
name := node.Get("name").StringOr("默认值")  // 一行搞定
```

`StringOr()` 系列方法会自动处理以下情况：
- 字段不存在
- 类型不匹配  
- JSON 格式错误

这样您就不需要写繁琐的错误处理代码了！

## 第四步：处理数组数据

数组是 JSON 中的常见结构，FxJSON 提供了高性能的数组处理方式。

### 数组访问
```go
jsonData := `{
    "users": [
        {"name": "张三", "age": 25},
        {"name": "李四", "age": 30}, 
        {"name": "王五", "age": 35}
    ]
}`

node := fxjson.FromString(jsonData)
users := node.Get("users")

// 获取数组长度
count := users.Len()
fmt.Printf("共有 %d 个用户\n", count)

// 访问特定索引
firstUser := users.Index(0)
name := firstUser.Get("name").StringOr("")
fmt.Printf("第一个用户: %s\n", name)
```

### 高性能数组遍历
```go
// ArrayForEach：零分配遍历，比标准库快 67 倍
users.ArrayForEach(func(index int, user fxjson.Node) bool {
    name := user.Get("name").StringOr("")
    age := user.Get("age").IntOr(0)
    fmt.Printf("用户 %d: %s (年龄 %d)\n", index+1, name, age)
    return true  // 返回 true 继续遍历，false 停止
})
```

## 第五步：处理对象数据

### 对象遍历
```go
jsonData := `{
    "departments": {
        "tech": {"count": 25, "budget": 500000},
        "marketing": {"count": 15, "budget": 300000},
        "sales": {"count": 20, "budget": 400000}
    }
}`

node := fxjson.FromString(jsonData)
departments := node.Get("departments")

// ForEach：遍历对象的所有键值对，比标准库快 20 倍
departments.ForEach(func(deptName string, deptInfo fxjson.Node) bool {
    count := deptInfo.Get("count").IntOr(0)
    budget := deptInfo.Get("budget").IntOr(0)
    fmt.Printf("%s 部门: %d 人，预算 %d 元\n", deptName, count, budget)
    return true
})
```

## 第六步：常用数据类型

FxJSON 支持所有 JSON 数据类型，并提供安全的转换方法：

```go
jsonData := `{
    "name": "张三",          // 字符串
    "age": 30,              // 整数
    "height": 175.5,        // 浮点数
    "married": true,        // 布尔值
    "address": null,        // null 值
    "hobbies": ["读书", "游泳"] // 数组
}`

node := fxjson.FromString(jsonData)

// 各种类型的安全转换
name := node.Get("name").StringOr("默认名字")          // 字符串，默认值
age := node.Get("age").IntOr(0)                      // 整数，默认值 0
height := node.Get("height").FloatOr(0.0)           // 浮点数，默认值 0.0
married := node.Get("married").BoolOr(false)        // 布尔值，默认值 false

// 检查字段状态
fmt.Printf("name 字段存在: %t\n", node.Get("name").Exists())
fmt.Printf("address 是否为 null: %t\n", node.Get("address").IsNull())
fmt.Printf("age 是否为数字: %t\n", node.Get("age").IsNumber())
fmt.Printf("hobbies 数组长度: %d\n", node.Get("hobbies").Len())
```

## 第七步：完整实战示例

现在让我们结合前面学到的知识，处理一个真实的 API 响应：

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    // 模拟一个用户管理系统的 API 响应
    apiResponse := `{
        "status": "success",
        "data": {
            "total": 3,
            "users": [
                {"id": 1, "name": "张三", "role": "admin", "active": true},
                {"id": 2, "name": "李四", "role": "user", "active": false},
                {"id": 3, "name": "王五", "role": "user", "active": true}
            ]
        }
    }`
    
    node := fxjson.FromString(apiResponse)
    
    // 1. 检查 API 调用状态
    if node.Get("status").StringOr("") != "success" {
        fmt.Println("API 调用失败")
        return
    }
    
    // 2. 获取用户总数
    data := node.Get("data")
    total := data.Get("total").IntOr(0)
    fmt.Printf("用户总数: %d\n", total)
    
    // 3. 遍历用户列表
    users := data.Get("users")
    fmt.Println("\n用户详情:")
    users.ArrayForEach(func(index int, user fxjson.Node) bool {
        id := user.Get("id").IntOr(0)
        name := user.Get("name").StringOr("未知")
        role := user.Get("role").StringOr("guest")
        active := user.Get("active").BoolOr(false)
        
        status := "离线"
        if active {
            status = "在线"
        }
        
        fmt.Printf("  %d. [%s] %s (%s) - %s\n", id, role, name, status, "正常")
        return true
    })
    
    // 4. 统计活跃用户
    activeCount := 0
    users.ArrayForEach(func(i int, user fxjson.Node) bool {
        if user.Get("active").BoolOr(false) {
            activeCount++
        }
        return true
    })
    
    fmt.Printf("\n活跃用户: %d/%d\n", activeCount, total)
}
```

## 学习总结

恭喜！您已经掌握了 FxJSON 的核心用法：

### 核心概念回顾
- **Node**：代表 JSON 中的任意值，是操作的核心
- **链式访问**：`node.Get("key").Get("subkey")`
- **路径访问**：`node.GetPath("key.subkey")`
- **安全方法**：`StringOr()` `IntOr()` `FloatOr()` `BoolOr()`

### 主要优势
- **性能**: 比标准库快 20-67 倍
- **安全**: 自动错误处理，无需繁琐的 if err != nil
- **简洁**: API 设计直观，代码更清晰
- **完整**: 支持所有 JSON 操作需求

### 使用建议
1. **优先使用 `Or` 系列方法**，除非你确实需要错误信息
2. **数组遍历优先使用 `ArrayForEach`**，性能最优
3. **嵌套访问优先使用路径语法**，代码更简洁
4. **重复访问会自动缓存**，无需手动优化

## 下一步学习

现在您已经掌握了基础用法，可以继续探索：

- **[完整 API 文档](/api/)** - 查看所有可用方法和高级功能
- **[实用示例](/examples/)** - 学习更多实际应用场景
- **[性能详解](/performance/)** - 了解性能优化细节

### 常见问题速查
- **字段不存在怎么办？** 使用 `StringOr("默认值")` 自动处理
- **数组怎么遍历？** 使用 `ArrayForEach(func(i int, item Node) bool {...})`
- **深层嵌套怎么访问？** 使用 `GetPath("a.b.c.d")`
- **怎么检查类型？** 使用 `IsString()` `IsNumber()` `IsArray()` 等方法

开始在您的项目中使用 FxJSON 吧！