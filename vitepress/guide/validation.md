# 数据验证

FxJSON 提供了强大的数据验证功能，包括内置验证器和自定义验证规则，帮助您确保数据的完整性和正确性。

> **注意：** FxJSON 提供了内置的 `node.Validate(validator)` 方法（详见 [API 文档](/api/)），本指南中的示例展示了如何实现自定义验证逻辑以满足特定需求。

## 内置验证器

### 基础类型验证

```go
func basicValidation() {
    jsonData := `{
        "email": "user@example.com",
        "website": "https://example.com",
        "ip": "192.168.1.1",
        "uuid": "123e4567-e89b-12d3-a456-426614174000",
        "phone": "+86-13800138000",
        "age": 25,
        "score": 95.5
    }`

    node := fxjson.FromBytes([]byte(jsonData))

    // 邮箱验证
    if node.Get("email").IsValidEmail() {
        fmt.Println("✅ 邮箱格式正确")
    } else {
        fmt.Println("❌ 邮箱格式错误")
    }

    // URL验证
    if node.Get("website").IsValidURL() {
        fmt.Println("✅ URL格式正确")
    }

    // IP地址验证
    if node.Get("ip").IsValidIP() {
        fmt.Println("✅ IP地址格式正确")
    }

    // UUID验证
    if node.Get("uuid").IsValidUUID() {
        fmt.Println("✅ UUID格式正确")
    }

    // 数字范围验证
    age := node.Get("age").IntOr(0)
    if age >= 0 && age <= 150 {
        fmt.Println("✅ 年龄在有效范围内")
    }
}
```

### 字符串验证

```go
func stringValidation() {
    jsonData := `{
        "username": "user123",
        "password": "SecurePass123!",
        "code": "ABC123",
        "description": "这是一个描述文本"
    }`

    node := fxjson.FromBytes([]byte(jsonData))

    // 字符串长度验证
    username := node.Get("username").StringOr("")
    if len(username) >= 3 && len(username) <= 20 {
        fmt.Println("✅ 用户名长度符合要求")
    }

    // 密码强度验证
    password := node.Get("password").StringOr("")
    if validatePasswordStrength(password) {
        fmt.Println("✅ 密码强度符合要求")
    }

    // 字符模式验证
    code := node.Get("code").StringOr("")
    if validateAlphanumeric(code) {
        fmt.Println("✅ 验证码格式正确")
    }
}

func validatePasswordStrength(password string) bool {
    if len(password) < 8 {
        return false
    }
    
    hasUpper := false
    hasLower := false
    hasDigit := false
    hasSpecial := false
    
    for _, char := range password {
        switch {
        case char >= 'A' && char <= 'Z':
            hasUpper = true
        case char >= 'a' && char <= 'z':
            hasLower = true
        case char >= '0' && char <= '9':
            hasDigit = true
        default:
            hasSpecial = true
        }
    }
    
    return hasUpper && hasLower && hasDigit && hasSpecial
}

func validateAlphanumeric(s string) bool {
    for _, char := range s {
        if !((char >= 'A' && char <= 'Z') || 
             (char >= 'a' && char <= 'z') || 
             (char >= '0' && char <= '9')) {
            return false
        }
    }
    return len(s) > 0
}
```

## 自定义验证规则

### 验证规则定义

