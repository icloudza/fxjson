# 工具函数

FxJSON 提供了丰富的工具函数，涵盖数据转换、格式化、验证等多个方面，大大简化了 JSON 数据处理的复杂性。

## 数据转换工具

### 类型转换函数

```go
func typeConversionUtils() {
    jsonData := `{
        "numbers": [1, 2, 3, 4, 5],
        "strings": ["apple", "banana", "cherry"],
        "mixed": [1, "two", 3.14, true, null],
        "user": {
            "id": "123",
            "name": "张三",
            "active": "true",
            "score": "95.5"
        }
    }`

    node := fxjson.FromBytes([]byte(jsonData))

    // 数组转换为字符串切片
    var strings []string
    node.Get("strings").ArrayForEach(func(index int, value fxjson.Node) bool {
        strings = append(strings, value.StringOr(""))
        return true
    })
    fmt.Printf("字符串数组: %v\n", strings)

    // 数组转换为整数切片
    var numbers []int64
    node.Get("numbers").ArrayForEach(func(index int, value fxjson.Node) bool {
        numbers = append(numbers, value.IntOr(0))
        return true
    })
    fmt.Printf("整数数组: %v\n", numbers)

    // 数组转换为浮点数切片
    var floats []float64
    node.Get("numbers").ArrayForEach(func(index int, value fxjson.Node) bool {
        floats = append(floats, value.FloatOr(0))
        return true
    })
    fmt.Printf("浮点数组: %v\n", floats)

    // 对象转换为字符串映射
    userMap := make(map[string]string)
    node.Get("user").ForEach(func(key string, value fxjson.Node) bool {
        userMap[key] = value.StringOr("")
        return true
    })
    fmt.Printf("用户映射: %v\n", userMap)

    // 安全类型转换
    userID := node.GetPath("user.id").IntOr(0)
    userName := node.GetPath("user.name").StringOr("未知")
    isActive := node.GetPath("user.active").StringOr("") == "true"
    score := node.GetPath("user.score").FloatOr(0.0)

    fmt.Printf("用户信息: ID=%d, 姓名=%s, 活跃=%t, 分数=%.1f\n", 
        userID, userName, isActive, score)
}
```

### 批量获取工具

```go
func batchAccessUtils() {
    userData := `{
        "profile": {
            "name": "张三",
            "age": 30,
            "email": "zhangsan@example.com"
        },
        "settings": {
            "theme": "dark",
            "language": "zh-CN",
            "notifications": true
        },
        "stats": {
            "loginCount": 42,
            "lastLogin": "2024-01-15T10:30:00Z"
        }
    }`

    node := fxjson.FromBytes([]byte(userData))

    // 批量获取多个路径
    paths := []string{
        "profile.name",
        "profile.age", 
        "profile.email",
        "settings.theme",
        "settings.language",
        "stats.loginCount",
    }

    fmt.Println("批量获取结果:")
    for _, path := range paths {
        value := node.GetPath(path)
        fmt.Printf("%s: %s\n", path, value.StringOr("N/A"))
    }

    // 检查是否存在路径
    profilePaths := []string{"profile.phone", "profile.mobile", "profile.contact"}
    hasContact := false
    for _, path := range profilePaths {
        if node.GetPath(path).Exists() {
            hasContact = true
            break
        }
    }
    fmt.Printf("是否有联系方式: %t\n", hasContact)

    // 检查是否存在所有路径
    requiredPaths := []string{"profile.name", "profile.email"}
    hasRequired := true
    for _, path := range requiredPaths {
        if !node.GetPath(path).Exists() {
            hasRequired = false
            break
        }
    }
    fmt.Printf("是否有必要信息: %t\n", hasRequired)

    // 安全的批量获取
    profileData := extractProfile(node)
    fmt.Printf("用户资料: %+v\n", profileData)
}

type UserProfile struct {
    Name         string `json:"name"`
    Age          int    `json:"age"`
    Email        string `json:"email"`
    Theme        string `json:"theme"`
    Language     string `json:"language"`
    LoginCount   int    `json:"loginCount"`
    Notifications bool  `json:"notifications"`
}

func extractProfile(node fxjson.Node) UserProfile {
    return UserProfile{
        Name:          node.GetPath("profile.name").StringOr(""),
        Age:           int(node.GetPath("profile.age").IntOr(0)),
        Email:         node.GetPath("profile.email").StringOr(""),
        Theme:         node.GetPath("settings.theme").StringOr("light"),
        Language:      node.GetPath("settings.language").StringOr("en"),
        LoginCount:    int(node.GetPath("stats.loginCount").IntOr(0)),
        Notifications: node.GetPath("settings.notifications").BoolOr(true),
    }
}
```

