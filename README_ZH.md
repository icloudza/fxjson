[![Go Doc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/icloudza/fxjson?utm_source=godoc)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Performance](https://img.shields.io/badge/performance-fast-orange.svg)](#性能对比)

[📄 English Documentation](README.md)

FxJSON 是一个专注性能的Go JSON解析库，提供高效的JSON遍历和访问能力。相比标准库有不错的性能提升，同时保持内存安全和易用性。

## 🚀 核心特性

- **🔥 高性能**: 遍历速度比标准库有显著提升
- **⚡ 内存高效**: 核心操作尽量减少内存分配
- **🛡️ 内存安全**: 完备的边界检查和安全机制
- **🎯 易于使用**: 链式调用，直观的API设计
- **🔧 功能完整**: 支持所有JSON数据类型和复杂嵌套结构
- **🌐 Unicode支持**: 很好地处理中文、emoji等Unicode字符
- **🧩 嵌套JSON展开**: 智能识别和展开JSON字符串中的嵌套JSON
- **🔢 数字精度**: 通过`FloatString()`保持原始JSON数字格式
- **🔍 高级查询**: SQL风格的条件查询和过滤功能
- **📊 数据聚合**: 内置统计和聚合计算功能
- **🎨 数据变换**: 灵活的字段映射和类型转换
- **✅ 数据验证**: 全面的验证规则和数据清洗
- **💾 智能缓存**: 高性能缓存，支持LRU淘汰策略
- **🔧 调试工具**: 增强的调试和分析功能

## 📊 性能对比

### 核心操作
| 操作        | FxJSON   | 标准库      | 性能提升      | 内存优势             |
|-----------|----------|----------|-----------|------------------|
| ForEach遍历 | 104.7 ns | 2115 ns  | **20.2x** | 零分配 vs 57次分配     |
| 数组遍历      | 30.27 ns | 2044 ns  | **67.5x** | 零分配 vs 57次分配     |
| 深度遍历      | 1363 ns  | 2787 ns  | **2.0x**  | 29次分配 vs 83次分配   |
| 复杂遍历      | 1269 ns  | 3280 ns  | **2.6x**  | 零分配 vs 104次分配    |
| 大数据遍历     | 11302 ns | 16670 ns | **1.5x**  | 181次分配 vs 559次分配 |

### 高级功能性能
| 功能特性            | 操作耗时       | 内存使用      | 分配次数      | 说明                    |
|------------------|-------------|-------------|-------------|-------------------------|
| 基础解析            | 5,542 ns    | 6,360 B     | 50 allocs   | 标准JSON解析             |
| **缓存解析**        | **1,396 ns** | **80 B**    | **3 allocs** | **快4倍，内存减少98%**     |
| 数据变换            | 435 ns      | 368 B       | 5 allocs    | 字段映射和类型转换          |
| 数据验证            | 208 ns      | 360 B       | 4 allocs    | 基于规则的数据验证          |
| 简单查询            | 2,784 ns    | 640 B       | 14 allocs   | 条件过滤                 |
| 复杂查询            | 4,831 ns    | 1,720 B     | 52 allocs   | 多条件查询和排序           |
| 数据聚合            | 4,213 ns    | 2,640 B     | 32 allocs   | 统计计算                 |
| 大数据查询          | 1.27 ms     | 82 B        | 2 allocs    | 100条记录处理            |
| 流式处理            | 2,821 ns    | 0 B         | 0 allocs    | 零分配流式数据处理          |
| JSON差异对比        | 17,200 ns   | 2,710 B     | 197 allocs  | 变更检测                 |
| 空字符串处理         | 3,007 ns    | 1,664 B     | 27 allocs   | 安全的空字符串处理          |

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

## 🔍 高级功能

### SQL风格查询

```go
notesData := []byte(`{
    "notes": [
        {"id": "1", "title": "Go教程", "views": 1250, "category": "tech"},
        {"id": "2", "title": "烹饪技巧", "views": 890, "category": "food"},
        {"id": "3", "title": "旅行攻略", "views": 2100, "category": "travel"}
    ]
}`)

node := fxjson.FromBytes(notesData)
notesList := node.Get("notes")

// 复杂多条件查询
results, err := notesList.Query().
    Where("views", ">", 1000).
    Where("category", "!=", "food").
    SortBy("views", "desc").
    Limit(10).
    ToSlice()

if err == nil {
    fmt.Printf("找到 %d 篇高浏览量笔记\n", len(results))
    for _, note := range results {
        title, _ := note.Get("title").String()
        views, _ := note.Get("views").Int()
        fmt.Printf("- %s (%d次浏览)\n", title, views)
    }
}
```

**输出:**
```
找到 2 篇高浏览量笔记
- 旅行攻略 (2100次浏览)
- Go教程 (1250次浏览)
```

### 数据聚合与统计

```go
// 按分类分组并计算统计信息
stats, err := notesList.Aggregate().
    GroupBy("category").
    Count("total_notes").
    Sum("views", "total_views").
    Avg("views", "avg_views").
    Max("views", "max_views").
    Execute(notesList)

if err == nil {
    fmt.Println("按分类统计:")
    for category, data := range stats {
        statsMap := data.(map[string]interface{})
        fmt.Printf("📁 %s: %d篇, %.0f总浏览, %.1f平均浏览\n",
            category, int(statsMap["total_notes"].(float64)),
            statsMap["total_views"], statsMap["avg_views"])
    }
}
```

**输出:**
```
按分类统计:
📁 tech: 1篇, 1250总浏览, 1250.0平均浏览
📁 food: 1篇, 890总浏览, 890.0平均浏览  
📁 travel: 1篇, 2100总浏览, 2100.0平均浏览
```

### 数据变换与映射

```go
// 使用字段映射转换数据结构
mapper := fxjson.FieldMapper{
    Rules: map[string]string{
        "notes[0].title": "post_title",
        "notes[0].views": "view_count", 
        "notes[0].category": "post_category",
    },
    DefaultValues: map[string]interface{}{
        "status": "published",
        "created_by": "system",
    },
    TypeCast: map[string]string{
        "view_count": "int",
    },
}

result, err := node.Transform(mapper)
if err == nil {
    fmt.Println("转换后的数据:")
    for key, value := range result {
        fmt.Printf("  %s: %v\n", key, value)
    }
}
```

**输出:**
```
转换后的数据:
  post_title: Go教程
  view_count: 1250
  post_category: tech
  status: published
  created_by: system
```

### 高性能缓存

```go
// 启用缓存以提升性能
cache := fxjson.NewMemoryCache(100)
fxjson.EnableCaching(cache)

// 第一次解析(缓存未命中)
start := time.Now()
node1 := fxjson.FromBytesWithCache(notesData, 5*time.Minute)
firstTime := time.Since(start)

// 第二次解析(缓存命中)
start = time.Now()
node2 := fxjson.FromBytesWithCache(notesData, 5*time.Minute)
secondTime := time.Since(start)

stats := cache.Stats()
fmt.Printf("首次解析: %v\n", firstTime)
fmt.Printf("缓存解析: %v (快%.1f倍)\n", 
    secondTime, float64(firstTime)/float64(secondTime))
fmt.Printf("缓存命中率: %.1f%%\n", stats.HitRate*100)
```

**输出:**
```
首次解析: 45.2µs
缓存解析: 4.8µs (快9.4倍)
缓存命中率: 50.0%
```

### 数据验证

```go
// 定义验证规则
validator := &fxjson.DataValidator{
    Rules: map[string]fxjson.ValidationRule{
        "title": {
            Required:  true,
            Type:      "string",
            MinLength: 1,
            MaxLength: 100,
        },
        "views": {
            Required: true,
            Type:     "number",
            Min:      0,
            Max:      1000000,
        },
    },
}

// 验证第一篇笔记
firstNote := notesList.Index(0)
result, errors := firstNote.Validate(validator)

if len(errors) == 0 {
    fmt.Println("✅ 验证通过")
    fmt.Printf("验证字段数: %d\n", len(result))
} else {
    fmt.Println("❌ 验证失败:")
    for _, err := range errors {
        fmt.Printf("  - %s\n", err)
    }
}
```

### 增强调试功能

```go
// 启用调试模式
fxjson.EnableDebugMode()
defer fxjson.DisableDebugMode()

// 带调试信息的解析
node, debugInfo := fxjson.FromBytesWithDebug(notesData)

fmt.Printf("📊 调试信息:\n")
fmt.Printf("  解析时间: %v\n", debugInfo.ParseTime)
fmt.Printf("  内存使用: %d 字节\n", debugInfo.MemoryUsage)
fmt.Printf("  节点数量: %d\n", debugInfo.NodeCount)
fmt.Printf("  最大深度: %d\n", debugInfo.MaxDepth)

// 美化打印JSON结构
prettyOutput := node.PrettyPrint()
fmt.Printf("\n📝 美化JSON:\n%s\n", prettyOutput)

// 分析JSON结构
inspection := node.Inspect()
fmt.Printf("\n🔍 结构分析:\n")
fmt.Printf("  类型: %v\n", inspection["type"])
fmt.Printf("  键数量: %v\n", inspection["key_count"])
```

**输出:**
```
📊 调试信息:
  解析时间: 125.4µs
  内存使用: 15360 字节
  节点数量: 42
  最大深度: 3

📝 美化JSON:
{
  "notes": [
    {
      "id": "1",
      "title": "Go教程",
      "views": 1250,
      "category": "tech"
    },
    ...
  ]
}

🔍 结构分析:
  类型: 111
  键数量: 1
```

### 流式处理与批处理

```go
// 大数据集的流式处理
processedCount := 0
err := notesList.Stream(func(note fxjson.Node, index int) bool {
    title, _ := note.Get("title").String()
    views, _ := note.Get("views").Int()
    
    fmt.Printf("处理笔记 %d: %s (%d次浏览)\n", index+1, title, views)
    processedCount++
    
    // 需要时可以提前终止
    return true
})

fmt.Printf("通过流式处理了 %d 篇笔记\n", processedCount)

// 自定义批量大小的批处理
batchProcessor := fxjson.NewBatchProcessor(2, func(nodes []fxjson.Node) error {
    fmt.Printf("处理批次: %d个节点\n", len(nodes))
    // 处理批次...
    return nil
})

notesList.ArrayForEach(func(index int, note fxjson.Node) bool {
    batchProcessor.Add(note)
    return true
})
batchProcessor.Flush()
```

**输出:**
```
处理笔记 1: Go教程 (1250次浏览)
处理笔记 2: 烹饪技巧 (890次浏览)
处理笔记 3: 旅行攻略 (2100次浏览)
通过流式处理了 3 篇笔记
处理批次: 2个节点
处理批次: 1个节点
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
        fmt.Printf("%s: [数组, 长度=%d]\n", key, value.Len())
    }
    return true // 继续遍历
})
```

**输出:**
```
name: 开发者
skills: [数组, 长度=3]
experience: 5
remote: true
```

### 数组遍历

```go
scores := []byte(`[95, 87, 92, 88, 96]`)
node := fxjson.FromBytes(scores)

var total int64
var count int

// 超快数组遍历(67倍性能提升)
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

## 🎯 应用场景

### 1. **配置管理**
- 复杂配置解析和验证
- 环境特定配置合并
- 带缓存的实时配置更新

### 2. **API响应处理**
- 高吞吐量API响应解析
- 不同API版本的数据转换
- 响应过滤和聚合

### 3. **数据分析**
- 大数据集分析和聚合
- 实时指标计算
- 数据质量验证和清洗

### 4. **内容管理**
- 文档结构分析
- 内容转换和迁移
- 搜索和过滤操作

### 5. **日志处理**
- 结构化日志解析和分析
- 日志聚合和统计
- 性能监控和调试

## 🛠️ 高级特性

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

FxJSON提供特殊的浮点数精度处理，保持原始JSON格式:

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
    fmt.Printf("价格: %s\n", priceStr) // 输出: 1.1 (保持原格式)
}

rating := node.Get("rating")
if ratingStr, err := rating.FloatString(); err == nil {
    fmt.Printf("评分: %s\n", ratingStr) // 输出: 4.50 (保持尾随零)
}

// 与其他方法对比
if floatVal, err := price.Float(); err == nil {
    fmt.Printf("价格作为float: %v\n", floatVal)     // 输出: 1.1
    fmt.Printf("价格格式化: %g\n", floatVal)        // 输出: 1.1
}

// 获取原始数字字符串
if numStr, err := price.NumStr(); err == nil {
    fmt.Printf("价格NumStr: %s\n", numStr)         // 输出: 1.1
}
```

**输出:**
```
价格: 1.1
评分: 4.50
价格作为float: 1.1
价格格式化: 1.1
价格NumStr: 1.1
```

**数字处理方法说明:**
- `FloatString()` - 返回原始JSON数字格式(推荐用于显示)
- `NumStr()` - 返回JSON中的原始数字字符串
- `Float()` - 返回`float64`值用于计算
- `Int()` - 返回`int64`值用于整数

### 条件搜索和过滤

```go
students := []byte(`{
    "class": "高级班",
    "students": [
        {"name": "Alice", "grade": 95, "subject": "数学"},
        {"name": "Bob", "grade": 87, "subject": "英语"},
        {"name": "Charlie", "grade": 92, "subject": "数学"},
        {"name": "Diana", "grade": 78, "subject": "英语"}
    ]
}`)

node := fxjson.FromBytes(students)
studentsArray := node.Get("students")

// 查找第一个数学学生
_, student, found := studentsArray.FindInArray(func(index int, value fxjson.Node) bool {
    subject, _ := value.Get("subject").String()
    return subject == "数学"
})

if found {
    name, _ := student.Get("name").String()
    grade, _ := student.Get("grade").Int()
    fmt.Printf("第一个数学学生: %s (成绩: %d)\n", name, grade)
}

// 过滤所有高分学生(>90)
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
第一个数学学生: Alice (成绩: 95)
高分学生数量: 2
高分学生 1: Alice (95分)
高分学生 2: Charlie (92分)
```

## ⚙️ 高性能结构体解码

FxJSON提供多种优化的解码方法满足不同性能需求:

### 标准解码(基于Node)

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

### 直接解码(优化版)

为了更好的性能，可以直接从字节解码而不创建Node:

```go
// DecodeStruct - 直接从字节解码(更快)
var user1 User
if err := fxjson.DecodeStruct(jsonData, &user1); err != nil {
    fmt.Printf("DecodeStruct错误: %v\n", err)
} else {
    fmt.Printf("DecodeStruct结果: %+v\n", user1)
}

// DecodeStructFast - 超快解码(最快)
var user2 User
if err := fxjson.DecodeStructFast(jsonData, &user2); err != nil {
    fmt.Printf("DecodeStructFast错误: %v\n", err)
} else {
    fmt.Printf("DecodeStructFast结果: %+v\n", user2)
}
```

**输出:**
```
DecodeStruct结果: {Name:开发者 Age:28 Tags:[golang json performance] Email:dev@example.com}
DecodeStructFast结果: {Name:开发者 Age:28 Tags:[golang json performance] Email:dev@example.com}
```

### 性能对比

| 方法 | 速度 | 使用场景 |
|------|------|----------|
| `node.Decode()` | 快 | 需要Node功能时 |
| `DecodeStruct()` | 更快 | 直接结构体解码 |
| `DecodeStructFast()` | 最快 | 性能关键场景 |

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
    fmt.Println("missing_field存在")
} else {
    fmt.Println("missing_field不存在")
}