```go
type ValidationRule struct {
    Required     bool                          `json:"required"`
    Type         string                        `json:"type"`
    MinLength    int                           `json:"min_length"`
    MaxLength    int                           `json:"max_length"`
    Min          float64                       `json:"min"`
    Max          float64                       `json:"max"`
    Pattern      string                        `json:"pattern"`
    Default      interface{}                   `json:"default"`
    Enum         []interface{}                 `json:"enum"`
    CustomCheck  func(interface{}) bool        `json:"-"`
    ErrorMessage string                        `json:"error_message"`
}

type DataValidator struct {
    Rules map[string]ValidationRule `json:"rules"`
}

func customValidation() {
    // 定义验证规则
    validator := &DataValidator{
        Rules: map[string]ValidationRule{
            "name": {
                Required:     true,
                Type:         "string",
                MinLength:    2,
                MaxLength:    50,
                ErrorMessage: "姓名必须是2-50个字符",
            },
            "email": {
                Required:     true,
                Type:         "string",
                Pattern:      `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
                ErrorMessage: "请提供有效的邮箱地址",
            },
            "age": {
                Required:     true,
                Type:         "number",
                Min:          0,
                Max:          150,
                ErrorMessage: "年龄必须在0-150之间",
            },
            "gender": {
                Required:     false,
                Type:         "string",
                Enum:         []interface{}{"男", "女", "其他"},
                Default:      "未指定",
                ErrorMessage: "性别必须是：男、女、其他",
            },
            "phone": {
                Required: false,
                Type:     "string",
                CustomCheck: func(value interface{}) bool {
                    if str, ok := value.(string); ok {
                        return validateChinesePhone(str)
                    }
                    return false
                },
                ErrorMessage: "请提供有效的中国手机号码",
            },
        },
    }

    // 测试数据
    testData := `{
        "name": "张三",
        "email": "zhangsan@example.com",
        "age": 28,
        "gender": "男",
        "phone": "13800138000"
    }`

    node := fxjson.FromBytes([]byte(testData)
    
    // 执行验证（使用自定义验证器）
    errors := validator.Validate(node)
    if len(errors) == 0 {
        fmt.Println("✅ 所有数据验证通过")
    } else {
        fmt.Println("❌ 验证失败：")
        for _, err := range errors {
            fmt.Printf("  - %s\n", err)
        }
    }
}

func validateChinesePhone(phone string) bool {
    if len(phone) != 11 {
        return false
    }
    
    if phone[0] != '1' {
        return false
    }
    
    validSecondDigits := []byte{'3', '4', '5', '6', '7', '8', '9'}
    secondDigit := phone[1]
    
    for _, valid := range validSecondDigits {
        if secondDigit == valid {
            for i := 2; i < 11; i++ {
                if phone[i] < '0' || phone[i] > '9' {
                    return false
                }
            }
            return true
        }
    }
    
    return false
}

// DataValidator 的 Validate 方法实现
func (dv *DataValidator) Validate(node fxjson.Node) []string {
    var errors []string
    
    for fieldName, rule := range dv.Rules {
        fieldNode := node.Get(fieldName)
        
        // 检查必需字段
        if rule.Required && !fieldNode.Exists() {
            errors = append(errors, fmt.Sprintf("%s: 字段是必需的", fieldName))
            continue
        }
        
        // 如果字段不存在且不是必需的，使用默认值
        if !fieldNode.Exists() {
            if rule.Default != nil {
                continue // 有默认值，跳过验证
            }
            continue // 可选字段且无默认值，跳过验证
        }
        
        // 类型验证
        if !dv.validateType(fieldNode, rule.Type) {
            errors = append(errors, fmt.Sprintf("%s: %s", fieldName, rule.ErrorMessage))
            continue
        }
        
        // 具体验证
        if err := dv.validateField(fieldNode, rule, fieldName); err != nil {
            errors = append(errors, err.Error())
        }
    }
    
    return errors
}

func (dv *DataValidator) validateType(node fxjson.Node, expectedType string) bool {
    switch expectedType {
    case "string":
        return node.IsString()
    case "number":
        return node.IsNumber()
    case "boolean":
        return node.IsBool()
    case "array":
        return node.IsArray()
    case "object":
        return node.IsObject()
    default:
        return true
    }
}