## JSON 格式化工具

### 美化和压缩

```go
func jsonFormattingUtils() {
    compactJSON := `{"name":"张三","age":30,"address":{"city":"北京","district":"朝阳区"},"hobbies":["阅读","编程","旅行"]}`

    // 使用标准库进行JSON美化
    var jsonData interface{}
    json.Unmarshal([]byte(compactJSON), &jsonData)
    
    // JSON 美化
    prettyBytes, _ := json.MarshalIndent(jsonData, "", "  ")
    fmt.Printf("美化后的 JSON:\n%s\n", prettyBytes)

    // 自定义缩进美化
    customPrettyBytes, _ := json.MarshalIndent(jsonData, "", "    ") // 4个空格
    fmt.Printf("自定义缩进美化:\n%s\n", customPrettyBytes)

    // JSON 压缩（移除空白字符）
    messyJSON := `{
        "name" : "张三" ,
        "age"  : 30 ,
        "address" : {
            "city" : "北京" ,
            "district" : "朝阳区"
        }
    }`
    
    compactedBytes, _ := json.Marshal(jsonData)
    fmt.Printf("压缩后的 JSON: %s\n", compactedBytes)

    // 验证 JSON 格式
    if isValidJSON([]byte(compactJSON)) {
        fmt.Println("JSON 格式有效")
    }
}

// JSON 有效性验证
func isValidJSON(data []byte) bool {
    node := fxjson.FromBytes(data)
    return node.Exists()
}

// 使用FxJSON进行美化
func prettyJSONWithFxJSON(data []byte) []byte {
    node := fxjson.FromBytes(data)
    return []byte(node.String())
}
```

### JSON 差异对比

```go
func jsonDiffUtils() {
    json1 := `{
        "name": "张三",
        "age": 30,
        "city": "北京",
        "hobbies": ["阅读", "编程"]
    }`

    json2 := `{
        "name": "张三",
        "age": 31,
        "city": "上海",
        "hobbies": ["阅读", "编程", "旅行"],
        "email": "zhangsan@example.com"
    }`

    node1 := fxjson.FromBytes([]byte(json1))
    node2 := fxjson.FromBytes([]byte(json2))

    // 比较并找出差异
    differences := compareJSON(node1, node2)
    
    fmt.Println("JSON 差异:")
    for _, diff := range differences {
        fmt.Printf("- %s\n", diff)
    }
}

type JSONDiff struct {
    Path      string      `json:"path"`
    Type      string      `json:"type"` // added, removed, modified
    OldValue  interface{} `json:"old_value,omitempty"`
    NewValue  interface{} `json:"new_value,omitempty"`
}

func compareJSON(node1, node2 fxjson.Node) []string {
    var differences []string
    
    // 比较对象的键
    if node1.IsObject() && node2.IsObject() {
        // 收集所有键
        keys1 := make(map[string]bool)
        keys2 := make(map[string]bool)
        allKeys := make(map[string]bool)
        
        node1.ForEach(func(key string, value fxjson.Node) bool {
            keys1[key] = true
            allKeys[key] = true
            return true
        })
        
        node2.ForEach(func(key string, value fxjson.Node) bool {
            keys2[key] = true
            allKeys[key] = true
            return true
        })
        
        for key := range allKeys {
            has1 := keys1[key]
            has2 := keys2[key]
            
            if has1 && !has2 {
                differences = append(differences, fmt.Sprintf("键 '%s' 被删除", key))
            } else if !has1 && has2 {
                differences = append(differences, fmt.Sprintf("键 '%s' 被添加: %s", 
                    key, node2.Get(key).StringOr("")))
            } else if has1 && has2 {
                value1 := node1.Get(key).StringOr("")
                value2 := node2.Get(key).StringOr("")
                if value1 != value2 {
                    differences = append(differences, fmt.Sprintf("键 '%s' 值变化: '%s' -> '%s'", 
                        key, value1, value2))
                }
            }
        }
    }
    
    return differences
}
```

