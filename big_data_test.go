package fxjson

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// TestBigDataHandling 测试大数据量处理
func TestBigDataHandling(t *testing.T) {
	// 测试大型JSON数组解析
	t.Run("LargeArrayParsing", func(t *testing.T) {
		const arraySize = 50000
		var builder strings.Builder
		builder.WriteByte('[')
		
		for i := 0; i < arraySize; i++ {
			if i > 0 {
				builder.WriteByte(',')
			}
			builder.WriteString(fmt.Sprintf(`{"id":%d,"name":"user_%d","value":%f}`, i, i, rand.Float64()*1000))
		}
		builder.WriteByte(']')
		
		largeJSON := builder.String()
		t.Logf("生成大型JSON，大小: %d bytes", len(largeJSON))
		
		// 使用自定义选项增加限制
		opts := ParseOptions{
			MaxDepth:      1000,
			MaxStringLen:  1024 * 1024 * 10, // 10MB字符串限制
			MaxObjectKeys: 100000,            // 增加对象键限制
			MaxArrayItems: 200000,            // 增加数组项限制
			StrictMode:    false,
		}
		
		start := time.Now()
		node := FromBytesWithOptions([]byte(largeJSON), opts)
		parseTime := time.Since(start)
		
		if !node.Exists() {
			t.Fatal("大型JSON解析失败")
		}
		
		if node.Len() != arraySize {
			t.Fatalf("数组长度不匹配，期望: %d，实际: %d", arraySize, node.Len())
		}
		
		// 测试随机访问性能
		start = time.Now()
		for i := 0; i < 1000; i++ {
			idx := rand.Intn(arraySize)
			item := node.Index(idx)
			if !item.Exists() {
				t.Fatalf("索引 %d 的元素不存在", idx)
			}
			id, err := item.Get("id").Int()
			if err != nil || id != int64(idx) {
				t.Fatalf("索引 %d 的ID不匹配，期望: %d，实际: %d", idx, idx, id)
			}
		}
		accessTime := time.Since(start)
		
		t.Logf("解析时间: %v, 1000次随机访问时间: %v", parseTime, accessTime)
		
		// 性能要求：解析时间应该在合理范围内
		if parseTime > 5*time.Second {
			t.Errorf("解析时间过长: %v", parseTime)
		}
		if accessTime > 1*time.Second {
			t.Errorf("访问时间过长: %v", accessTime)
		}
	})
	
	// 测试大型JSON对象解析
	t.Run("LargeObjectParsing", func(t *testing.T) {
		const objSize = 50000
		var builder strings.Builder
		builder.WriteByte('{')
		
		for i := 0; i < objSize; i++ {
			if i > 0 {
				builder.WriteByte(',')
			}
			builder.WriteString(fmt.Sprintf(`"key_%d":{"type":"data","value":%f,"index":%d}`, i, rand.Float64()*1000, i))
		}
		builder.WriteByte('}')
		
		largeJSON := builder.String()
		t.Logf("生成大型对象JSON，大小: %d bytes", len(largeJSON))
		
		// 使用自定义选项增加限制
		opts := ParseOptions{
			MaxDepth:      1000,
			MaxStringLen:  1024 * 1024 * 10, // 10MB字符串限制
			MaxObjectKeys: 100000,            // 增加对象键限制
			MaxArrayItems: 200000,            // 增加数组项限制
			StrictMode:    false,
		}
		
		start := time.Now()
		node := FromBytesWithOptions([]byte(largeJSON), opts)
		parseTime := time.Since(start)
		
		if !node.Exists() {
			t.Fatal("大型对象JSON解析失败")
		}
		
		if node.Len() != objSize {
			t.Fatalf("对象键数量不匹配，期望: %d，实际: %d", objSize, node.Len())
		}
		
		// 测试键查找性能
		start = time.Now()
		for i := 0; i < 1000; i++ {
			keyIdx := rand.Intn(objSize)
			key := fmt.Sprintf("key_%d", keyIdx)
			item := node.Get(key)
			if !item.Exists() {
				t.Fatalf("键 %s 的值不存在", key)
			}
			index, err := item.Get("index").Int()
			if err != nil || index != int64(keyIdx) {
				t.Fatalf("键 %s 的index不匹配，期望: %d，实际: %d", key, keyIdx, index)
			}
		}
		accessTime := time.Since(start)
		
		t.Logf("对象解析时间: %v, 1000次键查找时间: %v", parseTime, accessTime)
		
		// 性能要求
		if parseTime > 3*time.Second {
			t.Errorf("对象解析时间过长: %v", parseTime)
		}
		if accessTime > 1*time.Second {
			t.Errorf("键查找时间过长: %v", accessTime)
		}
	})
	
	// 测试深度嵌套的JSON
	t.Run("DeeplyNestedJSON", func(t *testing.T) {
		const depth = 1000
		var builder strings.Builder
		
		// 构建深层嵌套结构
		for i := 0; i < depth; i++ {
			builder.WriteString(`{"level":`)
		}
		builder.WriteString(`"bottom"`)
		for i := 0; i < depth; i++ {
			builder.WriteByte('}')
		}
		
		deepJSON := builder.String()
		
		// 使用更高的深度限制
		opts := ParseOptions{
			MaxDepth:      2000,
			MaxStringLen:  1024 * 1024,
			MaxObjectKeys: 10000,
			MaxArrayItems: 100000,
			StrictMode:    false,
		}
		
		start := time.Now()
		node := FromBytesWithOptions([]byte(deepJSON), opts)
		parseTime := time.Since(start)
		
		if !node.Exists() {
			t.Fatal("深层嵌套JSON解析失败")
		}
		
		// 测试深层访问
		start = time.Now()
		current := node
		for i := 0; i < 100; i++ { // 测试前100层
			current = current.Get("level")
			if !current.Exists() {
				t.Fatalf("第%d层level不存在", i+1)
			}
		}
		accessTime := time.Since(start)
		
		t.Logf("深层嵌套解析时间: %v, 100层访问时间: %v", parseTime, accessTime)
		
		// 性能要求
		if parseTime > 2*time.Second {
			t.Errorf("深层嵌套解析时间过长: %v", parseTime)
		}
		if accessTime > 500*time.Millisecond {
			t.Errorf("深层访问时间过长: %v", accessTime)
		}
	})
}

