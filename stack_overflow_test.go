package fxjson

import (
	"strings"
	"testing"
)

// TestStackOverflowFix 测试栈溢出修复
func TestStackOverflowFix(t *testing.T) {
	// 测试解析深层嵌套的JSON（使用自定义选项允许更深的嵌套）
	t.Run("DeepNestedJSON", func(t *testing.T) {
		depth := 5000
		var builder strings.Builder
		
		// 构建深层嵌套的JSON字符串
		for i := 0; i < depth; i++ {
			builder.WriteString(`{"nested":`)
		}
		builder.WriteString(`"value"`)
		for i := 0; i < depth; i++ {
			builder.WriteString(`}`)
		}
		
		deepJSON := builder.String()
		
		// 使用自定义解析选项，允许更深的嵌套
		opts := ParseOptions{
			MaxDepth:      10000,       // 增加最大深度限制
			MaxStringLen:  1024 * 1024, // 1MB字符串限制
			MaxObjectKeys: 10000,       // 10000个键
			MaxArrayItems: 100000,      // 100000个数组项
			StrictMode:    false,
		}
		
		node := FromBytesWithOptions([]byte(deepJSON), opts)
		if !node.Exists() {
			t.Fatal("节点应该存在")
		}
		
		// 尝试访问深层数据
		current := node
		for i := 0; i < 10; i++ { // 只测试前10层，避免测试时间过长
			current = current.Get("nested")
			if !current.Exists() {
				t.Fatalf("第%d层nested节点应该存在", i+1)
			}
		}
	})
	
	// 测试深层转义JSON字符串的栈溢出修复
	t.Run("DeepEscapedJSON", func(t *testing.T) {
		var escaped strings.Builder
		escaped.WriteString(`{"data":"`)
		
		// 创建深层转义的JSON
		current := `{"level1":{"level2":{"level3":"value"}}}`
		for i := 0; i < 100; i++ { // 100层转义应该足够测试
			// 转义当前JSON字符串
			current = strings.ReplaceAll(current, `"`, `\"`)
			current = `{"nested":"` + current + `"}`
		}
		
		escaped.WriteString(current)
		escaped.WriteString(`"}`)
		
		node := FromBytes([]byte(escaped.String()))
		if !node.Exists() {
			t.Fatal("转义JSON节点应该存在")
		}
		
		// 测试数据访问
		dataNode := node.Get("data")
		if !dataNode.Exists() {
			t.Fatal("data节点应该存在")
		}
	})
}

// BenchmarkDeepJSON 基准测试深层JSON处理性能
func BenchmarkDeepJSON(b *testing.B) {
	// 创建中等深度的JSON进行性能测试
	depth := 1000
	var builder strings.Builder
	
	for i := 0; i < depth; i++ {
		builder.WriteString(`{"level":`)
	}
	builder.WriteString(`"value"`)
	for i := 0; i < depth; i++ {
		builder.WriteString(`}`)
	}
	
	deepJSON := []byte(builder.String())
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := FromBytes(deepJSON)
		_ = node.Get("level")
	}
}