if node.HasKey("valid_number") {
    fmt.Println("valid_number存在")
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
missing_field不存在
valid_number存在
使用默认值: 默认值
空字符串长度: 0
```

## 📝 性能优化建议

1. **遍历优化**: 对于大数据集，优先使用`ForEach`、`ArrayForEach`和`Walk`方法
2. **路径访问**: 使用`GetPath`一次性访问深层嵌套字段
3. **内存管理**: 核心遍历操作实现零分配，适合高频场景
4. **类型检查**: 使用`IsXXX()`方法进行类型检查，避免不必要的类型转换
5. **缓存利用**: 数组索引自动缓存，重复访问性能更好
6. **解码优化**: 
   - 需要Node功能时使用`node.Decode()`
   - 直接结构体解码使用`DecodeStruct()`(更快)
   - 性能关键场景使用`DecodeStructFast()`(最快)
   - 根据性能需求选择合适的方法
7. **查询优化**: 使用内置查询功能比手动遍历更高效
8. **缓存策略**: 开启智能缓存可显著提升重复解析性能

## ⚠️ 注意事项

1. **输入验证**: 假设JSON输入有效，专注性能而非错误处理
2. **内存安全**: 所有字符串操作都包含边界检查
3. **Unicode支持**: 完美支持中文、emoji等Unicode字符
4. **并发安全**: Node读操作是并发安全的
5. **Go版本**: 需要Go 1.18或更高版本
6. **空字符串处理**: 已修复空字符串导致的panic问题

## 📚 完整API参考

### 核心方法

#### 节点创建
- `FromBytes(data []byte) Node` - 从JSON字节创建节点，自动展开嵌套JSON
- `FromBytesWithCache(data []byte, ttl time.Duration) Node` - 带缓存的解析
- `FromBytesWithDebug(data []byte) (Node, DebugInfo)` - 带调试信息的解析

#### 基础访问
- `Get(key string) Node` - 通过键获取对象字段
- `GetPath(path string) Node` - 通过路径获取值(如"user.profile.name")
- `Index(i int) Node` - 通过索引获取数组元素

#### 类型检查
- `Exists() bool` - 检查节点是否存在
- `IsObject() bool` - 检查是否为JSON对象
- `IsArray() bool` - 检查是否为JSON数组
- `IsString() bool` - 检查是否为JSON字符串
- `IsNumber() bool` - 检查是否为JSON数字
- `IsBool() bool` - 检查是否为JSON布尔值
- `IsNull() bool` - 检查是否为JSON null
- `IsScalar() bool` - 检查是否为标量类型
- `IsContainer() bool` - 检查是否为容器类型

#### 高级查询
- `Query() *QueryBuilder` - 创建查询构建器
- `Where(field, operator, value)` - 添加查询条件
- `SortBy(field, order)` - 添加排序
- `Limit(count)` - 限制结果数量
- `Count()` - 统计匹配数量
- `First()` - 获取第一个匹配项

#### 数据聚合
- `Aggregate() *Aggregator` - 创建聚合器
- `GroupBy(field)` - 按字段分组
- `Sum(field, alias)` - 求和
- `Avg(field, alias)` - 求平均值
- `Count(alias)` - 计数
- `Max(field, alias)` - 求最大值
- `Min(field, alias)` - 求最小值

#### 数据处理
- `Transform(mapper FieldMapper)` - 数据变换
- `Validate(validator *DataValidator)` - 数据验证
- `Stream(fn StreamFunc)` - 流式处理

#### 缓存管理
- `NewMemoryCache(maxSize int)` - 创建内存缓存
- `EnableCaching(cache Cache)` - 启用缓存
- `DisableCaching()` - 禁用缓存

#### 调试工具
- `EnableDebugMode()` - 启用调试模式
- `DisableDebugMode()` - 禁用调试模式
- `PrettyPrint()` - 美化打印
- `Inspect()` - 结构分析
- `Diff(other Node)` - 差异对比

## 🤝 贡献

欢迎提交Issue和Pull Request!

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

---

**FxJSON - 让JSON解析飞起来!** 🚀