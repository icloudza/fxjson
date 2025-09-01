package fxjson

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestPerformanceOptimizations 测试性能优化
func TestPerformanceOptimizations(t *testing.T) {
	// 生成大型对象用于测试
	const objSize = 10000
	var builder strings.Builder
	builder.WriteByte('{')
	
	for i := 0; i < objSize; i++ {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(fmt.Sprintf(`"key_%d":{"type":"data","value":%f,"index":%d}`, i, float64(i)*0.5, i))
	}
	builder.WriteByte('}')
	
	largeJSON := []byte(builder.String())
	opts := ParseOptions{
		MaxDepth:      1000,
		MaxStringLen:  1024 * 1024 * 10,
		MaxObjectKeys: 50000,
		MaxArrayItems: 200000,
		StrictMode:    false,
	}
	
	node := FromBytesWithOptions(largeJSON, opts)
	if !node.Exists() {
		t.Fatal("大对象解析失败")
	}
	
	t.Run("OriginalGetPerformance", func(t *testing.T) {
		keys := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			keys[i] = fmt.Sprintf("key_%d", i*10) // 每隔10个取一个键
		}
		
		start := time.Now()
		for _, key := range keys {
			item := node.Get(key)
			if !item.Exists() {
				t.Fatalf("键 %s 不存在", key)
			}
		}
		originalTime := time.Since(start)
		t.Logf("原始Get方法1000次查找时间: %v", originalTime)
	})
	
	t.Run("OptimizedGetPerformance", func(t *testing.T) {
		
		keys := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			keys[i] = fmt.Sprintf("key_%d", i*10)
		}
		
		start := time.Now()
		for _, key := range keys {
			item := node.Get(key)
			if !item.Exists() {
				t.Fatalf("键 %s 不存在", key)
			}
		}
		optimizedTime := time.Since(start)
		t.Logf("优化后的Get方法1000次查找时间: %v", optimizedTime)
	})
	
}

// BenchmarkObjectAccess 对象访问性能基准测试
func BenchmarkObjectAccess(b *testing.B) {
	// 生成测试数据
	const objSize = 5000
	var builder strings.Builder
	builder.WriteByte('{')
	
	for i := 0; i < objSize; i++ {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(fmt.Sprintf(`"key_%d":{"value":%d}`, i, i))
	}
	builder.WriteByte('}')
	
	largeJSON := []byte(builder.String())
	opts := ParseOptions{
		MaxDepth:      1000,
		MaxStringLen:  1024 * 1024 * 10,
		MaxObjectKeys: 50000,
		MaxArrayItems: 200000,
		StrictMode:    false,
	}
	
	node := FromBytesWithOptions(largeJSON, opts)
	if !node.Exists() {
		b.Fatal("解析失败")
	}
	
	b.Run("OriginalGet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%objSize)
			node.Get(key)
		}
	})
	
	b.Run("OptimizedGet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%objSize)
			node.Get(key)
		}
	})
}

// TestGetPerformance 测试Get方法性能
func TestGetPerformance(t *testing.T) {
	const objSize = 1000
	var builder strings.Builder
	builder.WriteByte('{')
	
	for i := 0; i < objSize; i++ {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(fmt.Sprintf(`"key_%d":%d`, i, i))
	}
	builder.WriteByte('}')
	
	largeJSON := []byte(builder.String())
	opts := ParseOptions{
		MaxDepth:      1000,
		MaxStringLen:  1024 * 1024 * 10,
		MaxObjectKeys: 50000,
		MaxArrayItems: 200000,
		StrictMode:    false,
	}
	
	node := FromBytesWithOptions(largeJSON, opts)
	
	// 测试Get方法性能
	start := time.Now()
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := node.Get(key)
		if !value.Exists() {
			t.Fatalf("键 %s 不存在", key)
		}
		intVal, err := value.Int()
		if err != nil || intVal != int64(i) {
			t.Fatalf("值不匹配，期望: %d，实际: %d", i, intVal)
		}
	}
	elapsed := time.Since(start)
	
	t.Logf("100次Get操作时间: %v", elapsed)
	t.Logf("平均每次Get操作: %v", elapsed/100)
}

// TestMemoryUsageOptimization 测试内存使用优化
func TestMemoryUsageOptimization(t *testing.T) {
	const arraySize = 10000
	var builder strings.Builder
	builder.WriteByte('[')
	
	for i := 0; i < arraySize; i++ {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(fmt.Sprintf(`{"id":%d,"data":"value_%d"}`, i, i))
	}
	builder.WriteByte(']')
	
	testJSON := []byte(builder.String())
	
	// 测试多次解析，检查内存效率
	start := time.Now()
	var nodes []Node
	for i := 0; i < 10; i++ {
		node := FromBytes(testJSON)
		nodes = append(nodes, node)
		
		// 使用部分数据
		for j := 0; j < 100; j++ {
			item := node.Index(j)
			item.Get("id").Int()
		}
	}
	elapsed := time.Since(start)
	
	t.Logf("10次大数组解析和使用时间: %v", elapsed)
	t.Logf("平均每次: %v", elapsed/10)
	
	// 验证结果
	if len(nodes) != 10 {
		t.Fatalf("节点数量不匹配: %d != 10", len(nodes))
	}
	
	for _, node := range nodes {
		if node.Len() != arraySize {
			t.Fatalf("数组长度不匹配: %d != %d", node.Len(), arraySize)
		}
	}
}