package fxjson

import (
	"encoding/json"
	"fmt"
	"testing"
)

var sampleJSON = []byte(`{
	"id": 1234567890123456789,
	"name": "Alice",
	"active": true,
	"score": 99.99,
	"tags": ["go", "json", "benchmark"],
	"meta": {
		"age": 30,
		"nullVal": null,
		"nested": {
			"flag": false,
			"numbers": [1, 2, 3, 4, 5]
		}
	}
}`)

// ===== Get =====
func BenchmarkGet_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON)
	for i := 0; i < b.N; i++ {
		_ = node.Get("name")
	}
}

func BenchmarkGet_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		_ = m["name"]
	}
}

// ===== GetPath =====
func BenchmarkGetPath_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON)
	for i := 0; i < b.N; i++ {
		_ = node.GetPath("meta.nested.flag")
	}
}

func BenchmarkGetPath_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		_ = m["meta"].(map[string]any)["nested"].(map[string]any)["flag"]
	}
}

// ===== Int =====
func BenchmarkInt_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("id")
	for i := 0; i < b.N; i++ {
		_, _ = node.Int()
	}
}

func BenchmarkInt_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		_ = int64(m["id"].(float64))
	}
}

// ===== Float =====
func BenchmarkFloat_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("score")
	for i := 0; i < b.N; i++ {
		_, _ = node.Float()
	}
}

func BenchmarkFloat_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		_ = m["score"].(float64)
	}
}

// ===== Bool =====
func BenchmarkBool_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("active")
	for i := 0; i < b.N; i++ {
		_, _ = node.Bool()
	}
}

func BenchmarkBool_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		_ = m["active"].(bool)
	}
}

// ===== String =====
func BenchmarkString_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("name")
	for i := 0; i < b.N; i++ {
		_, _ = node.String()
	}
}

func BenchmarkString_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		_ = m["name"].(string)
	}
}

// ===== Len =====
func BenchmarkLen_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("tags")
	for i := 0; i < b.N; i++ {
		_ = node.Len()
	}
}

func BenchmarkLen_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		_ = len(m["tags"].([]any))
	}
}

// ===== Keys =====
func BenchmarkKeys_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("meta")
	for i := 0; i < b.N; i++ {
		_ = node.Keys()
	}
}

func BenchmarkKeys_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		keys := make([]string, 0, len(m["meta"].(map[string]any)))
		for k := range m["meta"].(map[string]any) {
			keys = append(keys, k)
		}
		_ = keys
	}
}

// ===== Index =====
func BenchmarkIndex_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("tags")
	for i := 0; i < b.N; i++ {
		_ = node.Index(1)
	}
}

func BenchmarkIndex_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		_ = m["tags"].([]any)[1]
	}
}

// ===== Exists =====
func BenchmarkExists_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("name")
	for i := 0; i < b.N; i++ {
		_ = node.Exists()
	}
}

func BenchmarkExists_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		_, _ = m["name"]
	}
}

// ===== JsonWithParam 格式化输出 =====
func BenchmarkJsonWithParam_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = node.Json()
	}
}

// ===== ForEach 遍历测试 =====
func BenchmarkForEach_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("meta")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.ForEach(func(key string, value Node) bool {
			return true
		})
	}
}

func BenchmarkForEach_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		for _ = range m["meta"].(map[string]any) {
		}
	}
}

// ===== ArrayForEach 数组遍历测试 =====
func BenchmarkArrayForEach_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("tags")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.ArrayForEach(func(index int, value Node) bool {
			return true
		})
	}
}

func BenchmarkArrayForEach_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		for _ = range m["tags"].([]any) {
		}
	}
}

// ===== Walk 深度遍历测试 =====
func BenchmarkWalk_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.Walk(func(path string, node Node) bool {
			return true
		})
	}
}

func BenchmarkWalk_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(sampleJSON, &m)
		walkStd(m, "")
	}
}

func walkStd(v any, path string) {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			childPath := path
			if childPath != "" {
				childPath += "."
			}
			childPath += k
			walkStd(child, childPath)
		}
	case []any:
		for i, child := range val {
			childPath := path + fmt.Sprintf("[%d]", i)
			walkStd(child, childPath)
		}
	}
}

