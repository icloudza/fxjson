package fxjson

import (
	"strings"
	"testing"
	"time"
)

// 基准测试用的JSON数据
const benchmarkJSON = `{
  "code": 0,
  "success": true,
  "data": {
    "notes": [
      {"id": "1", "title": "测试笔记1", "view_count": 1000, "category": "tech", "revenue": 50.5, "engagement_rate": 7.2},
      {"id": "2", "title": "测试笔记2", "view_count": 2000, "category": "life", "revenue": 80.0, "engagement_rate": 5.8},
      {"id": "3", "title": "测试笔记3", "view_count": 500, "category": "tech", "revenue": 30.0, "engagement_rate": 8.9},
      {"id": "4", "title": "测试笔记4", "view_count": 1500, "category": "food", "revenue": 45.0, "engagement_rate": 6.5},
      {"id": "5", "title": "测试笔记5", "view_count": 3000, "category": "travel", "revenue": 120.0, "engagement_rate": 9.1}
    ]
  }
}`

// 大数据量基准测试JSON（100个笔记）
var largeBenchmarkJSON string

func init() {
	// 生成包含100个笔记的大数据量JSON
	var builder strings.Builder
	builder.WriteString(`{"code": 0, "success": true, "data": {"notes": [`)

	for i := 0; i < 100; i++ {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(`{
			"id": "note_` + string(rune('0'+i%10)) + `",
			"title": "基准测试笔记` + string(rune('0'+i%10)) + `",
			"view_count": ` + string(rune('0'+(i*100)%10000)) + `,
			"category": "category_` + string(rune('0'+i%5)) + `",
			"revenue": ` + string(rune('0'+(i*10)%100)) + `.5,
			"engagement_rate": ` + string(rune('0'+i%10)) + `.` + string(rune('0'+(i*3)%10)) + `,
			"status": "published",
			"type": "normal"
		}`)
	}

	builder.WriteString(`]}}`)
	largeBenchmarkJSON = builder.String()
}

// BenchmarkBasicParsing 基础解析性能基准测试
func BenchmarkBasicParsing(b *testing.B) {
	data := []byte(benchmarkJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := FromBytes(data)
		_ = node.Get("data.notes")
	}
}

// BenchmarkParsingWithDebug 带调试信息的解析性能基准测试
func BenchmarkParsingWithDebug(b *testing.B) {
	data := []byte(benchmarkJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node, _ := FromBytesWithDebug(data)
		_ = node.Get("data.notes")
	}
}

// BenchmarkCachedParsing 缓存解析性能基准测试
func BenchmarkCachedParsing(b *testing.B) {
	data := []byte(benchmarkJSON)
	cache := NewMemoryCache(100)
	EnableCaching(cache)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := FromBytesWithCache(data, 5*time.Minute)
		_ = node.Get("data.notes")
	}
}

// BenchmarkConditionalQuery 条件查询性能基准测试
func BenchmarkConditionalQuery(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, _ := notesList.Query().
			Where("view_count", ">", 1000).
			Where("category", "=", "tech").
			SortBy("view_count", "desc").
			ToSlice()
		_ = results
	}
}

// BenchmarkSimpleQuery 简单查询性能基准测试
func BenchmarkSimpleQuery(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, _ := notesList.Query().
			Where("view_count", ">", 1000).
			ToSlice()
		_ = results
	}
}

// BenchmarkComplexQuery 复杂查询性能基准测试
func BenchmarkComplexQuery(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, _ := notesList.Query().
			Where("view_count", ">", 800).
			Where("engagement_rate", ">", 6.0).
			Where("category", "!=", "life").
			SortBy("engagement_rate", "desc").
			SortBy("view_count", "desc").
			Limit(3).
			ToSlice()
		_ = results
	}
}

// BenchmarkDataAggregation 数据聚合性能基准测试
func BenchmarkDataAggregation(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats, _ := notesList.Aggregate().
			GroupBy("category").
			Count("total_count").
			Sum("revenue", "total_revenue").
			Avg("engagement_rate", "avg_engagement").
			Execute(notesList)
		_ = stats
	}
}

// BenchmarkSimpleAggregation 简单聚合性能基准测试
func BenchmarkSimpleAggregation(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats, _ := notesList.Aggregate().
			Count("total_count").
			Sum("revenue", "total_revenue").
			Execute(notesList)
		_ = stats
	}
}

// BenchmarkDataTransformation 数据变换性能基准测试
func BenchmarkDataTransformation(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))

	mapper := FieldMapper{
		Rules: map[string]string{
			"data.notes[0].title":      "note_title",
			"data.notes[0].view_count": "views",
			"data.notes[0].revenue":    "income",
		},
		DefaultValues: map[string]interface{}{
			"status": "active",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, _ := node.Transform(mapper)
		_ = result
	}
}

