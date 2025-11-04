# 调试功能指南

FxJSON 提供了完善的调试工具，帮助您快速定位问题、分析 JSON 结构、对比数据差异，并优化性能。

## 目录

- [调试模式](#调试模式)
- [美化打印](#美化打印)
- [节点检查](#节点检查)
- [数据对比](#数据对比)
- [性能分析](#性能分析)
- [日志系统](#日志系统)
- [实用调试技巧](#实用调试技巧)

---

## 调试模式

### EnableDebugMode() / DisableDebugMode()

全局启用或禁用调试模式。

```go
func EnableDebugMode()
func DisableDebugMode()
```

### FromBytesWithDebug()

使用调试模式解析 JSON,返回详细的调试信息。

```go
func FromBytesWithDebug(b []byte) (Node, *DebugInfo)
```

### DebugInfo 结构

```go
type DebugInfo struct {
    ParseTime       time.Duration          // 解析耗时
    NodeCount       int                    // 节点总数
    MaxDepth        int                    // 最大嵌套深度
    DataSize        int                    // 数据大小(字节)
    ObjectCount     int                    // 对象数量
    ArrayCount      int                    // 数组数量
    StringCount     int                    // 字符串数量
    NumberCount     int                    // 数字数量
    PerformanceHints []string              // 性能优化建议
}
```

### 基础调试示例

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := `{
        "users": [
            {
                "id": 1,
                "name": "张三",
                "profile": {
                    "age": 28,
                    "city": "北京",
                    "hobbies": ["阅读", "编程", "旅行"]
                }
            },
            {
                "id": 2,
                "name": "李四",
                "profile": {
                    "age": 32,
                    "city": "上海",
                    "hobbies": ["运动", "音乐"]
                }
            }
        ],
        "metadata": {
            "total": 2,
            "version": "1.0"
        }
    }`

    // 启用调试模式
    fxjson.EnableDebugMode()
    defer fxjson.DisableDebugMode()

    // 使用调试模式解析
    node, debugInfo := fxjson.FromBytesWithDebug([]byte(jsonData))

    fmt.Println("=== 调试信息 ===")
    fmt.Printf("解析耗时: %v\n", debugInfo.ParseTime)
    fmt.Printf("数据大小: %d 字节\n", debugInfo.DataSize)
    fmt.Printf("节点总数: %d\n", debugInfo.NodeCount)
    fmt.Printf("最大深度: %d\n", debugInfo.MaxDepth)
    fmt.Printf("对象数量: %d\n", debugInfo.ObjectCount)
    fmt.Printf("数组数量: %d\n", debugInfo.ArrayCount)
    fmt.Printf("字符串数量: %d\n", debugInfo.StringCount)
    fmt.Printf("数字数量: %d\n", debugInfo.NumberCount)

    if len(debugInfo.PerformanceHints) > 0 {
        fmt.Println("\n性能优化建议:")
        for i, hint := range debugInfo.PerformanceHints {
            fmt.Printf("%d. %s\n", i+1, hint)
        }
    }

    // 继续使用解析的节点
    userName := node.GetPath("users.0.name").StringOr("")
    fmt.Printf("\n第一个用户: %s\n", userName)
}
```

---

## 美化打印

### PrettyPrint() / PrettyPrintWithIndent()

美化输出 JSON 结构,便于阅读和调试。

```go
func (n Node) PrettyPrint() string
func (n Node) PrettyPrintWithIndent(indent string) string
```

### 美化打印示例

```go
func prettyPrintExample() {
    compactJSON := `{"name":"张三","age":30,"hobbies":["阅读","编程"],"address":{"city":"北京","district":"朝阳区"}}`

    node := fxjson.FromString(compactJSON)

    // 默认美化(2空格缩进)
    fmt.Println("默认美化:")
    fmt.Println(node.PrettyPrint())

    // 自定义缩进(4空格)
    fmt.Println("\n自定义缩进(4空格):")
    fmt.Println(node.PrettyPrintWithIndent("    "))

    // 自定义缩进(制表符)
    fmt.Println("\n制表符缩进:")
    fmt.Println(node.PrettyPrintWithIndent("\t"))
}
```

**输出示例**:
```json
{
  "name": "张三",
  "age": 30,
  "hobbies": [
    "阅读",
    "编程"
  ],
  "address": {
    "city": "北京",
    "district": "朝阳区"
  }
}
```

### 条件美化

```go
func conditionalPrettyPrint() {
    node := fxjson.FromString(`{"data": {"users": [{"id": 1}, {"id": 2}]}}`)

    // 只美化特定部分
    usersNode := node.GetPath("data.users")
    fmt.Println("仅美化用户数组:")
    fmt.Println(usersNode.PrettyPrint())

    // 美化整个文档
    fmt.Println("\n美化整个文档:")
    fmt.Println(node.PrettyPrint())
}
```

---

## 节点检查

### Inspect()

深入检查节点的详细信息。

```go
func (n Node) Inspect() map[string]interface{}
```

### 检查示例

```go
func inspectExample() {
    jsonData := `{
        "user": {
            "id": 1001,
            "name": "张三",
            "tags": ["golang", "developer"],
            "active": true,
            "score": 95.5,
            "metadata": null
        }
    }`

    node := fxjson.FromString(jsonData)
    userNode := node.Get("user")

    // 检查节点
    info := userNode.Inspect()

    fmt.Println("=== 节点检查信息 ===")
    for key, value := range info {
        fmt.Printf("%s: %v\n", key, value)
    }

    // 检查特定字段
    fmt.Println("\n=== 字段详细检查 ===")
    fields := []string{"id", "name", "tags", "active", "score", "metadata"}

    for _, field := range fields {
        fieldNode := userNode.Get(field)
        fieldInfo := fieldNode.Inspect()

        fmt.Printf("\n字段: %s\n", field)
        fmt.Printf("  类型: %v\n", fieldInfo["type"])
        fmt.Printf("  存在: %v\n", fieldInfo["exists"])
        fmt.Printf("  值: %v\n", fieldInfo["value"])

        if size, ok := fieldInfo["size"]; ok {
            fmt.Printf("  大小: %v\n", size)
        }
    }
}
```

### 递归检查

```go
func recursiveInspect(node fxjson.Node, path string, depth int) {
    indent := strings.Repeat("  ", depth)
    info := node.Inspect()

    nodeType := info["type"]
    fmt.Printf("%s[%s] 类型: %v", indent, path, nodeType)

    if exists, ok := info["exists"].(bool); ok && !exists {
        fmt.Printf(" (不存在)\n")
        return
    }

    fmt.Println()

    // 递归检查子节点
    if node.IsObject() {
        node.ForEach(func(key string, child fxjson.Node) bool {
            childPath := path + "." + key
            if path == "" {
                childPath = key
            }
            recursiveInspect(child, childPath, depth+1)
            return true
        })
    } else if node.IsArray() {
        node.ArrayForEach(func(index int, child fxjson.Node) bool {
            childPath := fmt.Sprintf("%s[%d]", path, index)
            recursiveInspect(child, childPath, depth+1)
            return true
        })
    } else {
        if value, ok := info["value"]; ok {
            fmt.Printf("%s  值: %v\n", indent, value)
        }
    }
}

func inspectTree() {
    jsonData := `{
        "company": {
            "name": "科技公司",
            "employees": [
                {"name": "张三", "role": "工程师"},
                {"name": "李四", "role": "经理"}
            ]
        }
    }`

    node := fxjson.FromString(jsonData)

    fmt.Println("=== JSON 树结构 ===")
    recursiveInspect(node, "", 0)
}
```

---

## 数据对比

### Diff()

对比两个 JSON 节点的差异。

```go
func (n Node) Diff(other Node) []DiffResult
```

### DiffResult 结构

```go
type DiffResult struct {
    Path     string      // 差异路径
    Type     string      // 差异类型: "added", "removed", "modified", "type_changed"
    OldValue interface{} // 旧值
    NewValue interface{} // 新值
}
```

### 差异对比示例

```go
func diffExample() {
    json1 := `{
        "user": {
            "id": 1,
            "name": "张三",
            "age": 28,
            "city": "北京",
            "tags": ["golang", "developer"]
        }
    }`

    json2 := `{
        "user": {
            "id": 1,
            "name": "张三",
            "age": 29,
            "city": "上海",
            "email": "zhang@example.com",
            "tags": ["golang", "developer", "blogger"]
        }
    }`

    node1 := fxjson.FromString(json1)
    node2 := fxjson.FromString(json2)

    // 对比差异
    diffs := node1.Diff(node2)

    fmt.Printf("=== 发现 %d 处差异 ===\n\n", len(diffs))

    for i, diff := range diffs {
        fmt.Printf("差异 %d:\n", i+1)
        fmt.Printf("  路径: %s\n", diff.Path)
        fmt.Printf("  类型: %s\n", diff.Type)

        switch diff.Type {
        case "modified":
            fmt.Printf("  旧值: %v\n", diff.OldValue)
            fmt.Printf("  新值: %v\n", diff.NewValue)
        case "added":
            fmt.Printf("  新增值: %v\n", diff.NewValue)
        case "removed":
            fmt.Printf("  移除值: %v\n", diff.OldValue)
        case "type_changed":
            fmt.Printf("  类型变化: %T -> %T\n", diff.OldValue, diff.NewValue)
        }
        fmt.Println()
    }
}
```

### 版本对比

```go
func versionComparison() {
    v1 := `{
        "version": "1.0.0",
        "features": ["feature1", "feature2"],
        "config": {
            "debug": false,
            "timeout": 30
        }
    }`

    v2 := `{
        "version": "1.1.0",
        "features": ["feature1", "feature2", "feature3"],
        "config": {
            "debug": true,
            "timeout": 60,
            "retry": 3
        }
    }`

    node1 := fxjson.FromString(v1)
    node2 := fxjson.FromString(v2)

    diffs := node1.Diff(node2)

    fmt.Println("=== 版本变更 ===")
    fmt.Printf("从 %s 到 %s\n\n",
        node1.Get("version").StringOr(""),
        node2.Get("version").StringOr(""))

    // 分类显示差异
    added := []DiffResult{}
    modified := []DiffResult{}
    removed := []DiffResult{}

    for _, diff := range diffs {
        switch diff.Type {
        case "added":
            added = append(added, diff)
        case "modified":
            modified = append(modified, diff)
        case "removed":
            removed = append(removed, diff)
        }
    }

    if len(added) > 0 {
        fmt.Println("新增:")
        for _, diff := range added {
            fmt.Printf("  + %s: %v\n", diff.Path, diff.NewValue)
        }
        fmt.Println()
    }

    if len(modified) > 0 {
        fmt.Println("修改:")
        for _, diff := range modified {
            fmt.Printf("  ~ %s: %v -> %v\n", diff.Path, diff.OldValue, diff.NewValue)
        }
        fmt.Println()
    }

    if len(removed) > 0 {
        fmt.Println("移除:")
        for _, diff := range removed {
            fmt.Printf("  - %s: %v\n", diff.Path, diff.OldValue)
        }
    }
}
```

---

## 性能分析

### 获取堆栈跟踪

```go
func GetStackTrace() []string
```

获取当前的调用堆栈,用于定位问题。

```go
func stackTraceExample() {
    stack := fxjson.GetStackTrace()

    fmt.Println("=== 调用堆栈 ===")
    for i, frame := range stack {
        fmt.Printf("%d. %s\n", i+1, frame)
    }
}
```

### 性能基准测试

```go
func performanceBenchmark() {
    testData := generateLargeJSON(1000)

    operations := []struct {
        name string
        fn   func()
    }{
        {
            "解析",
            func() { fxjson.FromString(testData) },
        },
        {
            "路径访问",
            func() {
                node := fxjson.FromString(testData)
                node.GetPath("users.0.name").StringOr("")
            },
        },
        {
            "数组遍历",
            func() {
                node := fxjson.FromString(testData)
                users := node.Get("users")
                users.ArrayForEach(func(i int, user fxjson.Node) bool {
                    _ = user.Get("name").StringOr("")
                    return true
                })
            },
        },
    }

    fmt.Println("=== 性能基准测试 ===")
    iterations := 100

    for _, op := range operations {
        start := time.Now()

        for i := 0; i < iterations; i++ {
            op.fn()
        }

        duration := time.Since(start)
        avgTime := duration / time.Duration(iterations)

        fmt.Printf("%s:\n", op.name)
        fmt.Printf("  总耗时: %v\n", duration)
        fmt.Printf("  平均耗时: %v\n", avgTime)
        fmt.Printf("  吞吐量: %.0f ops/sec\n\n",
            float64(iterations)/duration.Seconds())
    }
}

func generateLargeJSON(userCount int) string {
    var users []string
    for i := 0; i < userCount; i++ {
        user := fmt.Sprintf(`{"id":%d,"name":"用户%d","age":%d}`,
            i, i, 20+i%50)
        users = append(users, user)
    }
    return fmt.Sprintf(`{"users":[%s]}`, strings.Join(users, ","))
}
```

---

## 日志系统

### 设置日志记录器

```go
func SetLogger(logger Logger)
```

### Logger 接口

```go
type Logger interface {
    Debug(message string, fields map[string]interface{})
    Info(message string, fields map[string]interface{})
    Warn(message string, fields map[string]interface{})
    Error(message string, fields map[string]interface{})
}
```

### 自定义日志记录器

```go
type CustomLogger struct {
    prefix string
}

func (cl *CustomLogger) Debug(message string, fields map[string]interface{}) {
    cl.log("DEBUG", message, fields)
}

func (cl *CustomLogger) Info(message string, fields map[string]interface{}) {
    cl.log("INFO", message, fields)
}

func (cl *CustomLogger) Warn(message string, fields map[string]interface{}) {
    cl.log("WARN", message, fields)
}

func (cl *CustomLogger) Error(message string, fields map[string]interface{}) {
    cl.log("ERROR", message, fields)
}

func (cl *CustomLogger) log(level, message string, fields map[string]interface{}) {
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    fmt.Printf("[%s] [%s] %s%s", timestamp, level, cl.prefix, message)

    if len(fields) > 0 {
        fmt.Print(" | ")
        for k, v := range fields {
            fmt.Printf("%s=%v ", k, v)
        }
    }
    fmt.Println()
}

func loggerExample() {
    // 设置自定义日志记录器
    logger := &CustomLogger{prefix: "FxJSON: "}
    fxjson.SetLogger(logger)

    // 启用调试模式
    fxjson.EnableDebugMode()
    defer fxjson.DisableDebugMode()

    // 进行 JSON 操作,将自动记录日志
    jsonData := `{"user": {"name": "张三", "age": 30}}`
    node := fxjson.FromString(jsonData)

    name := node.GetPath("user.name").StringOr("")
    fmt.Printf("\n用户名: %s\n", name)
}
```

---

## 实用调试技巧

### 1. 快速定位问题

```go
func quickDebug() {
    jsonData := `{"users": [{"id": 1, "name": "张三"}]}`
    node := fxjson.FromString(jsonData)

    // 检查节点是否存在
    userNode := node.GetPath("users.0")
    if !userNode.Exists() {
        fmt.Println("❌ 用户节点不存在")
        // 打印整个结构
        fmt.Println("实际结构:")
        fmt.Println(node.PrettyPrint())
        return
    }

    // 检查类型
    if !userNode.IsObject() {
        fmt.Printf("❌ 期望对象,实际是: %v\n", userNode.Kind())
        return
    }

    fmt.Println("✅ 节点验证通过")
}
```

### 2. 调试深层嵌套

```go
func debugNestedPath() {
    jsonData := `{
        "data": {
            "response": {
                "users": [
                    {"profile": {"name": "张三"}}
                ]
            }
        }
    }`

    node := fxjson.FromString(jsonData)
    targetPath := "data.response.users.0.profile.name"

    // 逐级检查路径
    parts := strings.Split(targetPath, ".")
    currentNode := node
    currentPath := ""

    fmt.Println("=== 路径跟踪 ===")
    for i, part := range parts {
        if currentPath == "" {
            currentPath = part
        } else {
            currentPath += "." + part
        }

        // 尝试访问
        if i == len(parts)-1 {
            currentNode = currentNode.Get(part)
        } else {
            // 数组索引
            if len(part) > 0 && part[0] >= '0' && part[0] <= '9' {
                idx := 0
                fmt.Sscanf(part, "%d", &idx)
                currentNode = currentNode.Index(idx)
            } else {
                currentNode = currentNode.Get(part)
            }
        }

        if !currentNode.Exists() {
            fmt.Printf("❌ 路径在 '%s' 处断开\n", currentPath)
            return
        }

        fmt.Printf("✅ %s: %v\n", currentPath, currentNode.Kind())
    }

    fmt.Printf("\n最终值: %s\n", currentNode.StringOr(""))
}
```

### 3. 性能瓶颈分析

```go
func analyzePerformanceBottleneck() {
    jsonData := generateLargeJSON(5000)

    operations := map[string]func(fxjson.Node){
        "GetPath访问": func(n fxjson.Node) {
            for i := 0; i < 100; i++ {
                n.GetPath(fmt.Sprintf("users.%d.name", i%1000)).StringOr("")
            }
        },
        "Index访问": func(n fxjson.Node) {
            users := n.Get("users")
            for i := 0; i < 100; i++ {
                users.Index(i % 1000).Get("name").StringOr("")
            }
        },
        "ArrayForEach遍历": func(n fxjson.Node) {
            count := 0
            n.Get("users").ArrayForEach(func(i int, user fxjson.Node) bool {
                _ = user.Get("name").StringOr("")
                count++
                return count < 100
            })
        },
    }

    node := fxjson.FromString(jsonData)

    fmt.Println("=== 性能瓶颈分析 ===")
    for name, op := range operations {
        start := time.Now()
        op(node)
        duration := time.Since(start)

        fmt.Printf("%s: %v\n", name, duration)
    }
}
```

### 4. 内存使用监控

```go
func monitorMemoryUsage() {
    var m1, m2 runtime.MemStats

    // 初始内存
    runtime.ReadMemStats(&m1)

    // 执行操作
    largeJSON := generateLargeJSON(10000)
    node := fxjson.FromString(largeJSON)

    // 遍历数据
    count := 0
    node.Get("users").ArrayForEach(func(i int, user fxjson.Node) bool {
        _ = user.Get("name").StringOr("")
        count++
        return true
    })

    // 操作后内存
    runtime.ReadMemStats(&m2)

    fmt.Println("=== 内存使用分析 ===")
    fmt.Printf("处理节点数: %d\n", count)
    fmt.Printf("内存分配: %d KB\n", (m2.Alloc-m1.Alloc)/1024)
    fmt.Printf("总分配: %d KB\n", (m2.TotalAlloc-m1.TotalAlloc)/1024)
    fmt.Printf("GC 次数: %d\n", m2.NumGC-m1.NumGC)
}
```

---

## 最佳实践

### 1. 开发时启用调试

```go
func init() {
    if os.Getenv("DEBUG") == "true" {
        fxjson.EnableDebugMode()
    }
}
```

### 2. 生产环境错误追踪

```go
func productionErrorTracking() {
    node := fxjson.FromString(jsonData)

    value := node.GetPath("some.deep.path")
    if !value.Exists() {
        // 记录详细信息用于调试
        log.Printf("路径不存在: some.deep.path")
        log.Printf("实际结构: %s", node.PrettyPrint())

        // 获取堆栈信息
        stack := fxjson.GetStackTrace()
        log.Printf("调用堆栈: %v", stack)
    }
}
```

### 3. 单元测试中的调试

```go
func TestJSONParsing(t *testing.T) {
    jsonData := `{"user": {"name": "张三"}}`
    node := fxjson.FromString(jsonData)

    name := node.GetPath("user.name").StringOr("")

    if name != "张三" {
        t.Errorf("期望 '张三', 实际 '%s'", name)

        // 输出调试信息
        t.Logf("完整结构:\n%s", node.PrettyPrint())

        // 检查节点
        info := node.GetPath("user.name").Inspect()
        t.Logf("节点信息: %+v", info)
    }
}
```

调试功能让 FxJSON 的开发体验更加友好,通过这些工具,您可以快速定位问题、优化性能、确保代码质量。
