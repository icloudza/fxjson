# 示例和用法

本节提供了 FxJSON 的详细使用示例，按功能分类组织，从基础到高级，帮助您全面掌握库的各种用法。

## 基础示例

### 1. JSON解析和数据访问

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := []byte(`{
        "user": {
            "id": 123,
            "name": "张三",
            "email": "zhangsan@example.com",
            "active": true,
            "score": 95.5,
            "tags": ["golang", "json", "programming"]
        }
    }`)

    // 解析JSON
    node := fxjson.FromBytes(jsonData)
    
    // 基本访问
    userNode := node.Get("user")
    fmt.Printf("用户ID: %d\n", userNode.Get("id").IntOr(0))
    fmt.Printf("用户名: %s\n", userNode.Get("name").StringOr("未知"))
    fmt.Printf("邮箱: %s\n", userNode.Get("email").StringOr(""))
    fmt.Printf("是否激活: %t\n", userNode.Get("active").BoolOr(false))
    fmt.Printf("评分: %.1f\n", userNode.Get("score").FloatOr(0.0))
    
    // 路径访问
    fmt.Printf("直接路径访问名字: %s\n", node.GetPath("user.name").StringOr(""))
    
    // 数组访问
    tags := userNode.Get("tags")
    fmt.Printf("标签数量: %d\n", tags.Len())
    fmt.Printf("第一个标签: %s\n", tags.Index(0).StringOr(""))
}
```

### 2. 数组遍历

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := []byte(`{
        "users": [
            {"name": "张三", "age": 25, "city": "北京"},
            {"name": "李四", "age": 30, "city": "上海"},
            {"name": "王五", "age": 35, "city": "广州"}
        ]
    }`)

    node := fxjson.FromBytes(jsonData)
    users := node.Get("users")

    // 高性能零分配遍历
    fmt.Println("用户列表:")
    users.ArrayForEach(func(index int, user fxjson.Node) bool {
        name := user.Get("name").StringOr("")
        age := user.Get("age").IntOr(0)
        city := user.Get("city").StringOr("")
        
        fmt.Printf("%d. %s，年龄 %d，来自 %s\n", index+1, name, age, city)
        return true // 继续遍历
    })

    // 获取特定索引的用户
    firstUser := users.Index(0)
    fmt.Printf("\n第一个用户: %s\n", firstUser.Get("name").StringOr(""))
}
```

### 3. 对象遍历

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := []byte(`{
        "departments": {
            "engineering": {"count": 25, "budget": 500000},
            "marketing": {"count": 15, "budget": 300000},
            "sales": {"count": 20, "budget": 400000}
        }
    }`)

    node := fxjson.FromBytes(jsonData)
    departments := node.Get("departments")

    // 遍历对象的所有键值对
    fmt.Println("部门信息:")
    departments.ForEach(func(dept string, info fxjson.Node) bool {
        count := info.Get("count").IntOr(0)
        budget := info.Get("budget").IntOr(0)
        
        fmt.Printf("%s: %d人，预算 %d\n", dept, count, budget)
        return true
    })
}
```

## 高级示例

### 4. 数据验证

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := []byte(`{
        "contact": {
            "email": "user@example.com",
            "website": "https://example.com",
            "phone": "+1234567890",
            "ip": "192.168.1.1"
        }
    }`)

    node := fxjson.FromBytes(jsonData)
    contact := node.Get("contact")

    // 验证邮箱
    email := contact.Get("email")
    if email.IsValidEmail() {
        fmt.Printf("邮箱 %s 格式正确\n", email.StringOr(""))
    }

    // 验证URL
    website := contact.Get("website")
    if website.IsValidURL() {
        fmt.Printf("网站 %s 格式正确\n", website.StringOr(""))
    }

    // 验证IP地址
    ip := contact.Get("ip")
    if ip.IsValidIP() {
        fmt.Printf("IP地址 %s 格式正确\n", ip.StringOr(""))
    }
}
```

### 5. 数组数据转换

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := []byte(`{
        "data": {
            "names": ["张三", "李四", "王五"],
            "scores": [95, 87, 92],
            "prices": [19.99, 29.99, 39.99],
            "flags": [true, false, true]
        }
    }`)

    node := fxjson.FromBytes(jsonData)
    data := node.Get("data")

    // 转换为字符串切片
    names, err := data.Get("names").ToStringSlice()
    if err == nil {
        fmt.Printf("姓名列表: %v\n", names)
    }

    // 转换为整数切片
    scores, err := data.Get("scores").ToIntSlice()
    if err == nil {
        fmt.Printf("分数列表: %v\n", scores)
    }

    // 转换为浮点数切片
    prices, err := data.Get("prices").ToFloatSlice()
    if err == nil {
        fmt.Printf("价格列表: %v\n", prices)
    }

    // 转换为布尔值切片
    flags, err := data.Get("flags").ToBoolSlice()
    if err == nil {
        fmt.Printf("标志列表: %v\n", flags)
    }
}
```

### 6. 深度遍历

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := []byte(`{
        "company": {
            "name": "Tech Corp",
            "departments": {
                "engineering": {
                    "teams": ["backend", "frontend", "mobile"],
                    "budget": 500000
                }
            }
        }
    }`)

    node := fxjson.FromBytes(jsonData)

    // 深度遍历所有节点
    fmt.Println("深度遍历结果:")
    node.Walk(func(path string, n fxjson.Node) bool {
        if n.IsString() {
            fmt.Printf("%s (字符串): %s\n", path, n.StringOr(""))
        } else if n.IsNumber() {
            fmt.Printf("%s (数字): %d\n", path, n.IntOr(0))
        } else if n.IsArray() {
            fmt.Printf("%s (数组): 长度 %d\n", path, n.Len())
        } else if n.IsObject() {
            fmt.Printf("%s (对象): %d个字段\n", path, n.Len())
        }
        return true
    })
}
```