// BenchmarkDataValidation 数据验证性能基准测试
func BenchmarkDataValidation(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")
	firstNote := notesList.Index(0)

	validator := &DataValidator{
		Rules: map[string]ValidationRule{
			"title": {
				Required:  true,
				Type:      "string",
				MinLength: 1,
				MaxLength: 100,
			},
			"view_count": {
				Required: true,
				Type:     "number",
				Min:      0,
				Max:      100000,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, errors := firstNote.Validate(validator)
		_ = result
		_ = errors
	}
}

// BenchmarkStreamProcessing 流式处理性能基准测试
func BenchmarkStreamProcessing(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = notesList.Stream(func(note Node, index int) bool {
			_, _ = note.Get("title").String()
			_, _ = note.Get("view_count").Int()
			return true
		})
	}
}

// BenchmarkPrettyPrint 美化打印性能基准测试
func BenchmarkPrettyPrint(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")
	firstNote := notesList.Index(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output := firstNote.PrettyPrint()
		_ = output
	}
}

// BenchmarkNodeInspection 节点检查性能基准测试
func BenchmarkNodeInspection(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")
	firstNote := notesList.Index(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inspection := firstNote.Inspect()
		_ = inspection
	}
}

// BenchmarkBatchProcessing 批处理性能基准测试
func BenchmarkBatchProcessing(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor := NewBatchProcessor(3, func(nodes []Node) error {
			return nil
		})

		notesList.ArrayForEach(func(index int, note Node) bool {
			processor.Add(note)
			return true
		})
		processor.Flush()
	}
}

// BenchmarkLargeDataQuery 大数据量查询性能基准测试
func BenchmarkLargeDataQuery(b *testing.B) {
	node := FromBytes([]byte(largeBenchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, _ := notesList.Query().
			Where("engagement_rate", ">", 5.0).
			SortBy("view_count", "desc").
			Limit(10).
			ToSlice()
		_ = results
	}
}

// BenchmarkLargeDataAggregation 大数据量聚合性能基准测试
func BenchmarkLargeDataAggregation(b *testing.B) {
	node := FromBytes([]byte(largeBenchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats, _ := notesList.Aggregate().
			GroupBy("category").
			Count("count").
			Sum("revenue", "total_revenue").
			Avg("engagement_rate", "avg_engagement").
			Execute(notesList)
		_ = stats
	}
}

// BenchmarkEmptyStringHandling 空字符串处理性能基准测试
func BenchmarkEmptyStringHandling(b *testing.B) {
	emptyStringJSON := `{
		"data": {
			"notes": [
				{"title": "", "msg": "", "token": "", "source": ""},
				{"title": "", "msg": "", "token": "", "source": ""},
				{"title": "", "msg": "", "token": "", "source": ""}
			]
		}
	}`

	data := []byte(emptyStringJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := FromBytes(data)
		notesList := node.Get("data.notes")

		for j := 0; j < notesList.Len(); j++ {
			note := notesList.Index(j)
			_, _ = note.Get("title").String()
			_, _ = note.Get("msg").String()
			_, _ = note.Get("token").String()
			_, _ = note.Get("source").String()
		}
	}
}

// BenchmarkStructDecode 结构体解码性能基准测试
func BenchmarkStructDecode(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")
	firstNote := notesList.Index(0)

	type BenchNote struct {
		Id             string  `json:"id"`
		Title          string  `json:"title"`
		ViewCount      int     `json:"view_count"`
		Category       string  `json:"category"`
		Revenue        float64 `json:"revenue"`
		EngagementRate float64 `json:"engagement_rate"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var note BenchNote
		_ = firstNote.Decode(&note)
	}
}

// BenchmarkMemoryCache 内存缓存性能基准测试
func BenchmarkMemoryCache(b *testing.B) {
	cache := NewMemoryCache(100)
	testData := []byte(`{"test": "value", "number": 123}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟缓存操作
		key := "test_key"
		node := FromBytes(testData)
		cache.Set(key, node, 5*time.Minute)

		_, exists := cache.Get(key)
		_ = exists
	}
}

// BenchmarkJSONDiff JSON差异对比性能基准测试
func BenchmarkJSONDiff(b *testing.B) {
	node1 := FromBytes([]byte(benchmarkJSON))

	modifiedJSON := strings.Replace(benchmarkJSON, `"view_count": 1000`, `"view_count": 1500`, 1)
	node2 := FromBytes([]byte(modifiedJSON))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diffs := node1.Diff(node2)
		_ = diffs
	}
}

// BenchmarkQueryCount 查询计数性能基准测试
func BenchmarkQueryCount(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count, _ := notesList.Query().
			Where("view_count", ">", 800).
			Count()
		_ = count
	}
}

// BenchmarkQueryFirst 查询第一个匹配项性能基准测试
func BenchmarkQueryFirst(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		first, _ := notesList.Query().
			Where("category", "=", "tech").
			First()
		_ = first
	}
}

// BenchmarkContainsQuery 包含查询性能基准测试
func BenchmarkContainsQuery(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, _ := notesList.Query().
			WhereContains("title", "测试").
			ToSlice()
		_ = results
	}
}

// BenchmarkRangeQuery 范围查询性能基准测试
func BenchmarkRangeQuery(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, _ := notesList.Query().
			Where("view_count", ">=", 1000).
			Where("view_count", "<=", 2500).
			ToSlice()
		_ = results
	}
}

// BenchmarkMultiLevelAggregation 多级聚合性能基准测试
func BenchmarkMultiLevelAggregation(b *testing.B) {
	node := FromBytes([]byte(benchmarkJSON))
	notesList := node.Get("data.notes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats, _ := notesList.Aggregate().
			GroupBy("category").
			Count("count").
			Sum("view_count", "total_views").
			Sum("revenue", "total_revenue").
			Avg("engagement_rate", "avg_engagement").
			Min("engagement_rate", "min_engagement").
			Max("engagement_rate", "max_engagement").
			Execute(notesList)
		_ = stats
	}
}