## 数据验证工具

### 内置验证器扩展

```go
func extendedValidationUtils() {
    testData := `{
        "user": {
            "email": "user@example.com",
            "phone": "+86-138-0013-8000",
            "website": "https://example.com",
            "ip": "192.168.1.1",
            "uuid": "123e4567-e89b-12d3-a456-426614174000",
            "creditCard": "4532-1234-5678-9012",
            "idCard": "11010519491231002X"
        }
    }`

    node := fxjson.FromBytes([]byte(testData))
    user := node.Get("user")

    // 扩展验证函数
    validations := map[string]func() bool{
        "邮箱":   func() bool { return user.Get("email").IsValidEmail() },
        "电话":   func() bool { return isValidPhone(user.Get("phone").StringOr("")) },
        "网站":   func() bool { return user.Get("website").IsValidURL() },
        "IP地址": func() bool { return user.Get("ip").IsValidIP() },
        "UUID":  func() bool { return isValidUUID(user.Get("uuid").StringOr("")) },
        "信用卡":  func() bool { return isValidCreditCard(user.Get("creditCard").StringOr("")) },
        "身份证":  func() bool { return isValidChineseID(user.Get("idCard").StringOr("")) },
    }

    fmt.Println("数据验证结果:")
    for field, validator := range validations {
        status := "❌"
        if validator() {
            status = "✅"
        }
        fmt.Printf("%s %s\n", status, field)
    }
}

func isValidPhone(phone string) bool {
    // 支持多种电话号码格式
    patterns := []string{
        `^1[3-9]\d{9}$`,                    // 中国手机号
        `^\+86-1[3-9]\d-\d{4}-\d{4}$`,     // 带国家代码的中国手机号
        `^\(\d{3}\) \d{3}-\d{4}$`,         // 美国格式
        `^\d{3}-\d{3}-\d{4}$`,             // 简单格式
    }
    
    phone = strings.ReplaceAll(phone, " ", "")
    phone = strings.ReplaceAll(phone, "-", "")
    
    for _, pattern := range patterns {
        if matched, _ := regexp.MatchString(pattern, phone); matched {
            return true
        }
    }
    
    return false
}

func isValidUUID(uuid string) bool {
    pattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
    matched, _ := regexp.MatchString(pattern, uuid)
    return matched
}

func isValidCreditCard(cardNumber string) bool {
    // 移除空格和连字符
    cardNumber = strings.ReplaceAll(cardNumber, " ", "")
    cardNumber = strings.ReplaceAll(cardNumber, "-", "")
    
    // 检查长度
    if len(cardNumber) < 13 || len(cardNumber) > 19 {
        return false
    }
    
    // Luhn算法验证
    return luhnCheck(cardNumber)
}

func luhnCheck(cardNumber string) bool {
    var sum int
    alternate := false
    
    for i := len(cardNumber) - 1; i >= 0; i-- {
        digit := int(cardNumber[i] - '0')
        
        if alternate {
            digit *= 2
            if digit > 9 {
                digit = digit%10 + digit/10
            }
        }
        
        sum += digit
        alternate = !alternate
    }
    
    return sum%10 == 0
}

func isValidChineseID(id string) bool {
    if len(id) != 18 {
        return false
    }
    
    // 前17位应该都是数字
    for i := 0; i < 17; i++ {
        if id[i] < '0' || id[i] > '9' {
            return false
        }
    }
    
    // 最后一位可以是数字或X
    last := id[17]
    if !(last >= '0' && last <= '9') && last != 'X' {
        return false
    }
    
    // 可以进一步验证校验位，这里简化处理
    return true
}
```

