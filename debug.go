package fxjson

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// DebugInfo 调试信息
type DebugInfo struct {
	ParseTime        time.Duration `json:"parse_time"`
	MemoryUsage      int64         `json:"memory_usage"`
	NodeCount        int           `json:"node_count"`
	MaxDepth         int           `json:"max_depth"`
	ErrorCount       int           `json:"error_count"`
	Warnings         []string      `json:"warnings"`
	Suggestions      []string      `json:"suggestions"`
	PerformanceHints []string      `json:"performance_hints"`
	StackTrace       []string      `json:"stack_trace,omitempty"`
}

// ParseError 增强的解析错误
type ParseError struct {
	Message    string    `json:"message"`
	Position   int       `json:"position"`
	Line       int       `json:"line"`
	Column     int       `json:"column"`
	Context    string    `json:"context"`
	Suggestion string    `json:"suggestion"`
	ErrorType  string    `json:"error_type"`
	Timestamp  time.Time `json:"timestamp"`
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("JSON parse error at line %d, column %d: %s\nContext: %s\nSuggestion: %s",
		pe.Line, pe.Column, pe.Message, pe.Context, pe.Suggestion)
}

// ValidationError 数据验证错误
type ValidationError struct {
	Field      string    `json:"field"`
	Value      string    `json:"value"`
	Rule       string    `json:"rule"`
	Message    string    `json:"message"`
	Suggestion string    `json:"suggestion"`
	Timestamp  time.Time `json:"timestamp"`
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("Validation error for field '%s': %s (value: %s, rule: %s)\nSuggestion: %s",
		ve.Field, ve.Message, ve.Value, ve.Rule, ve.Suggestion)
}