func (dv *DataValidator) validateField(node fxjson.Node, rule ValidationRule, fieldName string) error {
    // 字符串验证
    if rule.Type == "string" {
        str := node.StringOr("")
        
        if rule.MinLength > 0 && len(str) < rule.MinLength {
            return fmt.Errorf("%s: %s", fieldName, rule.ErrorMessage)
        }
        
        if rule.MaxLength > 0 && len(str) > rule.MaxLength {
            return fmt.Errorf("%s: %s", fieldName, rule.ErrorMessage)
        }
        
        if rule.Pattern != "" {
            matched, _ := regexp.MatchString(rule.Pattern, str)
            if !matched {
                return fmt.Errorf("%s: %s", fieldName, rule.ErrorMessage)
            }
        }
        
        if len(rule.Enum) > 0 {
            valid := false
            for _, enum := range rule.Enum {
                if str == enum {
                    valid = true
                    break
                }
            }
            if !valid {
                return fmt.Errorf("%s: %s", fieldName, rule.ErrorMessage)
            }
        }
    }
    
    // 数字验证
    if rule.Type == "number" {
        num := node.FloatOr(0)
        
        if rule.Min != 0 && num < rule.Min {
            return fmt.Errorf("%s: %s", fieldName, rule.ErrorMessage)
        }
        
        if rule.Max != 0 && num > rule.Max {
            return fmt.Errorf("%s: %s", fieldName, rule.ErrorMessage)
        }
    }
    
    // 自定义验证
    if rule.CustomCheck != nil {
        var value interface{}
        switch rule.Type {
        case "string":
            value = node.StringOr("")
        case "number":
            value = node.FloatOr(0)
        case "boolean":
            value = node.BoolOr(false)
        default:
            value = node.StringOr("")
        }
        
        if !rule.CustomCheck(value) {
            return fmt.Errorf("%s: %s", fieldName, rule.ErrorMessage)
        }
    }
    
    return nil
}
```

## 复杂数据结构验证

### 嵌套对象验证

```go
func nestedValidation() {
    userSchema := &DataValidator{
        Rules: map[string]ValidationRule{
            "personal.name": {
                Required:  true,
                Type:      "string",
                MinLength: 2,
                MaxLength: 50,
            },
            "personal.age": {
                Required: true,
                Type:     "number",
                Min:      0,
                Max:      150,
            },
            "contact.email": {
                Required: true,
                Type:     "string",
                Pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
            },
            "contact.phone": {
                Required: false,
                Type:     "string",
                CustomCheck: func(value interface{}) bool {
                    if str, ok := value.(string); ok {
                        return validateChinesePhone(str)
                    }
                    return false
                },
            },
        },
    }

    testData := `{
        "personal": {
            "name": "张三",
            "age": 28
        },
        "contact": {
            "email": "zhangsan@example.com",
            "phone": "13800138000"
        }
    }`

    node := fxjson.FromBytes([]byte(testData)
    
    errors := validateNested(node, userSchema)
    if len(errors) == 0 {
        fmt.Println("✅ 嵌套数据验证通过")
    } else {
        fmt.Println("❌ 嵌套数据验证失败：")
        for _, err := range errors {
            fmt.Printf("  - %s\n", err)
        }
    }
}

func validateNested(node fxjson.Node, validator *DataValidator) []string {
    var errors []string
    
    for fieldPath, rule := range validator.Rules {
        fieldNode := node.GetPath(fieldPath)
        
        if rule.Required && !fieldNode.Exists() {
            errors = append(errors, fmt.Sprintf("%s: 字段是必需的", fieldPath))
            continue
        }
        
        if !fieldNode.Exists() {
            continue
        }
        
        if !validator.validateType(fieldNode, rule.Type) {
            errors = append(errors, fmt.Sprintf("%s: 类型不匹配", fieldPath))
            continue
        }
        
        if err := validator.validateField(fieldNode, rule, fieldPath); err != nil {
            errors = append(errors, err.Error())
        }
    }
    
    return errors
}
```

### 数组验证

```go
func arrayValidation() {
    arraySchema := &DataValidator{
        Rules: map[string]ValidationRule{
            "users": {
                Required: true,
                Type:     "array",
                CustomCheck: func(value interface{}) bool {
                    // 验证数组长度
                    return true
                },
            },
        },
    }

    itemSchema := &DataValidator{
        Rules: map[string]ValidationRule{
            "name": {
                Required:  true,
                Type:      "string",
                MinLength: 2,
                MaxLength: 50,
            },
            "email": {
                Required: true,
                Type:     "string",
                Pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
            },
        },
    }

    testData := `{
        "users": [
            {"name": "张三", "email": "zhang@example.com"},
            {"name": "李四", "email": "li@example.com"},
            {"name": "王五", "email": "invalid-email"}
        ]
    }`

    node := fxjson.FromBytes([]byte(testData)
    
    // 验证数组本身
    errors := arraySchema.Validate(node)
    
    // 验证数组元素
    users := node.Get("users")
    if users.IsArray() {
        users.ArrayForEach(func(index int, userNode fxjson.Node) bool {
            itemErrors := itemSchema.Validate(userNode)
            for _, err := range itemErrors {
                errors = append(errors, fmt.Sprintf("users[%d].%s", index, err))
            }
            return true
        })
    }
    
    if len(errors) == 0 {
        fmt.Println("✅ 数组数据验证通过")
    } else {
        fmt.Println("❌ 数组数据验证失败：")
        for _, err := range errors {
            fmt.Printf("  - %s\n", err)
        }
    }
}
```

## 数据清理和转换

### 数据清理

```go
func dataSanitization() {
    dirtyData := `{
        "name": "  张三  ",
        "email": "ZHANGSAN@EXAMPLE.COM",
        "phone": "138-0013-8000",
        "description": "<script>alert('xss')</script>正常内容",
        "tags": ["  Go  ", "JSON", "  Performance  "]
    }`

    node := fxjson.FromBytes([]byte(dirtyData)
    
    // 创建清理后的数据
    cleaned := map[string]interface{}{
        "name":        strings.TrimSpace(node.Get("name").StringOr("")),
        "email":       strings.ToLower(node.Get("email").StringOr("")),
        "phone":       sanitizePhone(node.Get("phone").StringOr("")),
        "description": sanitizeHTML(node.Get("description").StringOr("")),
        "tags":        sanitizeTags(node.Get("tags")),
    }

    // 输出清理结果
    for key, value := range cleaned {
        fmt.Printf("%s: %v\n", key, value)
    }
}

func sanitizePhone(phone string) string {
    // 移除所有非数字字符
    result := ""
    for _, char := range phone {
        if char >= '0' && char <= '9' {
            result += string(char)
        }
    }
    return result
}

func sanitizeHTML(input string) string {
    // 简单的HTML标签移除（实际使用中建议使用专门的库）
    re := regexp.MustCompile(`<[^>]*>`)
    return re.ReplaceAllString(input, "")
}

func sanitizeTags(tagsNode fxjson.Node) []string {
    var tags []string
    if tagsNode.IsArray() {
        tagsNode.ArrayForEach(func(index int, tag fxjson.Node) bool {
            cleaned := strings.TrimSpace(tag.StringOr(""))
            if cleaned != "" {
                tags = append(tags, cleaned)
            }
            return true
        })
    }
    return tags
}
```

### 数据转换和标准化

```go
func dataTransformation() {
    rawData := `{
        "user_name": "zhang_san",
        "user_email": "ZHANG@EXAMPLE.COM",
        "user_age": "28",
        "is_active": "true",
        "created_date": "2024-01-15"
    }`

    node := fxjson.FromBytes([]byte(rawData)
    
    // 转换为标准格式
    standardized := map[string]interface{}{
        "name":      toCamelCase(node.Get("user_name").StringOr("")),
        "email":     strings.ToLower(node.Get("user_email").StringOr("")),
        "age":       node.Get("user_age").IntOr(0),
        "isActive":  node.Get("is_active").StringOr("") == "true",
        "createdAt": formatDate(node.Get("created_date").StringOr("")),
    }

    fmt.Println("标准化后的数据：")
    for key, value := range standardized {
        fmt.Printf("%s: %v (%T)\n", key, value, value)
    }
}

func toCamelCase(snake_case string) string {
    parts := strings.Split(snake_case, "_")
    if len(parts) <= 1 {
        return snake_case
    }
    
    result := parts[0]
    for i := 1; i < len(parts); i++ {
        if len(parts[i]) > 0 {
            result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
        }
    }
    return result
}

func formatDate(dateStr string) string {
    // 简单的日期格式转换
    if len(dateStr) == 10 && strings.Count(dateStr, "-") == 2 {
        return dateStr + "T00:00:00Z"
    }
    return dateStr
}
```

## 批量验证

```go
func batchValidation() {
    // 批量验证多个用户数据
    usersData := `[
        {"name": "张三", "email": "zhang@example.com", "age": 28},
        {"name": "李四", "email": "li@example.com", "age": 25},
        {"name": "", "email": "invalid", "age": 200},
        {"name": "王五", "email": "wang@example.com", "age": 30}
    ]`

    schema := &DataValidator{
        Rules: map[string]ValidationRule{
            "name": {
                Required:  true,
                Type:      "string",
                MinLength: 1,
                MaxLength: 50,
            },
            "email": {
                Required: true,
                Type:     "string",
                Pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
            },
            "age": {
                Required: true,
                Type:     "number",
                Min:      0,
                Max:      150,
            },
        },
    }

    node := fxjson.FromBytes([]byte(usersData)
    
    var allErrors []string
    validCount := 0
    totalCount := 0

    node.ArrayForEach(func(index int, userNode fxjson.Node) bool {
        totalCount++
        errors := schema.Validate(userNode)
        
        if len(errors) == 0 {
            validCount++
            fmt.Printf("用户 %d: ✅ 验证通过\n", index+1)
        } else {
            fmt.Printf("用户 %d: ❌ 验证失败\n", index+1)
            for _, err := range errors {
                errorMsg := fmt.Sprintf("  用户%d - %s", index+1, err)
                fmt.Println(errorMsg)
                allErrors = append(allErrors, errorMsg)
            }
        }
        return true
    })

    fmt.Printf("\n批量验证结果: %d/%d 通过\n", validCount, totalCount)
    if len(allErrors) > 0 {
        fmt.Printf("总共 %d 个错误\n", len(allErrors))
    }
}
```

## 性能优化

```go
func optimizedValidation() {
    // 对于大量数据，使用并发验证
    largeDataset := generateLargeUserDataset(10000)
    
    schema := &DataValidator{
        Rules: map[string]ValidationRule{
            "name":  {Required: true, Type: "string", MinLength: 1},
            "email": {Required: true, Type: "string", Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`},
            "age":   {Required: true, Type: "number", Min: 0, Max: 150},
        },
    }

    start := time.Now()
    
    // 并发验证
    var wg sync.WaitGroup
    errorsChan := make(chan string, 100)
    validCount := int64(0)
    
    node := fxjson.FromBytes([]byte(largeDataset)
    node.ArrayForEach(func(index int, userNode fxjson.Node) bool {
        wg.Add(1)
        go func(idx int, user fxjson.Node) {
            defer wg.Done()
            
            errors := schema.Validate(user)
            if len(errors) == 0 {
                atomic.AddInt64(&validCount, 1)
            } else {
                for _, err := range errors {
                    select {
                    case errorsChan <- fmt.Sprintf("用户%d: %s", idx+1, err):
                    default:
                        // 错误通道满，跳过
                    }
                }
            }
        }(index, userNode)
        
        return true
    })
    
    wg.Wait()
    close(errorsChan)
    
    duration := time.Since(start)
    
    var errorCount int
    for range errorsChan {
        errorCount++
    }
    
    fmt.Printf("并发验证完成: 耗时 %v\n", duration)
    fmt.Printf("有效数据: %d, 错误数据: %d\n", validCount, errorCount)
}

func generateLargeUserDataset(count int) string {
    var users []string
    for i := 0; i < count; i++ {
        user := fmt.Sprintf(`{"name": "用户%d", "email": "user%d@example.com", "age": %d}`,
            i+1, i+1, 20+i%60)
        users = append(users, user)
    }
    return "[" + strings.Join(users, ",") + "]"
}
```

数据验证是确保应用程序数据质量的重要环节。FxJSON 的验证功能帮助您在数据处理的早期阶段发现和修正问题，提高应用程序的可靠性和安全性。