// TestMemoryUsage 测试内存使用情况
func TestMemoryUsage(t *testing.T) {
	t.Run("MemoryEfficiencyTest", func(t *testing.T) {
		// 生成适中大小的测试数据
		const arraySize = 10000
		var builder strings.Builder
		builder.WriteByte('[')
		
		for i := 0; i < arraySize; i++ {
			if i > 0 {
				builder.WriteByte(',')
			}
			builder.WriteString(fmt.Sprintf(`{"id":%d,"name":"test_user_%d","data":[1,2,3,4,5],"meta":{"created":"2023-01-01","active":true}}`, i, i))
		}
		builder.WriteByte(']')
		
		testJSON := []byte(builder.String())
		
		// 解析多次，检查是否有内存泄漏
		for iteration := 0; iteration < 100; iteration++ {
			node := FromBytes(testJSON)
			if !node.Exists() {
				t.Fatal("JSON解析失败")
			}
			
			// 遍历一些数据
			for i := 0; i < 100; i++ {
				item := node.Index(i)
				if item.Exists() {
					item.Get("name").String()
					item.Get("id").Int()
				}
			}
		}
		
		t.Log("内存效率测试完成")
	})
}

// TestConcurrencyHandling 测试并发处理
func TestConcurrencyHandling(t *testing.T) {
	const goroutineCount = 100
	const jsonSize = 1000
	
	// 生成测试JSON
	var builder strings.Builder
	builder.WriteByte('[')
	for i := 0; i < jsonSize; i++ {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(fmt.Sprintf(`{"id":%d,"value":%f}`, i, rand.Float64()))
	}
	builder.WriteByte(']')
	testJSON := []byte(builder.String())
	
	// 并发测试
	results := make(chan error, goroutineCount)
	
	for i := 0; i < goroutineCount; i++ {
		go func(id int) {
			node := FromBytes(testJSON)
			if !node.Exists() {
				results <- fmt.Errorf("goroutine %d: 解析失败", id)
				return
			}
			
			// 随机访问测试
			for j := 0; j < 100; j++ {
				idx := rand.Intn(jsonSize)
				item := node.Index(idx)
				if !item.Exists() {
					results <- fmt.Errorf("goroutine %d: 索引 %d 不存在", id, idx)
					return
				}
				
				idVal, err := item.Get("id").Int()
				if err != nil || idVal != int64(idx) {
					results <- fmt.Errorf("goroutine %d: ID不匹配 %d != %d", id, idVal, idx)
					return
				}
			}
			
			results <- nil
		}(i)
	}
	
	// 检查结果
	for i := 0; i < goroutineCount; i++ {
		if err := <-results; err != nil {
			t.Error(err)
		}
	}
	
	t.Log("并发测试完成")
}

// BenchmarkLargeDataProcessing 大数据处理性能基准测试
func BenchmarkLargeDataProcessing(b *testing.B) {
	// 生成测试数据
	const arraySize = 10000
	var builder strings.Builder
	builder.WriteByte('[')
	
	for i := 0; i < arraySize; i++ {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(fmt.Sprintf(`{"id":%d,"name":"user_%d","score":%f,"active":true}`, i, i, rand.Float64()*100))
	}
	builder.WriteByte(']')
	
	testData := []byte(builder.String())
	b.Logf("测试数据大小: %d bytes", len(testData))
	
	b.ResetTimer()
	
	b.Run("ParseOnly", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			node := FromBytes(testData)
			if !node.Exists() {
				b.Fatal("解析失败")
			}
		}
	})
	
	b.Run("ParseAndIterate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			node := FromBytes(testData)
			if !node.Exists() {
				b.Fatal("解析失败")
			}
			
			// 遍历所有元素
			count := 0
			node.ArrayForEach(func(index int, value Node) bool {
				count++
				return true
			})
			
			if count != arraySize {
				b.Fatalf("遍历数量不匹配: %d != %d", count, arraySize)
			}
		}
	})
	
	b.Run("RandomAccess", func(b *testing.B) {
		node := FromBytes(testData)
		if !node.Exists() {
			b.Fatal("解析失败")
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := rand.Intn(arraySize)
			item := node.Index(idx)
			if !item.Exists() {
				b.Fatal("随机访问失败")
			}
		}
	})
}