### 7. 结构体编解码

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

type User struct {
    ID     int64  `json:"id"`
    Name   string `json:"name"`
    Email  string `json:"email"`
    Active bool   `json:"active"`
}

func main() {
    // 序列化结构体
    user := User{
        ID:     123,
        Name:   "张三",
        Email:  "zhangsan@example.com",
        Active: true,
    }

    // 序列化为JSON
    jsonBytes, err := fxjson.Marshal(user)
    if err != nil {
        fmt.Printf("序列化失败: %v\n", err)
        return
    }
    fmt.Printf("序列化结果: %s\n", jsonBytes)

    // 解析JSON到结构体
    var newUser User
    err = fxjson.DecodeStruct(jsonBytes, &newUser)
    if err != nil {
        fmt.Printf("解码失败: %v\n", err)
        return
    }
    fmt.Printf("解码结果: %+v\n", newUser)
}
```

### 8. 错误处理

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := []byte(`{
        "user": {
            "name": "张三",
            "age": "invalid_number"
        }
    }`)

    node := fxjson.FromBytes(jsonData)
    user := node.Get("user")

    // 方式1: 使用Or方法（推荐）
    name := user.Get("name").StringOr("默认姓名")
    age := user.Get("age").IntOr(0) // 无效数字返回默认值0
    fmt.Printf("安全访问 - 姓名: %s, 年龄: %d\n", name, age)

    // 方式2: 手动处理错误
    nameVal, err := user.Get("name").String()
    if err != nil {
        fmt.Printf("获取姓名失败: %v\n", err)
    } else {
        fmt.Printf("姓名: %s\n", nameVal)
    }

    ageVal, err := user.Get("age").Int()
    if err != nil {
        fmt.Printf("获取年龄失败: %v\n", err)
    } else {
        fmt.Printf("年龄: %d\n", ageVal)
    }

    // 检查字段是否存在
    if user.Get("email").Exists() {
        fmt.Println("邮箱字段存在")
    } else {
        fmt.Println("邮箱字段不存在")
    }
}
```

### 9. 性能优化示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/icloudza/fxjson"
)

func main() {
    // 大JSON数据
    jsonData := []byte(`{
        "users": [` + 
        generateUsers(1000) + // 假设生成1000个用户的函数
        `]
    }`)

    // 性能测试：普通遍历
    start := time.Now()
    node := fxjson.FromBytes(jsonData)
    users := node.Get("users")
    
    count := 0
    users.ArrayForEach(func(index int, user fxjson.Node) bool {
        name := user.Get("name").StringOr("")
        if name != "" {
            count++
        }
        return true
    })
    
    elapsed := time.Since(start)
    fmt.Printf("遍历 %d 个用户耗时: %v\n", count, elapsed)

    // 利用缓存重复访问
    start = time.Now()
    for i := 0; i < 5; i++ {
        firstUser := users.Index(0)
        _ = firstUser.Get("name").StringOr("")
    }
    elapsed = time.Since(start)
    fmt.Printf("重复访问5次耗时: %v\n", elapsed)
}

func generateUsers(count int) string {
    // 简化示例，实际应用中可能从数据库或其他源生成
    return `{"name": "用户1", "age": 25}`
}
```

## 实际应用场景

### 10. API响应处理

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    // 模拟API响应
    apiResponse := []byte(`{
        "status": "success",
        "data": {
            "total": 150,
            "page": 1,
            "per_page": 10,
            "items": [
                {
                    "id": 1,
                    "title": "文章标题1",
                    "author": "作者1",
                    "published": true,
                    "tags": ["技术", "编程"]
                },
                {
                    "id": 2,
                    "title": "文章标题2", 
                    "author": "作者2",
                    "published": false,
                    "tags": ["生活", "随笔"]
                }
            ]
        }
    }`)

    node := fxjson.FromBytes(apiResponse)
    
    // 检查响应状态
    status := node.Get("status").StringOr("")
    if status != "success" {
        fmt.Printf("API调用失败: %s\n", status)
        return
    }

    // 处理分页信息
    data := node.Get("data")
    total := data.Get("total").IntOr(0)
    page := data.Get("page").IntOr(1)
    perPage := data.Get("per_page").IntOr(10)
    
    fmt.Printf("分页信息: 第%d页，每页%d条，共%d条\n", page, perPage, total)

    // 处理文章列表
    items := data.Get("items")
    fmt.Printf("\n文章列表(%d篇):\n", items.Len())
    
    items.ArrayForEach(func(index int, item fxjson.Node) bool {
        id := item.Get("id").IntOr(0)
        title := item.Get("title").StringOr("")
        author := item.Get("author").StringOr("")
        published := item.Get("published").BoolOr(false)
        
        status := "未发布"
        if published {
            status = "已发布"
        }
        
        fmt.Printf("%d. [%s] %s - %s\n", id, status, title, author)
        
        // 处理标签
        tags := item.Get("tags")
        if tags.Exists() && tags.Len() > 0 {
            fmt.Print("   标签: ")
            tags.ArrayForEach(func(i int, tag fxjson.Node) bool {
                if i > 0 {
                    fmt.Print(", ")
                }
                fmt.Print(tag.StringOr(""))
                return true
            })
            fmt.Println()
        }
        
        return true
    })
}
```

这些示例展示了FxJSON的各种使用场景，从基础的JSON解析到复杂的数据处理和验证。通过这些示例，您可以了解如何在实际项目中高效地使用FxJSON。
