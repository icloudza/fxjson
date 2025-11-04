# 错误处理完整指南

FxJSON 提供了完善的错误处理机制,包括详细的错误类型、位置信息和恢复策略,帮助您构建健壮的应用程序。

## 目录

- [错误处理理念](#错误处理理念)
- [错误类型系统](#错误类型系统)
- [安全模式 vs 严格模式](#安全模式-vs-严格模式)
- [错误信息与定位](#错误信息与定位)
- [错误恢复策略](#错误恢复策略)
- [最佳实践](#最佳实践)

---

## 错误处理理念

FxJSON 的错误处理设计遵循以下原则:

1. **双模式支持**: 提供安全模式(默认值)和严格模式(错误返回)
2. **详细错误信息**: 包含错误类型、位置、上下文
3. **零惊讶原则**: 错误行为符合直觉
4. **生产就绪**: 易于集成到现有错误处理流程

---

## 错误类型系统

### FxJSONError 结构

```go
type FxJSONError struct {
    Type      ErrorType  // 错误类型
    Message   string     // 错误消息
    Position  Position   // 错误位置
    Context   string     // 错误上下文
    Cause     error      // 原因错误(可选)
}

type ErrorType int

const (
    ErrorParse        ErrorType = iota  // 解析错误
    ErrorTypeMismatch                   // 类型不匹配
    ErrorNotFound                       // 字段不存在
    ErrorValidation                     // 验证错误
    ErrorDepthLimit                     // 深度限制
    ErrorMemoryLimit                    // 内存限制
)

type Position struct {
    Line   int  // 行号(从1开始)
    Column int  // 列号(从1开始)
    Offset int  // 字节偏移量
}
```

### 错误类型详解

#### 1. 解析错误 (ErrorParse)

JSON 格式错误导致的解析失败。

```go
func parseErrorExample() {
    // 无效的 JSON
    invalidJSONs := []string{
        `{"name": "张三",}`,           // 尾随逗号
        `{"name": "张三"`,              // 缺少右括号
        `{"name": 张三}`,               // 字符串未加引号
        `{name: "张三"}`,               // 键未加引号
        `{"age": 28a}`,                // 无效数字
    }

    for _, jsonStr := range invalidJSONs {
        node := fxjson.FromString(jsonStr)

        if !node.Exists() {
            fmt.Printf("❌ 解析失败: %s\n", jsonStr)
            fmt.Printf("   提示: 检查 JSON 格式是否正确\n\n")
        }
    }
}
```

#### 2. 类型不匹配 (ErrorTypeMismatch)

尝试将数据转换为不兼容的类型。

```go
func typeMismatchExample() {
    jsonData := `{
        "id": "abc123",
        "age": "twenty-five",
        "active": "yes",
        "score": true
    }`

    node := fxjson.FromString(jsonData)

    // 使用严格模式捕获类型错误
    id, err := node.Get("id").Int()
    if err != nil {
        if fxErr, ok := err.(*fxjson.FxJSONError); ok {
            fmt.Printf("类型错误: %s\n", fxErr.Message)
            fmt.Printf("位置: 行 %d, 列 %d\n", fxErr.Position.Line, fxErr.Position.Column)
            fmt.Printf("期望类型: 整数, 实际类型: %s\n\n", fxErr.Context)
        }
    }

    // 安全模式自动处理
    idSafe := node.Get("id").IntOr(0)
    fmt.Printf("安全模式ID: %d (使用默认值)\n", idSafe)
}
```

#### 3. 字段不存在 (ErrorNotFound)

访问不存在的字段。

```go
func notFoundExample() {
    jsonData := `{"user": {"name": "张三"}}`
    node := fxjson.FromString(jsonData)

    // 严格模式
    email, err := node.GetPath("user.email").String()
    if err != nil {
        if fxErr, ok := err.(*fxjson.FxJSONError); ok {
            if fxErr.Type == fxjson.ErrorNotFound {
                fmt.Printf("字段不存在: %s\n", fxErr.Message)
                fmt.Println("建议: 使用 Exists() 检查或使用安全模式")
            }
        }
    }

    // 推荐: 使用 Exists() 检查
    emailNode := node.GetPath("user.email")
    if emailNode.Exists() {
        fmt.Println("邮箱:", emailNode.StringOr(""))
    } else {
        fmt.Println("⚠️ 邮箱字段不存在,使用默认值")
    }

    // 或使用安全模式
    emailSafe := node.GetPath("user.email").StringOr("no-email@example.com")
    fmt.Printf("邮箱(安全): %s\n", emailSafe)
}
```

#### 4. 验证错误 (ErrorValidation)

数据验证失败。

```go
func validationErrorExample() {
    jsonData := `{
        "email": "invalid-email",
        "age": 150,
        "phone": "12345"
    }`

    node := fxjson.FromString(jsonData)

    // 验证邮箱
    email := node.Get("email")
    if email.Exists() && !email.IsValidEmail() {
        fmt.Println("❌ 邮箱格式无效:", email.StringOr(""))
    }

    // 验证年龄范围
    age := node.Get("age")
    if age.Exists() && !age.InRange(0, 120) {
        fmt.Printf("❌ 年龄 %d 超出合理范围\n", age.IntOr(0))
    }

    // 自定义验证
    phone := node.Get("phone").StringOr("")
    if len(phone) != 11 {
        fmt.Printf("❌ 电话号码 %s 长度不正确\n", phone)
    }
}
```

#### 5. 深度限制错误 (ErrorDepthLimit)

JSON 嵌套深度超过限制。

```go
func depthLimitExample() {
    // 创建深度嵌套的 JSON
    deepJSON := `{"level1": {"level2": {"level3": {"level4": {"level5": {"level6": "deep"}}}}}}`

    // 使用自定义解析选项
    opts := fxjson.ParseOptions{
        MaxDepth: 3,  // 限制最大深度为 3
    }

    node := fxjson.FromStringWithOptions(deepJSON, opts)

    // 尝试访问超出深度的数据
    value := node.GetPath("level1.level2.level3.level4")
    if !value.Exists() {
        fmt.Println("❌ 深度超过限制,数据被截断")
    }
}
```

#### 6. 内存限制错误 (ErrorMemoryLimit)

数据大小超过配置的限制。

```go
func memoryLimitExample() {
    opts := fxjson.ParseOptions{
        MaxStringLen:  100,    // 限制字符串最大 100 字节
        MaxObjectKeys: 10,     // 限制对象最多 10 个键
        MaxArrayItems: 50,     // 限制数组最多 50 个元素
    }

    // 超长字符串
    longString := strings.Repeat("x", 200)
    jsonData := fmt.Sprintf(`{"text": "%s"}`, longString)

    node := fxjson.FromStringWithOptions(jsonData, opts)
    text := node.Get("text").StringOr("")

    if len(text) < 200 {
        fmt.Println("⚠️ 字符串被截断以满足内存限制")
    }
}
```

---

## 安全模式 vs 严格模式

### 安全模式 (推荐用于大多数场景)

使用 `*Or()` 方法,自动处理错误并返回默认值。

```go
func safeModeExample() {
    jsonData := `{"user": {"name": "张三", "age": "invalid"}}`
    node := fxjson.FromString(jsonData)

    // 所有操作都是安全的,不会 panic
    name := node.GetPath("user.name").StringOr("未知")        // "张三"
    age := node.GetPath("user.age").IntOr(0)                 // 0 (转换失败)
    email := node.GetPath("user.email").StringOr("无")        // "无" (不存在)
    active := node.GetPath("user.active").BoolOr(false)      // false (不存在)

    fmt.Printf("用户: %s, 年龄: %d, 邮箱: %s, 活跃: %t\n",
        name, age, email, active)
}
```

**优点**:
- 代码简洁,无需大量 `if err != nil`
- 自动降级,提供默认值
- 适合大多数业务场景

### 严格模式 (用于需要明确错误处理的场景)

使用返回 error 的方法,完全掌控错误处理。

```go
func strictModeExample() {
    jsonData := `{"user": {"name": "张三", "age": 28}}`
    node := fxjson.FromString(jsonData)

    // 严格获取,需要显式处理错误
    name, err := node.GetPath("user.name").String()
    if err != nil {
        log.Fatalf("获取姓名失败: %v", err)
        return
    }

    age, err := node.GetPath("user.age").Int()
    if err != nil {
        log.Fatalf("获取年龄失败: %v", err)
        return
    }

    // 验证数据合法性
    if age < 0 || age > 150 {
        log.Fatalf("年龄 %d 不合理", age)
        return
    }

    fmt.Printf("验证通过: %s, %d岁\n", name, age)
}
```

**优点**:
- 明确的错误处理
- 适合关键数据验证
- 便于错误追踪和日志

### 混合使用

根据场景选择合适的模式。

```go
func hybridModeExample() {
    jsonData := `{
        "order": {
            "id": 1001,
            "amount": 299.99,
            "user": "张三",
            "note": ""
        }
    }`

    node := fxjson.FromString(jsonData)
    order := node.Get("order")

    // 关键字段使用严格模式
    orderID, err := order.Get("id").Int()
    if err != nil {
        return fmt.Errorf("订单ID无效: %w", err)
    }

    amount, err := order.Get("amount").Float()
    if err != nil {
        return fmt.Errorf("订单金额无效: %w", err)
    }

    // 非关键字段使用安全模式
    user := order.Get("user").StringOr("匿名用户")
    note := order.Get("note").StringOr("无备注")

    fmt.Printf("订单 #%d: 金额 ¥%.2f, 用户: %s, 备注: %s\n",
        orderID, amount, user, note)

    return nil
}
```

---

## 错误信息与定位

### 错误位置计算

```go
func errorLocationExample() {
    jsonData := `{
        "users": [
            {"name": "张三", "age": 28},
            {"name": "李四", "age": "invalid"},
            {"name": "王五", "age": 35}
        ]
    }`

    node := fxjson.FromString(jsonData)
    users := node.Get("users")

    users.ArrayForEach(func(index int, user fxjson.Node) bool {
        name := user.Get("name").StringOr("未知")

        age, err := user.Get("age").Int()
        if err != nil {
            if fxErr, ok := err.(*fxjson.FxJSONError); ok {
                fmt.Printf("用户 %s 的年龄错误:\n", name)
                fmt.Printf("  消息: %s\n", fxErr.Message)
                fmt.Printf("  位置: 行 %d, 列 %d\n",
                    fxErr.Position.Line, fxErr.Position.Column)
                fmt.Printf("  上下文: %s\n", fxErr.Context)

                // 提供修复建议
                fmt.Printf("  建议: 将年龄改为数字类型\n\n")
            }
        }

        return true
    })
}
```

### 计算错误位置

```go
func calculatePosition() {
    data := []byte(`{
        "user": {
            "name": "张三",
            "age": invalid
        }
    }`)

    // 假设错误在字节偏移量 52 处
    errorOffset := 52

    // 计算位置
    pos := fxjson.CalculatePosition(data, errorOffset)

    fmt.Printf("错误位置:\n")
    fmt.Printf("  行: %d\n", pos.Line)
    fmt.Printf("  列: %d\n", pos.Column)
    fmt.Printf("  偏移: %d\n", pos.Offset)

    // 显示错误上下文
    lines := strings.Split(string(data), "\n")
    if pos.Line > 0 && pos.Line <= len(lines) {
        errorLine := lines[pos.Line-1]
        fmt.Printf("\n错误行:\n%s\n", errorLine)
        fmt.Printf("%s^\n", strings.Repeat(" ", pos.Column-1))
    }
}
```

---

## 错误恢复策略

### 1. 字段级降级

```go
func fieldLevelFallback() {
    jsonData := `{
        "config": {
            "timeout": "invalid",
            "retry": 3,
            "debug": "true"
        }
    }`

    node := fxjson.FromString(jsonData)
    config := node.Get("config")

    // 字段级降级策略
    timeout := config.Get("timeout").IntOr(30)          // 降级到默认值 30
    retry := config.Get("retry").IntOr(3)               // 保持原值 3
    debug := config.Get("debug").StringOr("") == "true" // 字符串转布尔

    fmt.Printf("配置: timeout=%d, retry=%d, debug=%t\n", timeout, retry, debug)
}
```

### 2. 结构级验证

```go
type UserConfig struct {
    Name     string
    Age      int
    Email    string
    IsActive bool
}

func structLevelValidation() {
    jsonData := `{
        "name": "张三",
        "age": "invalid",
        "email": "zhang@example.com",
        "active": true
    }`

    node := fxjson.FromString(jsonData)

    // 收集所有错误
    var errors []string

    name := node.Get("name").StringOr("")
    if name == "" {
        errors = append(errors, "姓名不能为空")
    }

    age, err := node.Get("age").Int()
    if err != nil {
        errors = append(errors, "年龄格式错误")
        age = 0  // 使用默认值
    }

    email := node.Get("email").StringOr("")
    if !node.Get("email").IsValidEmail() {
        errors = append(errors, "邮箱格式无效")
    }

    isActive := node.Get("active").BoolOr(false)

    if len(errors) > 0 {
        fmt.Println("数据验证失败:")
        for i, err := range errors {
            fmt.Printf("%d. %s\n", i+1, err)
        }
        return
    }

    config := UserConfig{
        Name:     name,
        Age:      int(age),
        Email:    email,
        IsActive: isActive,
    }

    fmt.Printf("配置验证通过: %+v\n", config)
}
```

### 3. 批量处理的错误恢复

```go
func batchErrorRecovery() {
    jsonData := `{
        "users": [
            {"id": 1, "name": "张三", "age": 28},
            {"id": "invalid", "name": "李四", "age": 25},
            {"id": 3, "name": "王五", "age": "invalid"},
            {"id": 4, "name": "赵六", "age": 30}
        ]
    }`

    node := fxjson.FromString(jsonData)
    users := node.Get("users")

    type User struct {
        ID   int
        Name string
        Age  int
    }

    var validUsers []User
    var errorCount int

    users.ArrayForEach(func(index int, user fxjson.Node) bool {
        // 尝试解析每个用户
        id, idErr := user.Get("id").Int()
        age, ageErr := user.Get("age").Int()
        name := user.Get("name").StringOr("")

        // 跳过有错误的记录
        if idErr != nil || ageErr != nil {
            errorCount++
            fmt.Printf("⚠️ 用户 %d 数据有误,已跳过\n", index)
            return true  // 继续处理下一个
        }

        validUsers = append(validUsers, User{
            ID:   int(id),
            Name: name,
            Age:  int(age),
        })

        return true
    })

    fmt.Printf("\n处理完成:\n")
    fmt.Printf("  成功: %d 个\n", len(validUsers))
    fmt.Printf("  失败: %d 个\n", errorCount)
    fmt.Printf("  总计: %d 个\n", len(validUsers)+errorCount)
}
```

### 4. 链式错误处理

```go
func chainedErrorHandling() {
    jsonData := `{"data": {"user": {"profile": {"name": "张三"}}}}`
    node := fxjson.FromString(jsonData)

    // 链式访问 with 错误检查
    result := func() (string, error) {
        data := node.Get("data")
        if !data.Exists() {
            return "", fmt.Errorf("data 字段不存在")
        }

        user := data.Get("user")
        if !user.Exists() {
            return "", fmt.Errorf("user 字段不存在")
        }

        profile := user.Get("profile")
        if !profile.Exists() {
            return "", fmt.Errorf("profile 字段不存在")
        }

        name, err := profile.Get("name").String()
        if err != nil {
            return "", fmt.Errorf("name 字段错误: %w", err)
        }

        return name, nil
    }()

    if result, err := result; err != nil {
        fmt.Printf("错误: %v\n", err)
    } else {
        fmt.Printf("成功: %s\n", result)
    }
}
```

---

## 最佳实践

### 1. 选择合适的错误处理模式

```go
// ✅ 推荐: 非关键数据使用安全模式
displayName := user.Get("displayName").StringOr(user.Get("name").StringOr("匿名"))

// ✅ 推荐: 关键数据使用严格模式
orderID, err := order.Get("id").Int()
if err != nil {
    return fmt.Errorf("订单ID无效: %w", err)
}
```

### 2. 提前验证

```go
// ✅ 推荐: 提前检查存在性
if !node.GetPath("required.field").Exists() {
    return errors.New("缺少必需字段: required.field")
}

// ❌ 不推荐: 直接访问可能导致问题
value := node.GetPath("required.field").StringOr("")  // 可能返回空字符串
```

### 3. 错误日志记录

```go
func errorLogging() {
    node := fxjson.FromString(jsonData)

    age, err := node.Get("age").Int()
    if err != nil {
        // 记录详细错误信息
        if fxErr, ok := err.(*fxjson.FxJSONError); ok {
            log.Printf("错误类型: %v", fxErr.Type)
            log.Printf("错误消息: %s", fxErr.Message)
            log.Printf("错误位置: 行%d 列%d", fxErr.Position.Line, fxErr.Position.Column)
            log.Printf("错误上下文: %s", fxErr.Context)

            if fxErr.Cause != nil {
                log.Printf("原因: %v", fxErr.Cause)
            }
        }
    }
}
```

### 4. 优雅的错误处理

```go
func gracefulErrorHandling() error {
    node := fxjson.FromString(jsonData)

    // 使用自定义错误类型
    type ValidationError struct {
        Field   string
        Message string
    }

    var validationErrors []ValidationError

    // 验证所有字段
    if !node.Get("name").Exists() {
        validationErrors = append(validationErrors, ValidationError{
            Field:   "name",
            Message: "姓名字段缺失",
        })
    }

    if age := node.Get("age"); age.Exists() {
        if !age.InRange(0, 150) {
            validationErrors = append(validationErrors, ValidationError{
                Field:   "age",
                Message: "年龄超出合理范围",
            })
        }
    }

    // 汇总报告所有错误
    if len(validationErrors) > 0 {
        fmt.Println("数据验证失败:")
        for _, ve := range validationErrors {
            fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
        }
        return errors.New("数据验证失败")
    }

    return nil
}
```

### 5. 单元测试中的错误处理

```go
func TestJSONParsing(t *testing.T) {
    tests := []struct {
        name      string
        json      string
        wantErr   bool
        errorType fxjson.ErrorType
    }{
        {
            name:    "有效JSON",
            json:    `{"name": "张三"}`,
            wantErr: false,
        },
        {
            name:      "无效JSON",
            json:      `{"name": `,
            wantErr:   true,
            errorType: fxjson.ErrorParse,
        },
        {
            name:      "类型错误",
            json:      `{"age": "not-a-number"}`,
            wantErr:   true,
            errorType: fxjson.ErrorTypeMismatch,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            node := fxjson.FromString(tt.json)

            if tt.wantErr {
                if node.Exists() {
                    _, err := node.Get("age").Int()
                    if err == nil {
                        t.Error("期望错误,但没有发生")
                    }

                    if fxErr, ok := err.(*fxjson.FxJSONError); ok {
                        if fxErr.Type != tt.errorType {
                            t.Errorf("错误类型 = %v, 期望 %v", fxErr.Type, tt.errorType)
                        }
                    }
                }
            }
        })
    }
}
```

完善的错误处理是构建健壮应用的基础。FxJSON 的错误处理机制既简单易用,又足够灵活,能够满足从简单脚本到复杂企业应用的各种需求。
