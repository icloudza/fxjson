package fxjson

import (
	"encoding/json"
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