## 调试和日志工具

### 调试助手

```go
func debuggingUtils() {
    complexJSON := `{
        "level1": {
            "level2": {
                "level3": {
                    "data": "深层数据",
                    "array": [1, 2, {"nested": "value"}]
                }
            }
        },
        "users": [
            {"id": 1, "name": "用户1"},
            {"id": 2, "name": "用户2"}
        ]
    }`

    node := fxjson.FromBytes([]byte(complexJSON))
    
    // 调试信息输出
    debugInfo := analyzeJSONStructure(node)
    fmt.Printf("JSON 结构分析:\n%s\n", debugInfo)
    
    // 路径发现
    allPaths := discoverAllPaths(node, "")
    fmt.Println("所有可用路径:")
    for _, path := range allPaths {
        fmt.Printf("- %s\n", path)
    }
    
    // 类型分析
    typeAnalysis := analyzeTypes(node)
    fmt.Println("类型分析:")
    for typeName, count := range typeAnalysis {
        fmt.Printf("- %s: %d 个\n", typeName, count)
    }
}

func analyzeJSONStructure(node fxjson.Node) string {
    var analysis strings.Builder
    
    analysis.WriteString(fmt.Sprintf("根节点类型: %s\n", getNodeTypeName(node)))
    
    if node.IsObject() {
        keyCount := 0
        var keys []string
        node.ForEach(func(key string, value fxjson.Node) bool {
            keyCount++
            keys = append(keys, key)
            return true
        })
        analysis.WriteString(fmt.Sprintf("对象键数量: %d\n", keyCount))
        analysis.WriteString("键列表: " + strings.Join(keys, ", ") + "\n")
    } else if node.IsArray() {
        length := node.Len()
        analysis.WriteString(fmt.Sprintf("数组长度: %d\n", length))
    }
    
    return analysis.String()
}

func discoverAllPaths(node fxjson.Node, prefix string) []string {
    var paths []string
    
    if node.IsObject() {
        node.ForEach(func(key string, value fxjson.Node) bool {
            currentPath := key
            if prefix != "" {
                currentPath = prefix + "." + key
            }
            
            paths = append(paths, currentPath)
            
            // 递归发现嵌套路径
            childPaths := discoverAllPaths(value, currentPath)
            paths = append(paths, childPaths...)
            
            return true
        })
    } else if node.IsArray() {
        node.ArrayForEach(func(index int, value fxjson.Node) bool {
            currentPath := fmt.Sprintf("%d", index)
            if prefix != "" {
                currentPath = prefix + "." + currentPath
            }
            
            paths = append(paths, currentPath)
            
            // 递归发现嵌套路径
            childPaths := discoverAllPaths(value, currentPath)
            paths = append(paths, childPaths...)
            
            return true
        })
    }
    
    return paths
}

func analyzeTypes(node fxjson.Node) map[string]int {
    types := make(map[string]int)
    analyzeTypesRecursive(node, types)
    return types
}

func analyzeTypesRecursive(node fxjson.Node, types map[string]int) {
    typeName := getNodeTypeName(node)
    types[typeName]++
    
    if node.IsObject() {
        node.ForEach(func(key string, value fxjson.Node) bool {
            analyzeTypesRecursive(value, types)
            return true
        })
    } else if node.IsArray() {
        node.ArrayForEach(func(index int, value fxjson.Node) bool {
            analyzeTypesRecursive(value, types)
            return true
        })
    }
}

func getNodeTypeName(node fxjson.Node) string {
    switch {
    case node.IsObject():
        return "对象"
    case node.IsArray():
        return "数组"
    case node.IsString():
        return "字符串"
    case node.IsNumber():
        return "数字"
    case node.IsBool():
        return "布尔值"
    case node.IsNull():
        return "null"
    default:
        return "未知"
    }
}
```

### 错误诊断工具

