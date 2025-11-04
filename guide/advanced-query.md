# 高级查询与数据处理

FxJSON 提供了强大的高级查询、数据转换和聚合功能，让您能够像使用数据库一样查询和处理 JSON 数据。

## 目录

- [数据转换 (Transform)](#数据转换-transform)
- [条件查询 (Query)](#条件查询-query)
- [数据聚合 (Aggregate)](#数据聚合-aggregate)
- [流式处理 (Stream)](#流式处理-stream)
- [数据验证](#数据验证)
- [实战案例](#实战案例)

---

## 数据转换 (Transform)

### Transform()

使用 `Transform()` 方法可以灵活地转换 JSON 数据结构。

```go
func (n Node) Transform(mapper FieldMapper) (map[string]interface{}, error)
```

**FieldMapper 类型**：
```go
type FieldMapper map[string]func(Node) interface{}
```

### 基础转换示例

```go
package main

import (
    "fmt"
    "strings"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := `{
        "user_name": "Zhang_San",
        "user_email": "ZHANG@EXAMPLE.COM",
        "user_age": "28",
        "created_at": "2024-01-15T10:30:00Z",
        "is_active": "true"
    }`

    node := fxjson.FromString(jsonData)

    // 定义字段映射规则
    mapper := fxjson.FieldMapper{
        "name": func(n fxjson.Node) interface{} {
            // 转换为驼峰命名，首字母大写
            name := n.Get("user_name").StringOr("")
            return strings.Title(strings.ReplaceAll(name, "_", " "))
        },
        "email": func(n fxjson.Node) interface{} {
            // 邮箱转小写
            return strings.ToLower(n.Get("user_email").StringOr(""))
        },
        "age": func(n fxjson.Node) interface{} {
            // 字符串转整数
            ageStr := n.Get("user_age").StringOr("0")
            age := 0
            fmt.Sscanf(ageStr, "%d", &age)
            return age
        },
        "status": func(n fxjson.Node) interface{} {
            // 转换活跃状态
            active := n.Get("is_active").StringOr("false") == "true"
            if active {
                return "active"
            }
            return "inactive"
        },
        "timestamp": func(n fxjson.Node) interface{} {
            // 保留时间戳
            return n.Get("created_at").StringOr("")
        },
    }

    // 执行转换
    result, err := node.Transform(mapper)
    if err != nil {
        panic(err)
    }

    fmt.Printf("转换结果: %+v\n", result)
    // 输出: 转换结果: map[age:28 email:zhang@example.com name:Zhang San status:active timestamp:2024-01-15T10:30:00Z]
}
```

### 复杂数据转换

```go
func complexTransform() {
    apiResponse := `{
        "data": {
            "users": [
                {
                    "id": "1001",
                    "full_name": "zhang_san",
                    "contact": {
                        "email": "zhang@example.com",
                        "phone": "+86-138-0000-0000"
                    },
                    "role": "admin",
                    "permissions": ["read", "write", "delete"]
                },
                {
                    "id": "1002",
                    "full_name": "li_si",
                    "contact": {
                        "email": "li@example.com",
                        "phone": "+86-139-0000-0000"
                    },
                    "role": "user",
                    "permissions": ["read"]
                }
            ]
        }
    }`

    node := fxjson.FromString(apiResponse)
    users := node.GetPath("data.users")

    // 转换每个用户
    var transformedUsers []map[string]interface{}

    users.ArrayForEach(func(index int, user fxjson.Node) bool {
        mapper := fxjson.FieldMapper{
            "userId": func(n fxjson.Node) interface{} {
                return n.Get("id").IntOr(0)
            },
            "username": func(n fxjson.Node) interface{} {
                // 下划线转驼峰
                name := n.Get("full_name").StringOr("")
                parts := strings.Split(name, "_")
                result := parts[0]
                for i := 1; i < len(parts); i++ {
                    result += strings.Title(parts[i])
                }
                return result
            },
            "email": func(n fxjson.Node) interface{} {
                return n.GetPath("contact.email").StringOr("")
            },
            "isAdmin": func(n fxjson.Node) interface{} {
                return n.Get("role").StringOr("") == "admin"
            },
            "permissionCount": func(n fxjson.Node) interface{} {
                return n.Get("permissions").Len()
            },
        }

        transformed, _ := user.Transform(mapper)
        transformedUsers = append(transformedUsers, transformed)
        return true
    })

    fmt.Printf("转换后的用户列表: %+v\n", transformedUsers)
}
```

---

## 条件查询 (Query)

### Query() 查询构建器

使用链式 API 构建复杂的查询条件。

```go
func (n Node) Query() *QueryBuilder
```

### QueryBuilder 方法

```go
type QueryBuilder struct {
    // 条件方法
    func (qb *QueryBuilder) Where(field, operator string, value interface{}) *QueryBuilder
    func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder
    func (qb *QueryBuilder) WhereNotIn(field string, values []interface{}) *QueryBuilder
    func (qb *QueryBuilder) WhereContains(field, substring string) *QueryBuilder

    // 排序和分页
    func (qb *QueryBuilder) SortBy(field, order string) *QueryBuilder  // order: "asc" 或 "desc"
    func (qb *QueryBuilder) Limit(count int) *QueryBuilder
    func (qb *QueryBuilder) Offset(offset int) *QueryBuilder

    // 结果获取
    func (qb *QueryBuilder) ToSlice() ([]Node, error)
    func (qb *QueryBuilder) Count() (int, error)
    func (qb *QueryBuilder) First() (Node, error)
}
```

### 支持的查询操作符

- `"="` - 等于
- `"!="` - 不等于
- `">"` - 大于
- `">="` - 大于等于
- `"<"` - 小于
- `"<="` - 小于等于
- `"contains"` - 包含（字符串）
- `"startsWith"` - 开头匹配
- `"endsWith"` - 结尾匹配

### 基础查询示例

```go
func basicQuery() {
    productsJSON := `{
        "products": [
            {"id": 1, "name": "笔记本电脑", "price": 5999, "category": "电子产品", "stock": 50},
            {"id": 2, "name": "机械键盘", "price": 399, "category": "电子产品", "stock": 120},
            {"id": 3, "name": "显示器", "price": 1299, "category": "电子产品", "stock": 80},
            {"id": 4, "name": "鼠标", "price": 99, "category": "电子产品", "stock": 200},
            {"id": 5, "name": "办公椅", "price": 899, "category": "家具", "stock": 30}
        ]
    }`

    node := fxjson.FromString(productsJSON)
    products := node.Get("products")

    // 查询价格大于500的产品
    expensiveProducts, err := products.Query().
        Where("price", ">", 500).
        ToSlice()

    if err != nil {
        panic(err)
    }

    fmt.Printf("找到 %d 个昂贵产品:\n", len(expensiveProducts))
    for _, product := range expensiveProducts {
        name := product.Get("name").StringOr("")
        price := product.Get("price").FloatOr(0)
        fmt.Printf("- %s: ¥%.2f\n", name, price)
    }
}
```

### 复杂查询示例

```go
func advancedQuery() {
    usersJSON := `{
        "users": [
            {"id": 1, "name": "张三", "age": 28, "city": "北京", "salary": 15000, "department": "技术部"},
            {"id": 2, "name": "李四", "age": 32, "city": "上海", "salary": 18000, "department": "技术部"},
            {"id": 3, "name": "王五", "age": 25, "city": "北京", "salary": 12000, "department": "市场部"},
            {"id": 4, "name": "赵六", "age": 35, "city": "深圳", "salary": 22000, "department": "技术部"},
            {"id": 5, "name": "陈七", "age": 29, "city": "上海", "salary": 16000, "department": "产品部"},
            {"id": 6, "name": "周八", "age": 27, "city": "北京", "salary": 14000, "department": "技术部"}
        ]
    }`

    node := fxjson.FromString(usersJSON)
    users := node.Get("users")

    // 多条件查询：技术部 AND 年龄>=28 AND 薪资>=15000
    techSeniors, _ := users.Query().
        Where("department", "=", "技术部").
        Where("age", ">=", 28).
        Where("salary", ">=", 15000).
        SortBy("salary", "desc").
        ToSlice()

    fmt.Println("技术部高级员工:")
    for _, user := range techSeniors {
        fmt.Printf("- %s, 年龄: %d, 薪资: ¥%d, 城市: %s\n",
            user.Get("name").StringOr(""),
            user.Get("age").IntOr(0),
            user.Get("salary").IntOr(0),
            user.Get("city").StringOr(""))
    }

    // WhereIn 查询：在指定城市
    citiesUsers, _ := users.Query().
        WhereIn("city", []interface{}{"北京", "上海"}).
        SortBy("age", "asc").
        ToSlice()

    fmt.Printf("\n北京和上海的用户 (%d人):\n", len(citiesUsers))
    for _, user := range citiesUsers {
        fmt.Printf("- %s, %s, %d岁\n",
            user.Get("name").StringOr(""),
            user.Get("city").StringOr(""),
            user.Get("age").IntOr(0))
    }

    // 分页查询
    page1, _ := users.Query().
        SortBy("id", "asc").
        Limit(3).
        Offset(0).
        ToSlice()

    fmt.Printf("\n第1页 (共 %d 条):\n", len(page1))
    for _, user := range page1 {
        fmt.Printf("- ID: %d, %s\n",
            user.Get("id").IntOr(0),
            user.Get("name").StringOr(""))
    }

    // 获取总数
    count, _ := users.Query().
        Where("department", "=", "技术部").
        Count()
    fmt.Printf("\n技术部总人数: %d\n", count)

    // 获取第一条
    firstUser, _ := users.Query().
        Where("city", "=", "北京").
        SortBy("age", "asc").
        First()
    fmt.Printf("\n北京最年轻的员工: %s, %d岁\n",
        firstUser.Get("name").StringOr(""),
        firstUser.Get("age").IntOr(0))
}
```

### 字符串查询

```go
func stringQuery() {
    articlesJSON := `{
        "articles": [
            {"id": 1, "title": "Go 语言入门教程", "author": "张三", "tags": ["golang", "tutorial"]},
            {"id": 2, "title": "Python 数据分析实战", "author": "李四", "tags": ["python", "data"]},
            {"id": 3, "title": "Go 高性能编程", "author": "王五", "tags": ["golang", "performance"]},
            {"id": 4, "title": "Java 设计模式", "author": "赵六", "tags": ["java", "design"]},
            {"id": 5, "title": "Go 微服务架构", "author": "陈七", "tags": ["golang", "microservice"]}
        ]
    }`

    node := fxjson.FromString(articlesJSON)
    articles := node.Get("articles")

    // 标题包含 "Go" 的文章
    goArticles, _ := articles.Query().
        WhereContains("title", "Go").
        ToSlice()

    fmt.Printf("找到 %d 篇 Go 相关文章:\n", len(goArticles))
    for _, article := range goArticles {
        fmt.Printf("- %s (作者: %s)\n",
            article.Get("title").StringOr(""),
            article.Get("author").StringOr(""))
    }
}
```

---

## 数据聚合 (Aggregate)

### Aggregate() 聚合器

对数据进行统计分析和聚合计算。

```go
func (n Node) Aggregate() *Aggregator
```

### Aggregator 方法

```go
type Aggregator struct {
    func (agg *Aggregator) Count(alias string) *Aggregator
    func (agg *Aggregator) Sum(field, alias string) *Aggregator
    func (agg *Aggregator) Avg(field, alias string) *Aggregator
    func (agg *Aggregator) Max(field, alias string) *Aggregator
    func (agg *Aggregator) Min(field, alias string) *Aggregator
    func (agg *Aggregator) GroupBy(fields ...string) *Aggregator
    func (agg *Aggregator) Execute(node Node) (map[string]interface{}, error)
}
```

### 基础聚合示例

```go
func basicAggregation() {
    salesJSON := `{
        "sales": [
            {"product": "笔记本", "amount": 5999, "quantity": 2, "region": "北京"},
            {"product": "键盘", "amount": 399, "quantity": 5, "region": "上海"},
            {"product": "鼠标", "amount": 99, "quantity": 10, "region": "北京"},
            {"product": "显示器", "amount": 1299, "quantity": 3, "region": "深圳"},
            {"product": "笔记本", "amount": 5999, "quantity": 1, "region": "上海"}
        ]
    }`

    node := fxjson.FromString(salesJSON)
    sales := node.Get("sales")

    // 简单聚合：总数、总金额、平均金额
    result, err := sales.Aggregate().
        Count("总订单数").
        Sum("amount", "总销售额").
        Avg("amount", "平均订单金额").
        Max("amount", "最大订单").
        Min("amount", "最小订单").
        Execute(sales)

    if err != nil {
        panic(err)
    }

    fmt.Println("销售统计:")
    for key, value := range result {
        fmt.Printf("  %s: %v\n", key, value)
    }
    // 输出:
    // 总订单数: 5
    // 总销售额: 13795
    // 平均订单金额: 2759
    // 最大订单: 5999
    // 最小订单: 99
}
```

### 分组聚合

```go
func groupAggregation() {
    salesJSON := `{
        "sales": [
            {"product": "笔记本", "amount": 5999, "quantity": 2, "region": "北京", "category": "电子产品"},
            {"product": "键盘", "amount": 399, "quantity": 5, "region": "上海", "category": "电子产品"},
            {"product": "鼠标", "amount": 99, "quantity": 10, "region": "北京", "category": "电子产品"},
            {"product": "办公椅", "amount": 899, "quantity": 3, "region": "深圳", "category": "家具"},
            {"product": "书架", "amount": 599, "quantity": 2, "region": "北京", "category": "家具"},
            {"product": "笔记本", "amount": 5999, "quantity": 1, "region": "上海", "category": "电子产品"}
        ]
    }`

    node := fxjson.FromString(salesJSON)
    sales := node.Get("sales")

    // 按地区分组统计
    regionStats, _ := sales.Aggregate().
        GroupBy("region").
        Count("订单数").
        Sum("amount", "销售额").
        Avg("amount", "平均金额").
        Execute(sales)

    fmt.Println("按地区统计:")
    for region, stats := range regionStats {
        fmt.Printf("\n地区: %s\n", region)
        if statsMap, ok := stats.(map[string]interface{}); ok {
            for key, value := range statsMap {
                fmt.Printf("  %s: %v\n", key, value)
            }
        }
    }

    // 按类别分组
    categoryStats, _ := sales.Aggregate().
        GroupBy("category").
        Count("商品数").
        Sum("quantity", "总数量").
        Sum("amount", "总金额").
        Execute(sales)

    fmt.Println("\n\n按类别统计:")
    for category, stats := range categoryStats {
        fmt.Printf("\n类别: %s\n", category)
        if statsMap, ok := stats.(map[string]interface{}); ok {
            for key, value := range statsMap {
                fmt.Printf("  %s: %v\n", key, value)
            }
        }
    }
}
```

### 多维度分组

```go
func multiDimensionAggregation() {
    salesJSON := `{
        "sales": [
            {"date": "2024-01", "product": "笔记本", "amount": 5999, "region": "北京"},
            {"date": "2024-01", "product": "键盘", "amount": 399, "region": "北京"},
            {"date": "2024-01", "product": "笔记本", "amount": 5999, "region": "上海"},
            {"date": "2024-02", "product": "鼠标", "amount": 99, "region": "北京"},
            {"date": "2024-02", "product": "笔记本", "amount": 5999, "region": "上海"}
        ]
    }`

    node := fxjson.FromString(salesJSON)
    sales := node.Get("sales")

    // 按日期和地区分组
    multiStats, _ := sales.Aggregate().
        GroupBy("date", "region").
        Count("订单数").
        Sum("amount", "销售额").
        Execute(sales)

    fmt.Println("按日期和地区统计:")
    for key, stats := range multiStats {
        fmt.Printf("\n%s:\n", key)
        if statsMap, ok := stats.(map[string]interface{}); ok {
            for statKey, value := range statsMap {
                fmt.Printf("  %s: %v\n", statKey, value)
            }
        }
    }
}
```

---

## 流式处理 (Stream)

### Stream()

对大数据集进行流式处理，避免一次性加载到内存。

```go
func (n Node) Stream(processor func(Node, int) bool) error
```

### 流式处理示例

```go
func streamProcessing() {
    // 生成大量数据
    var items []string
    for i := 0; i < 10000; i++ {
        items = append(items, fmt.Sprintf(`{"id": %d, "value": %d}`, i, i*100))
    }
    largeJSON := fmt.Sprintf(`{"items": [%s]}`, strings.Join(items, ","))

    node := fxjson.FromString(largeJSON)
    itemsNode := node.Get("items")

    // 流式处理：只处理符合条件的数据
    count := 0
    sum := 0.0

    err := itemsNode.Stream(func(item fxjson.Node, index int) bool {
        value := item.Get("value").FloatOr(0)

        // 只处理值大于 50000 的项
        if value > 50000 {
            count++
            sum += value
        }

        // 可以提前终止处理
        if count >= 100 {
            return false  // 返回 false 终止
        }

        return true  // 返回 true 继续
    })

    if err != nil {
        panic(err)
    }

    fmt.Printf("流式处理结果:\n")
    fmt.Printf("  处理数量: %d\n", count)
    fmt.Printf("  总值: %.0f\n", sum)
    fmt.Printf("  平均值: %.2f\n", sum/float64(count))
}
```

---

## 数据验证

### Validate()

使用验证器对数据进行校验。

```go
func (n Node) Validate(validator *DataValidator) (map[string]interface{}, []error)
```

### 验证规则示例

```go
func dataValidation() {
    userJSON := `{
        "username": "zhang_san",
        "email": "zhang@example.com",
        "age": 25,
        "phone": "13800138000",
        "website": "https://example.com"
    }`

    node := fxjson.FromString(userJSON)

    // 定义验证规则
    validator := &fxjson.DataValidator{
        Rules: map[string]fxjson.ValidationRule{
            "username": {
                Required: true,
                Type:     "string",
                MinLen:   3,
                MaxLen:   20,
            },
            "email": {
                Required: true,
                Type:     "string",
                Pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
            },
            "age": {
                Required: true,
                Type:     "number",
                Min:      18,
                Max:      100,
            },
            "phone": {
                Required: false,
                Type:     "string",
                Pattern:  `^1[3-9]\d{9}$`,
            },
        },
    }

    // 执行验证
    validData, errors := node.Validate(validator)

    if len(errors) > 0 {
        fmt.Println("验证失败:")
        for _, err := range errors {
            fmt.Printf("  - %v\n", err)
        }
    } else {
        fmt.Println("验证通过!")
        fmt.Printf("有效数据: %+v\n", validData)
    }
}
```

---

## 实战案例

### 案例1：电商订单分析

```go
func ecommerceAnalysis() {
    ordersJSON := `{
        "orders": [
            {"id": 1001, "user": "user1", "amount": 299, "status": "completed", "city": "北京", "category": "电子"},
            {"id": 1002, "user": "user2", "amount": 159, "status": "completed", "city": "上海", "category": "服装"},
            {"id": 1003, "user": "user1", "amount": 89, "status": "cancelled", "city": "北京", "category": "图书"},
            {"id": 1004, "user": "user3", "amount": 599, "status": "completed", "city": "深圳", "category": "电子"},
            {"id": 1005, "user": "user2", "amount": 199, "status": "completed", "city": "上海", "category": "服装"}
        ]
    }`

    node := fxjson.FromString(ordersJSON)
    orders := node.Get("orders")

    // 1. 查询已完成订单
    completedOrders, _ := orders.Query().
        Where("status", "=", "completed").
        ToSlice()
    fmt.Printf("已完成订单数: %d\n", len(completedOrders))

    // 2. 统计各城市销售额
    cityStats, _ := orders.Query().
        Where("status", "=", "completed").
        ToSlice()

    citySales := make(map[string]float64)
    for _, order := range cityStats {
        city := order.Get("city").StringOr("")
        amount := order.Get("amount").FloatOr(0)
        citySales[city] += amount
    }

    fmt.Println("\n各城市销售额:")
    for city, amount := range citySales {
        fmt.Printf("  %s: ¥%.2f\n", city, amount)
    }

    // 3. 聚合统计
    stats, _ := orders.Aggregate().
        Count("总订单数").
        Sum("amount", "总销售额").
        Avg("amount", "平均订单金额").
        Execute(orders)

    fmt.Println("\n整体统计:")
    for key, value := range stats {
        fmt.Printf("  %s: %v\n", key, value)
    }
}
```

### 案例2：日志分析

```go
func logAnalysis() {
    logsJSON := `{
        "logs": [
            {"level": "INFO", "message": "Server started", "timestamp": "2024-01-15 10:00:00"},
            {"level": "ERROR", "message": "Database connection failed", "timestamp": "2024-01-15 10:05:00"},
            {"level": "WARN", "message": "High memory usage", "timestamp": "2024-01-15 10:10:00"},
            {"level": "ERROR", "message": "API timeout", "timestamp": "2024-01-15 10:15:00"},
            {"level": "INFO", "message": "Request processed", "timestamp": "2024-01-15 10:20:00"}
        ]
    }`

    node := fxjson.FromString(logsJSON)
    logs := node.Get("logs")

    // 查询错误日志
    errorLogs, _ := logs.Query().
        Where("level", "=", "ERROR").
        ToSlice()

    fmt.Printf("发现 %d 条错误日志:\n", len(errorLogs))
    for _, log := range errorLogs {
        fmt.Printf("  [%s] %s\n",
            log.Get("timestamp").StringOr(""),
            log.Get("message").StringOr(""))
    }

    // 统计各级别日志数量
    levelCount := make(map[string]int)
    logs.ArrayForEach(func(index int, log fxjson.Node) bool {
        level := log.Get("level").StringOr("")
        levelCount[level]++
        return true
    })

    fmt.Println("\n日志级别统计:")
    for level, count := range levelCount {
        fmt.Printf("  %s: %d\n", level, count)
    }
}
```

---

## 最佳实践

### 1. 性能优化

```go
// 对于大数据集，优先使用流式处理
itemsNode.Stream(func(item fxjson.Node, index int) bool {
    // 处理逻辑
    return true
})

// 而不是一次性加载
allItems, _ := itemsNode.ToSlice()  // 可能占用大量内存
```

### 2. 查询优化

```go
// 链式查询，一次性应用所有条件
results, _ := data.Query().
    Where("status", "=", "active").
    Where("age", ">=", 18).
    SortBy("created_at", "desc").
    Limit(10).
    ToSlice()

// 避免多次查询
```

### 3. 合理使用聚合

```go
// 对于需要多个统计指标，使用聚合器一次性计算
stats, _ := data.Aggregate().
    Count("count").
    Sum("amount", "total").
    Avg("amount", "average").
    Execute(data)

// 避免多次遍历数据
```

高级查询功能让 FxJSON 不仅仅是一个 JSON 解析库，更是一个强大的数据处理工具。通过合理使用这些功能，您可以高效地处理和分析 JSON 数据。
