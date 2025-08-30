# 快速开始

FxJSON 是一个专注性能的 Go JSON 解析库，提供高效的 JSON 遍历和访问能力。

## 安装

使用 `go get` 命令安装 FxJSON：

```bash
go get github.com/icloudza/fxjson
```

## 基础用法

### 解析 JSON

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := []byte(`{
        "name": "张三",
        "age": 30,
        "profile": {
            "city": "北京",
            "hobby": "编程"
        }
    }`)

    // 从字节数组解析
    node := fxjson.FromBytes(jsonData)

}
```

### 访问数据

```go
// 安全访问，使用默认值
name := node.Get("name").StringOr("未知")
age := node.Get("age").IntOr(0)
city := node.GetPath("profile.city").StringOr("")

fmt.Printf("姓名: %s, 年龄: %d, 城市: %s\n", name, age, city)
// 输出: 姓名: 张三, 年龄: 30, 城市: 北京
```

### 遍历数组

```go
jsonData := []byte(`[
    {"name": "张三", "age": 25},
    {"name": "李四", "age": 30},
    {"name": "王五", "age": 35}
]`)

node := fxjson.FromBytes(jsonData)

// 高性能零分配遍历 - 比标准库快67倍
node.ArrayForEach(func(index int, user fxjson.Node) bool {
    name := user.Get("name").StringOr("")
    age := user.Get("age").IntOr(0)
    fmt.Printf("用户 %d: %s, 年龄 %d\n", index+1, name, age)
    return true // 继续遍历
})
```

### 遍历对象

```go
jsonData := []byte(`{
    "user1": {"name": "张三", "active": true},
    "user2": {"name": "李四", "active": false},
    "user3": {"name": "王五", "active": true}
}`)

node := fxjson.FromBytes(jsonData)

// 遍历对象键值对 - 比标准库快20倍
node.ForEach(func(key string, value fxjson.Node) bool {
    name := value.Get("name").StringOr("")
    active := value.Get("active").BoolOr(false)
    fmt.Printf("%s: %s (活跃: %t)\n", key, name, active)
    return true // 继续遍历
})
```

## 数据类型转换

FxJSON 提供安全的类型转换方法，支持默认值：

```go
node := fxjson.FromBytes(`{
    "name": "张三",
    "age": 30,
    "score": 95.5,
    "active": true,
    "tags": ["golang", "json", "performance"]
}`)

// 字符串类型
name := node.Get("name").StringOr("默认姓名")

// 数值类型
age := node.Get("age").IntOr(0)
score := node.Get("score").FloatOr(0.0)

// 布尔类型
active := node.Get("active").BoolOr(false)

// 数组长度
tagsCount := node.Get("tags").Len()

// 检查字段是否存在
if node.Get("name").Exists() {
    fmt.Println("name 字段存在")
}

// 检查数据类型
if node.Get("age").IsNumber() {
    fmt.Println("age 是数字类型")
}
```

## 错误处理

FxJSON 提供两种访问模式：

### 1. 安全模式（推荐）

使用 `Or` 系列方法，自动处理错误并提供默认值：

```go
// 即使字段不存在或类型不匹配，也会返回默认值
name := node.Get("non_exist").StringOr("默认值")
age := node.Get("invalid_number").IntOr(0)
```

### 2. 严格模式

使用无后缀方法，需要手动检查错误：

```go
name, err := node.Get("name").String()
if err != nil {
    // 处理错误
    fmt.Printf("获取 name 失败: %v\n", err)
    return
}
```

## 性能特点

FxJSON 的主要性能优势：

- **零分配遍历**: 数组和对象遍历不产生内存分配
- **缓存机制**: 重复访问同一数据时使用缓存加速
- **路径优化**: 支持 `profile.city` 格式的路径访问
- **类型检查优化**: 快速的类型判断和转换

```go
// 缓存示例 - 第二次访问会更快
node := fxjson.FromString(`{"users": [{"name": "test"}]}`)

// 第一次访问，会建立缓存
user := node.GetPath("users.0.name").StringOr("")

// 第二次访问，使用缓存，速度提升约4倍
user2 := node.GetPath("users.0.name").StringOr("")
```

## 下一步

- 了解 [API 文档](/api/) 获取完整的方法列表
- 查看 [示例和用法](/examples/) 了解更多使用场景
- 阅读 [性能对比](/performance/) 了解性能优势