```go
func errorDiagnosticUtils() {
    // 测试各种可能出错的JSON
    testCases := []struct {
        name string
        json string
    }{
        {"正常JSON", `{"name": "test", "value": 123}`},
        {"无效JSON", `{"name": "test", "value":}`},
        {"不完整JSON", `{"name": "test"`},
        {"类型错误", `{"name": 123, "value": "should_be_number"}`},
        {"空JSON", ``},
        {"只有空白", `   `},
    }

    diagnostics := &JSONDiagnostics{}
    
    for _, testCase := range testCases {
        result := diagnostics.Diagnose(testCase.name, testCase.json)
        fmt.Printf("诊断 '%s': %s\n", testCase.name, result.Summary)
        
        if len(result.Issues) > 0 {
            fmt.Println("  问题:")
            for _, issue := range result.Issues {
                fmt.Printf("    - %s\n", issue)
            }
        }
        fmt.Println()
    }
}

type JSONDiagnostics struct{}

type DiagnosticResult struct {
    Name    string   `json:"name"`
    Valid   bool     `json:"valid"`
    Summary string   `json:"summary"`
    Issues  []string `json:"issues"`
}

func (jd *JSONDiagnostics) Diagnose(name, jsonStr string) DiagnosticResult {
    result := DiagnosticResult{
        Name:   name,
        Issues: []string{},
    }
    
    if strings.TrimSpace(jsonStr) == "" {
        result.Valid = false
        result.Summary = "JSON为空"
        result.Issues = append(result.Issues, "输入为空或只包含空白字符")
        return result
    }
    
    node := fxjson.FromBytes([]byte(jsonStr))
    
    if !node.Exists() {
        result.Valid = false
        result.Summary = "JSON格式无效"
        result.Issues = append(result.Issues, "JSON语法错误")
    } else {
        result.Valid = true
        result.Summary = "JSON格式有效"
        
        // 进一步分析
        if node.IsObject() {
            keyCount := 0
            node.ForEach(func(key string, value fxjson.Node) bool {
                keyCount++
                return true
            })
            if keyCount == 0 {
                result.Issues = append(result.Issues, "对象为空")
            }
        } else if node.IsArray() {
            length := node.Len()
            if length == 0 {
                result.Issues = append(result.Issues, "数组为空")
            }
        }
    }
    
    return result
}
```

## 性能优化工具

### 性能测试工具

```go
func performanceTestUtils() {
    // 生成测试数据
    testData := generateTestData(1000)
    
    // 测试不同操作的性能
    operations := []struct {
        name string
        fn   func(fxjson.Node)
    }{
        {"基础访问", func(n fxjson.Node) {
            _ = n.Get("name").StringOr("")
        }},
        {"路径访问", func(n fxjson.Node) {
            _ = n.GetPath("users.0.profile.name").StringOr("")
        }},
        {"数组遍历", func(n fxjson.Node) {
            n.Get("users").ArrayForEach(func(i int, v fxjson.Node) bool {
                _ = v.Get("name").StringOr("")
                return true
            })
        }},
        {"类型转换", func(n fxjson.Node) {
            _ = n.Get("count").IntOr(0)
        }},
    }
    
    node := fxjson.FromBytes([]byte(testData))
    
    fmt.Println("性能测试结果:")
    for _, op := range operations {
        start := time.Now()
        
        // 执行操作多次以获得可靠的性能数据
        for i := 0; i < 1000; i++ {
            op.fn(node)
        }
        
        duration := time.Since(start)
        fmt.Printf("%s: 1000次操作耗时 %v (平均 %v/次)\n", 
            op.name, duration, duration/1000)
    }
}

func generateTestData(userCount int) string {
    var users []string
    
    for i := 0; i < userCount; i++ {
        user := fmt.Sprintf(`{
            "id": %d,
            "name": "用户%d",
            "profile": {
                "name": "用户%d",
                "age": %d
            }
        }`, i, i, i, 20+(i%50))
        
        users = append(users, user)
    }
    
    return fmt.Sprintf(`{
        "name": "测试数据",
        "count": %d,
        "users": [%s]
    }`, userCount, strings.Join(users, ","))
}
```

工具函数是 FxJSON 的强大补充，通过这些实用的工具，您可以更高效地处理各种 JSON 数据操作场景，从简单的类型转换到复杂的性能分析和错误诊断。