// Logger 日志接口
type Logger interface {
	Debug(message string, fields map[string]interface{})
	Info(message string, fields map[string]interface{})
	Warn(message string, fields map[string]interface{})
	Error(message string, fields map[string]interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct{}

func (dl *DefaultLogger) Debug(message string, fields map[string]interface{}) {
	fmt.Printf("[DEBUG] %s %+v\n", message, fields)
}

func (dl *DefaultLogger) Info(message string, fields map[string]interface{}) {
	fmt.Printf("[INFO] %s %+v\n", message, fields)
}

func (dl *DefaultLogger) Warn(message string, fields map[string]interface{}) {
	fmt.Printf("[WARN] %s %+v\n", message, fields)
}

func (dl *DefaultLogger) Error(message string, fields map[string]interface{}) {
	fmt.Printf("[ERROR] %s %+v\n", message, fields)
}

// 全局日志实例
var globalLogger Logger = &DefaultLogger{}

// SetLogger 设置全局日志器
func SetLogger(logger Logger) {
	globalLogger = logger
}

// DebugMode 调试模式
var DebugMode bool = false

// EnableDebugMode 启用调试模式
func EnableDebugMode() {
	DebugMode = true
}

// DisableDebugMode 禁用调试模式
func DisableDebugMode() {
	DebugMode = false
}

// FromBytesWithDebug 带调试信息的JSON解析
func FromBytesWithDebug(b []byte) (Node, *DebugInfo) {
	debugInfo := &DebugInfo{
		Warnings:         make([]string, 0),
		Suggestions:      make([]string, 0),
		PerformanceHints: make([]string, 0),
	}

	start := time.Now()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// 执行解析
	node := FromBytes(b)

	// 收集调试信息
	debugInfo.ParseTime = time.Since(start)
	runtime.ReadMemStats(&m2)
	debugInfo.MemoryUsage = int64(m2.Alloc - m1.Alloc)

	// 分析节点结构
	analyzeNode(node, debugInfo, 0)

	// 生成性能建议
	generatePerformanceHints(b, debugInfo)

	// 记录调试信息
	if DebugMode {
		globalLogger.Debug("JSON parsed with debug info", map[string]interface{}{
			"parse_time":   debugInfo.ParseTime,
			"memory_usage": debugInfo.MemoryUsage,
			"node_count":   debugInfo.NodeCount,
			"max_depth":    debugInfo.MaxDepth,
			"data_size":    len(b),
		})
	}

	return node, debugInfo
}

// analyzeNode 分析节点结构
func analyzeNode(node Node, debugInfo *DebugInfo, depth int) {
	debugInfo.NodeCount++

	if depth > debugInfo.MaxDepth {
		debugInfo.MaxDepth = depth
	}

	// 深度警告
	if depth > 50 {
		debugInfo.Warnings = append(debugInfo.Warnings,
			fmt.Sprintf("Deep nesting detected at depth %d, consider flattening the structure", depth))
	}

	switch node.Type() {
	case 'o':
		// 分析对象
		node.ForEach(func(key string, value Node) bool {
			analyzeNode(value, debugInfo, depth+1)

			// 检查空字符串键
			if key == "" {
				debugInfo.Warnings = append(debugInfo.Warnings, "Empty string key detected")
			}

			// 检查长键名
			if len(key) > 100 {
				debugInfo.Warnings = append(debugInfo.Warnings,
					fmt.Sprintf("Long key name detected: %s (length: %d)", key, len(key)))
			}

			return true
		})

	case 'a':
		// 分析数组
		arrayLen := node.Len()
		if arrayLen > 10000 {
			debugInfo.PerformanceHints = append(debugInfo.PerformanceHints,
				fmt.Sprintf("Large array detected (%d items), consider pagination or streaming", arrayLen))
		}

		for i := 0; i < arrayLen; i++ {
			analyzeNode(node.Index(i), debugInfo, depth+1)
		}

	case 's':
		// 分析字符串
		if str, err := node.String(); err == nil {
			if len(str) > 10000 {
				debugInfo.PerformanceHints = append(debugInfo.PerformanceHints,
					fmt.Sprintf("Large string detected (length: %d), consider external storage", len(str)))
			}
		}
	}
}

// generatePerformanceHints 生成性能建议
func generatePerformanceHints(data []byte, debugInfo *DebugInfo) {
	dataSize := len(data)

	// 数据大小建议
	if dataSize > 1024*1024 { // 1MB
		debugInfo.PerformanceHints = append(debugInfo.PerformanceHints,
			"Large JSON detected (>1MB), consider using streaming parser or compression")
	}

	// 解析时间建议
	if debugInfo.ParseTime > 100*time.Millisecond {
		debugInfo.PerformanceHints = append(debugInfo.PerformanceHints,
			"Slow parsing detected, consider caching parsed results")
	}

	// 内存使用建议
	if debugInfo.MemoryUsage > 10*1024*1024 { // 10MB
		debugInfo.PerformanceHints = append(debugInfo.PerformanceHints,
			"High memory usage detected, consider streaming processing")
	}

	// 结构复杂度建议
	if debugInfo.MaxDepth > 20 {
		debugInfo.Suggestions = append(debugInfo.Suggestions,
			"Consider flattening deeply nested structures for better performance")
	}

	if debugInfo.NodeCount > 50000 {
		debugInfo.PerformanceHints = append(debugInfo.PerformanceHints,
			"High node count detected, consider batch processing")
	}
}

// PrettyPrint 美化打印JSON结构
func (n Node) PrettyPrint() string {
	return n.PrettyPrintWithIndent("  ")
}

// PrettyPrintWithIndent 带自定义缩进的美化打印
func (n Node) PrettyPrintWithIndent(indent string) string {
	return prettyPrintNode(n, 0, indent)
}

// prettyPrintNode 递归打印节点
func prettyPrintNode(node Node, depth int, indent string) string {
	currentIndent := strings.Repeat(indent, depth)
	nextIndent := strings.Repeat(indent, depth+1)

	switch node.Type() {
	case 'o':
		if !node.Exists() {
			return "null"
		}

		var parts []string
		node.ForEach(func(key string, value Node) bool {
			valuePrint := prettyPrintNode(value, depth+1, indent)
			parts = append(parts, fmt.Sprintf("%s\"%s\": %s", nextIndent, key, valuePrint))
			return true
		})

		if len(parts) == 0 {
			return "{}"
		}

		return fmt.Sprintf("{\n%s\n%s}", strings.Join(parts, ",\n"), currentIndent)

	case 'a':
		if node.Len() == 0 {
			return "[]"
		}

		var parts []string
		for i := 0; i < node.Len(); i++ {
			itemPrint := prettyPrintNode(node.Index(i), depth+1, indent)
			parts = append(parts, fmt.Sprintf("%s%s", nextIndent, itemPrint))
		}

		return fmt.Sprintf("[\n%s\n%s]", strings.Join(parts, ",\n"), currentIndent)

	case 's':
		if str, err := node.String(); err == nil {
			return fmt.Sprintf("\"%s\"", escapeString(str))
		}
		return "\"<invalid string>\""

	case 'n':
		if num, err := node.Float(); err == nil {
			// 检查是否为整数
			if float64(int64(num)) == num {
				return fmt.Sprintf("%d", int64(num))
			}
			return fmt.Sprintf("%g", num)
		}
		return "<invalid number>"

	case 'b':
		if b, err := node.Bool(); err == nil {
			return fmt.Sprintf("%t", b)
		}
		return "<invalid boolean>"

	case 'l':
		return "null"

	default:
		return "<unknown type>"
	}
}

// escapeString 转义字符串
func escapeString(s string) string {
	// 简化的字符串转义
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// Inspect 详细检查节点
func (n Node) Inspect() map[string]interface{} {
	info := map[string]interface{}{
		"type":   n.Type(),
		"exists": n.Exists(),
		"raw":    string(n.Raw()),
	}

	switch n.Type() {
	case 'o':
		info["key_count"] = 0
		keys := make([]string, 0)
		n.ForEach(func(key string, value Node) bool {
			keys = append(keys, key)
			return true
		})
		info["key_count"] = len(keys)
		info["keys"] = keys

	case 'a':
		info["length"] = n.Len()
		if n.Len() > 0 {
			info["first_item_type"] = n.Index(0).Type()
		}

	case 's':
		if str, err := n.String(); err == nil {
			info["length"] = len(str)
			info["value"] = str
			if len(str) > 100 {
				info["preview"] = str[:100] + "..."
			}
		}

	case 'n':
		if num, err := n.Float(); err == nil {
			info["value"] = num
			info["is_integer"] = float64(int64(num)) == num
		}

	case 'b':
		if b, err := n.Bool(); err == nil {
			info["value"] = b
		}
	}

	return info
}

// Diff 比较两个JSON节点的差异
func (n Node) Diff(other Node) []DiffResult {
	var results []DiffResult
	diffNodes(n, other, "", &results)
	return results
}

// DiffResult 差异结果
type DiffResult struct {
	Path     string      `json:"path"`
	Type     string      `json:"type"` // added, removed, changed, type_changed
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
	OldType  string      `json:"old_type,omitempty"`
	NewType  string      `json:"new_type,omitempty"`
}

// diffNodes 递归比较节点
func diffNodes(node1, node2 Node, path string, results *[]DiffResult) {
	if !node1.Exists() && !node2.Exists() {
		return
	}

	if !node1.Exists() {
		*results = append(*results, DiffResult{
			Path:     path,
			Type:     "added",
			NewValue: getNodeValue(node2),
		})
		return
	}

	if !node2.Exists() {
		*results = append(*results, DiffResult{
			Path:     path,
			Type:     "removed",
			OldValue: getNodeValue(node1),
		})
		return
	}

	if node1.Type() != node2.Type() {
		*results = append(*results, DiffResult{
			Path:     path,
			Type:     "type_changed",
			OldValue: getNodeValue(node1),
			NewValue: getNodeValue(node2),
			OldType:  string(node1.Type()),
			NewType:  string(node2.Type()),
		})
		return
	}

	switch node1.Type() {
	case 'o':
		// 收集所有键
		keys := make(map[string]bool)
		node1.ForEach(func(key string, value Node) bool {
			keys[key] = true
			return true
		})
		node2.ForEach(func(key string, value Node) bool {
			keys[key] = true
			return true
		})

		// 比较每个键
		for key := range keys {
			keyPath := path
			if keyPath != "" {
				keyPath += "."
			}
			keyPath += key

			diffNodes(node1.Get(key), node2.Get(key), keyPath, results)
		}

	case 'a':
		len1, len2 := node1.Len(), node2.Len()
		maxLen := len1
		if len2 > maxLen {
			maxLen = len2
		}

		for i := 0; i < maxLen; i++ {
			indexPath := fmt.Sprintf("%s[%d]", path, i)

			var item1, item2 Node
			if i < len1 {
				item1 = node1.Index(i)
			}
			if i < len2 {
				item2 = node2.Index(i)
			}

			diffNodes(item1, item2, indexPath, results)
		}

	default:
		// 比较值
		val1 := getNodeValue(node1)
		val2 := getNodeValue(node2)

		if !equalValues(val1, val2) {
			*results = append(*results, DiffResult{
				Path:     path,
				Type:     "changed",
				OldValue: val1,
				NewValue: val2,
			})
		}
	}
}

// getNodeValue 获取节点值
func getNodeValue(node Node) interface{} {
	if !node.Exists() {
		return nil
	}

	switch node.Type() {
	case 's':
		if val, err := node.String(); err == nil {
			return val
		}
	case 'n':
		if val, err := node.Float(); err == nil {
			return val
		}
	case 'b':
		if val, err := node.Bool(); err == nil {
			return val
		}
	case 'l':
		return nil
	default:
		return string(node.Raw())
	}

	return nil
}

// equalValues 比较两个值是否相等
func equalValues(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// 使用JSON序列化进行深度比较
	aBytes, aErr := json.Marshal(a)
	bBytes, bErr := json.Marshal(b)

	if aErr != nil || bErr != nil {
		return false
	}

	return string(aBytes) == string(bBytes)
}

// GetStackTrace 获取调用栈
func GetStackTrace() []string {
	var traces []string
	for i := 1; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		traces = append(traces, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	return traces
}