// ===== 复杂遍历测试（模拟实际使用场景）=====
var complexJSON = []byte(`{
	"data": {
		"users": [
			{"id": 1, "name": "Alice", "tags": ["admin", "active"]},
			{"id": 2, "name": "Bob", "tags": ["user", "inactive"]},
			{"id": 3, "name": "Charlie", "tags": ["user", "active"]},
			{"id": 4, "name": "David", "tags": ["admin", "active"]},
			{"id": 5, "name": "Eve", "tags": ["user", "active"]}
		],
		"meta": {
			"total": 5,
			"page": 1,
			"limit": 10
		}
	}
}`)

func BenchmarkComplexTraversal_fxjson(b *testing.B) {
	node := FromBytes(complexJSON)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users := node.Get("data").Get("users")
		users.ArrayForEach(func(index int, user Node) bool {
			_ = user.Get("name")
			_ = user.Get("id")
			tags := user.Get("tags")
			tags.ArrayForEach(func(idx int, tag Node) bool {
				_ = tag
				return true
			})
			return true
		})
	}
}

func BenchmarkComplexTraversal_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(complexJSON, &m)
		users := m["data"].(map[string]any)["users"].([]any)
		for _, userAny := range users {
			user := userAny.(map[string]any)
			_ = user["name"]
			_ = user["id"]
			tags := user["tags"].([]any)
			for _, tag := range tags {
				_ = tag
			}
		}
	}
}

// ===== 大规模遍历测试 =====
var largeJSON = []byte(`{
	"users": [
		{"id": 1, "name": "Alice", "email": "alice@example.com", "age": 30, "active": true, "tags": ["admin", "developer", "active"], "meta": {"department": "engineering", "level": "senior", "projects": ["project1", "project2"]}},
		{"id": 2, "name": "Bob", "email": "bob@example.com", "age": 25, "active": false, "tags": ["user", "tester", "inactive"], "meta": {"department": "qa", "level": "junior", "projects": ["project3"]}},
		{"id": 3, "name": "Charlie", "email": "charlie@example.com", "age": 35, "active": true, "tags": ["admin", "manager", "active"], "meta": {"department": "management", "level": "senior", "projects": ["project1", "project4", "project5"]}},
		{"id": 4, "name": "David", "email": "david@example.com", "age": 28, "active": true, "tags": ["developer", "active"], "meta": {"department": "engineering", "level": "mid", "projects": ["project2", "project6"]}},
		{"id": 5, "name": "Eve", "email": "eve@example.com", "age": 32, "active": true, "tags": ["designer", "active"], "meta": {"department": "design", "level": "senior", "projects": ["project7", "project8"]}}
	],
	"products": [
		{"id": 101, "name": "Product A", "price": 99.99, "category": "electronics", "in_stock": true, "reviews": [{"rating": 5, "comment": "Great!"}, {"rating": 4, "comment": "Good"}]},
		{"id": 102, "name": "Product B", "price": 149.99, "category": "books", "in_stock": false, "reviews": [{"rating": 3, "comment": "OK"}, {"rating": 5, "comment": "Excellent!"}]},
		{"id": 103, "name": "Product C", "price": 199.99, "category": "clothing", "in_stock": true, "reviews": [{"rating": 4, "comment": "Nice"}, {"rating": 2, "comment": "Could be better"}]}
	]
}`)

func BenchmarkLargeDataTraversal_fxjson(b *testing.B) {
	node := FromBytes(largeJSON)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		node.Walk(func(path string, n Node) bool {
			count++
			return true
		})
		_ = count
	}
}

func BenchmarkLargeDataTraversal_std(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = json.Unmarshal(largeJSON, &m)
		count := 0
		walkStdCount(m, "", &count)
		_ = count
	}
}

func walkStdCount(v any, path string, count *int) {
	*count++
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			childPath := path
			if childPath != "" {
				childPath += "."
			}
			childPath += k
			walkStdCount(child, childPath, count)
		}
	case []any:
		for i, child := range val {
			childPath := path + fmt.Sprintf("[%d]", i)
			walkStdCount(child, childPath, count)
		}
	}
}
