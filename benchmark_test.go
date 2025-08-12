package fxjson

import (
	"testing"

	"github.com/tidwall/gjson"
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

// ===== Get / GetPath =====
func BenchmarkGet_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON)
	for i := 0; i < b.N; i++ {
		_ = node.Get("name")
	}
}

func BenchmarkGet_gjson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = gjson.GetBytes(sampleJSON, "name")
	}
}

func BenchmarkGetPath_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON)
	for i := 0; i < b.N; i++ {
		_ = node.GetPath("meta.nested.flag")
	}
}

func BenchmarkGetPath_gjson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = gjson.GetBytes(sampleJSON, "meta.nested.flag")
	}
}

// ===== Int / Float / Bool / String =====
func BenchmarkInt_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("id")
	for i := 0; i < b.N; i++ {
		_, _ = node.Int()
	}
}

func BenchmarkInt_gjson(b *testing.B) {
	r := gjson.GetBytes(sampleJSON, "id")
	for i := 0; i < b.N; i++ {
		_ = r.Int()
	}
}

func BenchmarkFloat_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("score")
	for i := 0; i < b.N; i++ {
		_, _ = node.Float()
	}
}

func BenchmarkFloat_gjson(b *testing.B) {
	r := gjson.GetBytes(sampleJSON, "score")
	for i := 0; i < b.N; i++ {
		_ = r.Float()
	}
}

func BenchmarkBool_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("active")
	for i := 0; i < b.N; i++ {
		_, _ = node.Bool()
	}
}

func BenchmarkBool_gjson(b *testing.B) {
	r := gjson.GetBytes(sampleJSON, "active")
	for i := 0; i < b.N; i++ {
		_ = r.Bool()
	}
}

func BenchmarkString_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("name")
	for i := 0; i < b.N; i++ {
		_ = node.String()
	}
}

func BenchmarkString_gjson(b *testing.B) {
	r := gjson.GetBytes(sampleJSON, "name")
	for i := 0; i < b.N; i++ {
		_ = r.String()
	}
}

func BenchmarkNumStr_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("score")
	for i := 0; i < b.N; i++ {
		_ = node.NumStr()
	}
}

func BenchmarkNumStr_gjson(b *testing.B) {
	r := gjson.GetBytes(sampleJSON, "score")
	for i := 0; i < b.N; i++ {
		_ = r.Raw
	}
}

// ===== Len / Keys =====
func BenchmarkLen_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("tags")
	for i := 0; i < b.N; i++ {
		_ = node.Len()
	}
}

func BenchmarkLen_gjson(b *testing.B) {
	r := gjson.GetBytes(sampleJSON, "tags")
	for i := 0; i < b.N; i++ {
		_ = len(r.Array())
	}
}

func BenchmarkKeys_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("meta")
	for i := 0; i < b.N; i++ {
		_ = node.Keys()
	}
}

func BenchmarkKeys_gjson(b *testing.B) {
	r := gjson.GetBytes(sampleJSON, "meta")
	for i := 0; i < b.N; i++ {
		_ = r.Map()
	}
}

// ===== Index =====
func BenchmarkIndex_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("tags")
	for i := 0; i < b.N; i++ {
		_ = node.Index(1)
	}
}

func BenchmarkIndex_gjson(b *testing.B) {
	r := gjson.GetBytes(sampleJSON, "tags.1")
	for i := 0; i < b.N; i++ {
		_ = r
	}
}

// ===== Exists / IsNull =====
func BenchmarkExists_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("name")
	for i := 0; i < b.N; i++ {
		_ = node.Exists()
	}
}

func BenchmarkExists_gjson(b *testing.B) {
	r := gjson.GetBytes(sampleJSON, "name")
	for i := 0; i < b.N; i++ {
		_ = r.Exists()
	}
}

func BenchmarkIsNull_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON).Get("meta.nullVal")
	for i := 0; i < b.N; i++ {
		_ = node.IsNull()
	}
}

func BenchmarkIsNull_gjson(b *testing.B) {
	r := gjson.GetBytes(sampleJSON, "meta.nullVal")
	for i := 0; i < b.N; i++ {
		_ = r.Type == gjson.Null
	}
}

// ===== Decode =====
func BenchmarkDecode_fxjson(b *testing.B) {
	node := FromBytes(sampleJSON)
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = node.Decode(&m)
	}
}

func BenchmarkDecode_gjson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]any
		_ = gjson.GetBytes(sampleJSON, "").Value() // gjson 无直接 decode 方法
		_ = m
	}
}
