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
	
	t.Run("OptimizedGetFastPerformance", func(t *testing.T) {
		ClearKeyCache() // 清除缓存确保公平比较
		
		keys := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			keys[i] = fmt.Sprintf("key_%d", i*10)
		}
		
		start := time.Now()
		for _, key := range keys {
			item := node.GetFast(key)
			if !item.Exists() {
				t.Fatalf("键 %s 不存在", key)
			}
		}
		optimizedTime := time.Since(start)
		t.Logf("优化GetFast方法1000次查找时间: %v", optimizedTime)
	})
	
	t.Run("BatchAccessPerformance", func(t *testing.T) {
		keys := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			keys[i] = fmt.Sprintf("key_%d", i*10)
		}
		
		start := time.Now()
		batchAccess := node.NewBatchAccess(keys)
		results := batchAccess.GetAll()
		batchTime := time.Since(start)
		
		// 验证结果
		if len(results) != 1000 {
			t.Fatalf("批量访问结果数量不匹配: %d != 1000", len(results))
		}
		
		for _, key := range keys {
			if !results[key].Exists() {
				t.Fatalf("批量访问中键 %s 不存在", key)
			}
		}
		
		t.Logf("批量访问方法1000个键时间: %v", batchTime)
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
	
	b.Run("OptimizedGetFast", func(b *testing.B) {
		ClearKeyCache()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%objSize)
			node.GetFast(key)
		}
	})
	
	b.Run("BatchAccess", func(b *testing.B) {
		keys := make([]string, 100)
		for i := 0; i < 100; i++ {
			keys[i] = fmt.Sprintf("key_%d", i)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			batchAccess := node.NewBatchAccess(keys)
			batchAccess.GetAll()
		}
	})
}

// TestCacheEffectiveness 测试缓存效果
func TestCacheEffectiveness(t *testing.T) {
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
	ClearKeyCache()
	
	// 第一次访问（冷缓存）
	start := time.Now()
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		node.GetFast(key)
	}
	coldTime := time.Since(start)
	
	// 第二次访问（热缓存）
	start = time.Now()
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		node.GetFast(key)
	}
	hotTime := time.Since(start)
	
	t.Logf("冷缓存访问时间: %v", coldTime)
	t.Logf("热缓存访问时间: %v", hotTime)
	
	if hotTime > coldTime {
		t.Log("缓存可能没有生效，或者对象太小缓存未启用")
	} else {
		speedup := float64(coldTime) / float64(hotTime)
		t.Logf("缓存加速比: %.2fx", speedup)
	}
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