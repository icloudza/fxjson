# 实用示例集合

本页面精选了 FxJSON 在各种真实场景中的应用示例，代码精简易懂，可直接用于生产环境。

## 目录

- [基础操作](#基础操作) - JSON 解析、数据访问、类型转换
- [数组处理](#数组处理) - 遍历、过滤、统计  
- [对象操作](#对象操作) - 键值遍历、条件查找
- [API 响应处理](#api-响应处理) - 实际业务场景
- [配置文件解析](#配置文件解析) - 应用配置管理
- [数据验证](#数据验证) - 输入校验和格式验证
- [性能优化](#性能优化) - 高性能使用技巧

---

## 基础操作

### 示例 1: 用户信息解析
处理常见的用户数据 JSON。

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    userJSON := `{
        "id": 12345,
        "name": "张三",
        "email": "zhangsan@example.com", 
        "age": 28,
        "vip": true,
        "balance": 1299.50
    }`

    user := fxjson.FromString(userJSON)
    
    // 基础信息提取
    id := user.Get("id").IntOr(0)
    name := user.Get("name").StringOr("匿名用户")
    email := user.Get("email").StringOr("")
    age := user.Get("age").IntOr(0)
    isVip := user.Get("vip").BoolOr(false)
    balance := user.Get("balance").FloatOr(0.0)
    
    fmt.Printf("用户 #%d: %s\n", id, name)
    fmt.Printf("邮箱: %s, 年龄: %d岁\n", email, age)
    fmt.Printf("VIP用户: %t, 余额: ¥%.2f\n", isVip, balance)
}
```

### 示例 2: 嵌套数据访问
处理多层嵌套的 JSON 数据。

```go
func parseNestedData() {
    json := `{
        "company": {
            "name": "科技公司",
            "address": {
                "country": "中国",
                "city": "北京",
                "district": "朝阳区"
            },
            "contact": {
                "phone": "010-12345678",
                "email": "info@company.com"
            }
        }
    }`
    
    node := fxjson.FromString(json)
    
    // 方式1: 链式访问
    companyName := node.Get("company").Get("name").StringOr("")
    
    // 方式2: 路径访问（推荐）
    city := node.GetPath("company.address.city").StringOr("")
    phone := node.GetPath("company.contact.phone").StringOr("")
    
    fmt.Printf("%s 位于 %s，联系电话: %s\n", companyName, city, phone)
}
```

---

## 数组处理

### 示例 3: 商品列表处理
电商场景中的商品数据处理。

```go
func processProducts() {
    productJSON := `{
        "products": [
            {"id": 1, "name": "手机", "price": 3999, "stock": 50},
            {"id": 2, "name": "笔记本", "price": 7999, "stock": 30},
            {"id": 3, "name": "耳机", "price": 299, "stock": 100}
        ]
    }`
    
    node := fxjson.FromString(productJSON)
    products := node.Get("products")
    
    // 商品总览
    fmt.Printf("商品总数: %d\n", products.Len())
    
    // 高性能遍历 - 零分配
    var totalValue float64
    var lowStockCount int
    
    products.ArrayForEach(func(i int, product fxjson.Node) bool {
        name := product.Get("name").StringOr("")
        price := product.Get("price").FloatOr(0)
        stock := product.Get("stock").IntOr(0)
        
        fmt.Printf("%d. %s - ¥%.0f (库存: %d)\n", i+1, name, price, stock)
        
        totalValue += price * float64(stock)
        if stock < 50 {
            lowStockCount++
        }
        
        return true // 继续遍历
    })
    
    fmt.Printf("\n库存总价值: ¥%.0f\n", totalValue)
    fmt.Printf("低库存商品数: %d\n", lowStockCount)
}
```

### 示例 4: 成绩统计分析
教育场景中的成绩数据处理。

```go
func analyzeScores() {
    scoresJSON := `{
        "class": "高三(1)班",
        "subject": "数学",
        "scores": [95, 87, 92, 78, 89, 94, 82, 96, 88, 91]
    }`
    
    node := fxjson.FromString(scoresJSON)
    className := node.Get("class").StringOr("")
    subject := node.Get("subject").StringOr("")
    scores := node.Get("scores")
    
    // 转换为切片进行复杂计算
    scoreList, err := scores.ToIntSlice()
    if err != nil {
        fmt.Printf("分数数据格式错误: %v\n", err)
        return
    }
    
    // 统计分析
    var total, max, min int64
    min = 100 // 初始化最小值
    
    for _, score := range scoreList {
        total += score
        if score > max { max = score }
        if score < min { min = score }
    }
    
    average := float64(total) / float64(len(scoreList))
    fmt.Printf("%s %s 成绩分析:\n", className, subject)
    fmt.Printf("平均分: %.1f, 最高分: %d, 最低分: %d\n", average, max, min)
}
```

---

## 对象操作

### 示例 5: 部门员工统计
企业管理场景中的部门数据处理。

```go
func analyzeDepartments() {
    deptJSON := `{
        "departments": {
            "技术部": {"employees": 25, "budget": 500000, "manager": "李经理"},
            "市场部": {"employees": 15, "budget": 300000, "manager": "王经理"}, 
            "人事部": {"employees": 8, "budget": 200000, "manager": "赵经理"}
        }
    }`
    
    node := fxjson.FromString(deptJSON)
    departments := node.Get("departments")
    
    var totalEmployees int64
    var totalBudget int64
    
    fmt.Println("部门概况:")
    departments.ForEach(func(deptName string, deptInfo fxjson.Node) bool {
        employees := deptInfo.Get("employees").IntOr(0)
        budget := deptInfo.Get("budget").IntOr(0)
        manager := deptInfo.Get("manager").StringOr("")
        
        fmt.Printf("• %s: %d人, 预算¥%d, 负责人: %s\n", 
                   deptName, employees, budget, manager)
        
        totalEmployees += employees
        totalBudget += budget
        return true
    })
    
    fmt.Printf("\n总计: %d人, 总预算: ¥%d\n", totalEmployees, totalBudget)
}
```

---

## API 响应处理

### 示例 6: 分页数据处理  
处理带分页信息的 API 响应。

```go
func handlePaginatedAPI() {
    apiResponse := `{
        "status": 200,
        "message": "success", 
        "data": {
            "total": 156,
            "page": 1,
            "per_page": 10,
            "pages": 16,
            "items": [
                {"id": 1, "title": "文章标题1", "author": "作者A", "views": 1200},
                {"id": 2, "title": "文章标题2", "author": "作者B", "views": 800}
            ]
        }
    }`
    
    response := fxjson.FromString(apiResponse)
    
    // API 状态检查
    status := response.Get("status").IntOr(0)
    if status != 200 {
        fmt.Printf("API 错误: %s\n", response.Get("message").StringOr("未知错误"))
        return
    }
    
    // 分页信息
    data := response.Get("data")
    total := data.Get("total").IntOr(0)
    page := data.Get("page").IntOr(1)
    perPage := data.Get("per_page").IntOr(10)
    pages := data.Get("pages").IntOr(1)
    
    fmt.Printf("第 %d/%d 页 (每页%d条，共%d条)\n", page, pages, perPage, total)
    
    // 处理数据项
    items := data.Get("items")
    items.ArrayForEach(func(i int, item fxjson.Node) bool {
        title := item.Get("title").StringOr("")
        author := item.Get("author").StringOr("")
        views := item.Get("views").IntOr(0)
        
        fmt.Printf("%d. %s - %s (浏览量: %d)\n", i+1, title, author, views)
        return true
    })
}
```

### 示例 7: 错误响应处理
优雅处理 API 错误情况。

```go
func handleAPIError() {
    errorResponse := `{
        "error": {
            "code": 400,
            "message": "用户名不能为空",
            "details": {
                "field": "username",
                "reason": "required_field_missing"
            }
        }
    }`
    
    response := fxjson.FromString(errorResponse)
    
    // 检查是否有错误
    if response.Get("error").Exists() {
        error := response.Get("error")
        code := error.Get("code").IntOr(0)
        message := error.Get("message").StringOr("未知错误")
        
        fmt.Printf("错误 %d: %s\n", code, message)
        
        // 处理详细错误信息
        details := error.Get("details")
        if details.Exists() {
            field := details.Get("field").StringOr("")
            reason := details.Get("reason").StringOr("")
            fmt.Printf("字段: %s, 原因: %s\n", field, reason)
        }
    }
}
```

---

## 配置文件解析

### 示例 8: 应用配置管理
处理应用程序的配置文件。

```go
func loadAppConfig() {
    configJSON := `{
        "app": {
            "name": "我的应用",
            "version": "1.2.0",
            "debug": false
        },
        "database": {
            "host": "localhost",
            "port": 3306,
            "name": "myapp",
            "username": "admin"
        },
        "features": {
            "enable_cache": true,
            "max_connections": 100,
            "timeout": 30
        }
    }`
    
    config := fxjson.FromString(configJSON)
    
    // 应用配置
    appName := config.GetPath("app.name").StringOr("未命名应用")
    version := config.GetPath("app.version").StringOr("1.0.0")
    debug := config.GetPath("app.debug").BoolOr(false)
    
    // 数据库配置
    dbHost := config.GetPath("database.host").StringOr("localhost")
    dbPort := config.GetPath("database.port").IntOr(3306)
    dbName := config.GetPath("database.name").StringOr("default")
    
    // 功能配置
    enableCache := config.GetPath("features.enable_cache").BoolOr(false)
    maxConn := config.GetPath("features.max_connections").IntOr(50)
    timeout := config.GetPath("features.timeout").IntOr(10)
    
    fmt.Printf("应用: %s v%s (调试模式: %t)\n", appName, version, debug)
    fmt.Printf("数据库: %s:%d/%s\n", dbHost, dbPort, dbName)
    fmt.Printf("缓存: %t, 最大连接数: %d, 超时: %ds\n", enableCache, maxConn, timeout)
}
```

---

## 数据验证

### 示例 9: 用户注册验证
验证用户注册表单数据。

```go
func validateUserRegistration() {
    regData := `{
        "username": "newuser123",
        "email": "user@example.com",
        "password": "secretpass",
        "age": 25,
        "website": "https://mysite.com",
        "phone": "+8613800138000"
    }`
    
    user := fxjson.FromString(regData)
    
    var errors []string
    
    // 用户名验证
    username := user.Get("username").StringOr("")
    if len(username) < 3 {
        errors = append(errors, "用户名长度不能少于3位")
    }
    
    // 邮箱验证
    email := user.Get("email")
    if !email.IsValidEmail() {
        errors = append(errors, "邮箱格式不正确")
    }
    
    // 年龄验证
    age := user.Get("age")
    if !age.InRange(18, 100) {
        errors = append(errors, "年龄必须在18-100之间")
    }
    
    // URL验证
    website := user.Get("website")
    if website.Exists() && !website.IsValidURL() {
        errors = append(errors, "网站地址格式不正确")
    }
    
    // 结果输出
    if len(errors) == 0 {
        fmt.Println("用户注册信息验证通过!")
        fmt.Printf("欢迎用户: %s (%s)\n", username, email.StringOr(""))
    } else {
        fmt.Println("验证失败:")
        for i, err := range errors {
            fmt.Printf("%d. %s\n", i+1, err)
        }
    }
}
```

---

## 性能优化

### 示例 10: 大数据处理优化
处理大量数据时的性能优化技巧。

```go
func processBigData() {
    // 模拟大数据场景
    bigDataJSON := `{
        "records": [
            {"id": 1, "value": 100}, {"id": 2, "value": 200},
            {"id": 3, "value": 300}, {"id": 4, "value": 400}
        ]
    }`
    
    data := fxjson.FromString(bigDataJSON)
    records := data.Get("records")
    
    // 优化技巧1: 使用ArrayForEach而不是索引遍历
    var sum int64
    count := 0
    
    // 高性能遍历 - 零分配
    records.ArrayForEach(func(i int, record fxjson.Node) bool {
        value := record.Get("value").IntOr(0)
        sum += value
        count++
        
        // 可以提前退出遍历
        if count >= 1000 { // 处理前1000条
            return false // 停止遍历
        }
        
        return true // 继续
    })
    
    fmt.Printf("处理了 %d 条记录，总计: %d\n", count, sum)
    
    // 优化技巧2: 利用缓存机制
    // 重复访问相同路径会自动缓存
    for i := 0; i < 5; i++ {
        firstValue := data.GetPath("records.0.value").IntOr(0)
        // 第2-5次访问会使用缓存，速度快4倍
        fmt.Printf("第一条记录的值: %d\n", firstValue)
    }
}
```

### 示例 11: 内存优化示例
最小化内存分配的使用方式。

```go
func memoryOptimized() {
    json := `{"users": ["Alice", "Bob", "Charlie"]}`
    node := fxjson.FromString(json)
    users := node.Get("users")
    
    // 方式1: 零分配遍历（推荐）
    fmt.Println("零分配方式:")
    users.ArrayForEach(func(i int, user fxjson.Node) bool {
        name := user.StringOr("")  // 零分配
        fmt.Printf("用户%d: %s\n", i+1, name)
        return true
    })
    
    // 方式2: 避免不必要的切片转换
    // 如果只需要遍历，不要使用ToStringSlice()
    fmt.Println("\n推荐的访问方式:")
    for i := 0; i < users.Len(); i++ {
        name := users.Index(i).StringOr("")  // 零分配
        fmt.Printf("用户%d: %s\n", i+1, name)
    }
}
```

---

## 实际应用场景

### 示例 12: 微服务配置中心
微服务架构中的配置管理。

```go
func microserviceConfig() {
    serviceConfig := `{
        "service": {
            "name": "user-service",
            "port": 8080,
            "environment": "production"
        },
        "dependencies": {
            "redis": {"host": "redis.example.com", "port": 6379},
            "mysql": {"host": "db.example.com", "port": 3306, "database": "users"}
        },
        "monitoring": {
            "metrics_enabled": true,
            "health_check_interval": 30,
            "log_level": "info"
        }
    }`
    
    config := fxjson.FromString(serviceConfig)
    
    // 服务基本信息
    serviceName := config.GetPath("service.name").StringOr("unknown-service")
    port := config.GetPath("service.port").IntOr(8080)
    env := config.GetPath("service.environment").StringOr("development")
    
    fmt.Printf("服务: %s, 端口: %d, 环境: %s\n", serviceName, port, env)
    
    // 依赖服务配置
    dependencies := config.Get("dependencies")
    dependencies.ForEach(func(serviceName string, serviceConfig fxjson.Node) bool {
        host := serviceConfig.Get("host").StringOr("")
        port := serviceConfig.Get("port").IntOr(0)
        fmt.Printf("依赖服务 %s: %s:%d\n", serviceName, host, port)
        return true
    })
    
    // 监控配置
    metricsEnabled := config.GetPath("monitoring.metrics_enabled").BoolOr(false)
    healthInterval := config.GetPath("monitoring.health_check_interval").IntOr(60)
    logLevel := config.GetPath("monitoring.log_level").StringOr("error")
    
    fmt.Printf("监控: 指标收集=%t, 健康检查间隔=%ds, 日志级别=%s\n", 
               metricsEnabled, healthInterval, logLevel)
}
```

---

## 总结

这些示例涵盖了 FxJSON 在实际项目中的主要使用场景：

### 性能要点
- 优先使用 `ArrayForEach()` 和 `ForEach()` 进行遍历
- 使用 `StringOr()` 等安全方法避免错误处理代码
- 重复访问相同路径会自动缓存加速

### 最佳实践  
- 深层嵌套使用 `GetPath()` 路径访问
- 先检查 `Exists()` 再处理可选字段
- 使用内置验证方法简化数据校验

### 常见模式
- API 响应: 先检查状态码，再处理数据
- 配置文件: 提供合理默认值
- 数组处理: 结合遍历和统计计算
- 错误处理: 使用 `Or` 方法优雅处理异常

通过这些精选示例，您可以快速在自己的项目中应用 FxJSON，提升 JSON 处理的效率和代码质量。