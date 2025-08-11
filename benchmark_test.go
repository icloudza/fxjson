package fxjson

import (
	"encoding/json"
	"testing"

	"github.com/tidwall/gjson"
)

// 测试数据
var (
	// 简单JSON
	simpleJSON = []byte(`{
		"name": "Alice",
		"age": 30,
		"active": true,
		"balance": 1234.56,
		"address": null
	}`)

	// 复杂嵌套JSON
	complexJSON = []byte(`{
		"logs": [
			{"level": "info", "msg": "start"},
			{"level": "error", "msg": "fail"}
		],
		"data": {
			"user": {
				"name": "Alice",
				"age": 30,
				"nested": {
					"flag": true,
					"items": [1, 2, 3, 4, 5],
					"deep": {
						"layer1": {
							"layer2": {
								"layer3": {
									"value": "deep value",
									"number": 9876543210.12345,
									"bool": false,
									"nullval": null,
									"array": [
										{"name": "obj1", "score": 99.9},
										{"name": "obj2", "score": 88.8},
										{"name": "obj3", "score": 77.7}
									]
								}
							}
						}
					}
				}
			}
		},
		"meta": {
			"version": "1.2.3"
		}
	}`)

	// 大型数组JSON
	largeArrayJSON = []byte(`{
		"items": [
			{"id": 1, "name": "item1", "price": 10.5},
			{"id": 2, "name": "item2", "price": 20.5},
			{"id": 3, "name": "item3", "price": 30.5},
			{"id": 4, "name": "item4", "price": 40.5},
			{"id": 5, "name": "item5", "price": 50.5},
			{"id": 6, "name": "item6", "price": 60.5},
			{"id": 7, "name": "item7", "price": 70.5},
			{"id": 8, "name": "item8", "price": 80.5},
			{"id": 9, "name": "item9", "price": 90.5},
			{"id": 10, "name": "item10", "price": 100.5}
		]
	}`)

	// 各种类型的JSON
	typesJSON = []byte(`{
		"string": "hello world",
		"number": 42,
		"float": 3.14159,
		"bool": true,
		"null": null,
		"object": {"key": "value"},
		"array": [1, 2, 3]
	}`)
)

// ==================== FromBytes 基准测试 ====================

func BenchmarkFromBytes_Simple(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FromBytes(simpleJSON)
	}
}

func BenchmarkFromBytes_Complex(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FromBytes(complexJSON)
	}
}

func BenchmarkFromBytes_LargeArray(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FromBytes(largeArrayJSON)
	}
}

// ==================== GetByPath 基准测试 ====================

func BenchmarkGetByPath_ShortPath(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(simpleJSON)
		_ = root.GetByPath("name")
	}
}

func BenchmarkGetByPath_MediumPath(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(complexJSON)
		_ = root.GetByPath("data.user.name")
	}
}

func BenchmarkGetByPath_DeepPath(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(complexJSON)
		_ = root.GetByPath("data.user.nested.deep.layer1.layer2.layer3.value")
	}
}

func BenchmarkGetByPath_ArrayAccess(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(complexJSON)
		_ = root.GetByPath("logs[0].msg")
	}
}

func BenchmarkGetByPath_DeepArrayAccess(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(complexJSON)
		_ = root.GetByPath("data.user.nested.deep.layer1.layer2.layer3.array[1].score")
	}
}

func BenchmarkGetByPath_MixedPath(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(largeArrayJSON)
		_ = root.GetByPath("items[5].name")
	}
}

// ==================== Get 基准测试 ====================

func BenchmarkGet_Simple(b *testing.B) {
	root := FromBytes(simpleJSON)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = root.Get("name")
	}
}

func BenchmarkGet_Nested(b *testing.B) {
	root := FromBytes(complexJSON)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data := root.Get("data")
		user := data.Get("user")
		_ = user.Get("name")
	}
}

func BenchmarkGet_NotFound(b *testing.B) {
	root := FromBytes(simpleJSON)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = root.Get("nonexistent")
	}
}

// ==================== String 基准测试 ====================

func BenchmarkString_Short(b *testing.B) {
	root := FromBytes(simpleJSON)
	node := root.GetByPath("name")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = node.String()
	}
}

func BenchmarkString_Long(b *testing.B) {
	root := FromBytes(complexJSON)
	node := root.GetByPath("data.user.nested.deep.layer1.layer2.layer3.value")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = node.String()
	}
}

func BenchmarkString_NotString(b *testing.B) {
	root := FromBytes(simpleJSON)
	node := root.GetByPath("age")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = node.String() // 应该返回空字符串
	}
}

// ==================== Type 和 Kind 基准测试 ====================

func BenchmarkType(b *testing.B) {
	root := FromBytes(typesJSON)
	nodes := []Node{
		root.GetByPath("string"),
		root.GetByPath("number"),
		root.GetByPath("bool"),
		root.GetByPath("null"),
		root.GetByPath("object"),
		root.GetByPath("array"),
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, node := range nodes {
			_ = node.Type()
		}
	}
}

func BenchmarkKind(b *testing.B) {
	root := FromBytes(typesJSON)
	nodes := []Node{
		root.GetByPath("string"),
		root.GetByPath("number"),
		root.GetByPath("bool"),
		root.GetByPath("null"),
		root.GetByPath("object"),
		root.GetByPath("array"),
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, node := range nodes {
			_ = node.Kind()
		}
	}
}

// ==================== 对比基准测试 ====================

// 与 encoding/json 对比
func BenchmarkVsEncodingJSON_Simple(b *testing.B) {
	b.Run("zeronode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			root := FromBytes(simpleJSON)
			val := root.Get("name")
			_ = val.String()
		}
	})

	b.Run("encoding/json", func(b *testing.B) {
		b.ReportAllocs()
		var data map[string]interface{}
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(simpleJSON, &data)
			_ = data["name"].(string)
		}
	})
}

func BenchmarkVsEncodingJSON_DeepPath(b *testing.B) {
	targetPath := "data.user.nested.deep.layer1.layer2.layer3.value"

	b.Run("zeronode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			root := FromBytes(complexJSON)
			val := root.GetByPath(targetPath)
			_ = val.String()
		}
	})

	b.Run("encoding/json", func(b *testing.B) {
		type L3 struct {
			Value string `json:"value"`
		}
		type L2 struct {
			Layer3 L3 `json:"layer3"`
		}
		type L1 struct {
			Layer2 L2 `json:"layer2"`
		}
		type Deep struct {
			Layer1 L1 `json:"layer1"`
		}
		type Nested struct {
			Deep Deep `json:"deep"`
		}
		type User struct {
			Nested Nested `json:"nested"`
		}
		type Data struct {
			User User `json:"user"`
		}
		type Root struct {
			Data Data `json:"data"`
		}

		b.ReportAllocs()
		var root Root
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(complexJSON, &root)
			_ = root.Data.User.Nested.Deep.Layer1.Layer2.Layer3.Value
		}
	})
}

// 与 gjson 对比
func BenchmarkVsGJSON_Simple(b *testing.B) {
	b.Run("zeronode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			root := FromBytes(simpleJSON)
			val := root.Get("name")
			_ = val.String()
		}
	})

	b.Run("gjson", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			val := gjson.GetBytes(simpleJSON, "name")
			_ = val.String()
		}
	})
}

func BenchmarkVsGJSON_DeepPath(b *testing.B) {
	targetPath := "data.user.nested.deep.layer1.layer2.layer3.value"

	b.Run("zeronode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			root := FromBytes(complexJSON)
			val := root.GetByPath(targetPath)
			_ = val.String()
		}
	})

	b.Run("gjson", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			val := gjson.GetBytes(complexJSON, targetPath)
			_ = val.String()
		}
	})
}

func BenchmarkVsGJSON_ArrayAccess(b *testing.B) {
	b.Run("zeronode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			root := FromBytes(largeArrayJSON)
			val := root.GetByPath("items[5].name")
			_ = val.String()
		}
	})

	b.Run("gjson", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			val := gjson.GetBytes(largeArrayJSON, "items.5.name")
			_ = val.String()
		}
	})
}

// ==================== 压力测试 ====================

func BenchmarkStress_ManyFields(b *testing.B) {
	// 创建有很多字段的JSON
	manyFieldsJSON := []byte(`{
		"field1": "value1", "field2": "value2", "field3": "value3",
		"field4": "value4", "field5": "value5", "field6": "value6",
		"field7": "value7", "field8": "value8", "field9": "value9",
		"field10": "value10", "field11": "value11", "field12": "value12",
		"field13": "value13", "field14": "value14", "field15": "value15",
		"field16": "value16", "field17": "value17", "field18": "value18",
		"field19": "value19", "field20": "value20"
	}`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(manyFieldsJSON)
		_ = root.Get("field20") // 最后一个字段
	}
}

func BenchmarkStress_DeepNesting(b *testing.B) {
	// 创建深度嵌套的JSON
	deepJSON := []byte(`{
		"l1": {
			"l2": {
				"l3": {
					"l4": {
						"l5": {
							"l6": {
								"l7": {
									"l8": {
										"l9": {
											"l10": "deep value"
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(deepJSON)
		_ = root.GetByPath("l1.l2.l3.l4.l5.l6.l7.l8.l9.l10")
	}
}

func BenchmarkStress_LargeArray(b *testing.B) {
	// 创建大数组JSON
	largeArr := []byte(`{"arr":[`)
	for i := 0; i < 100; i++ {
		if i > 0 {
			largeArr = append(largeArr, ',')
		}
		largeArr = append(largeArr, []byte(`{"id":`+string(rune(i))+`,"val":"test"}`)...)
	}
	largeArr = append(largeArr, []byte(`]}`)...)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(largeArr)
		_ = root.GetByPath("arr[99].val")
	}
}

// ==================== 边界情况基准测试 ====================

func BenchmarkEdgeCase_EmptyJSON(b *testing.B) {
	emptyJSON := []byte(`{}`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(emptyJSON)
		_ = root.Get("nonexistent")
	}
}

func BenchmarkEdgeCase_InvalidPath(b *testing.B) {
	root := FromBytes(complexJSON)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = root.GetByPath("invalid.path.that.does.not.exist")
	}
}

func BenchmarkEdgeCase_SpecialChars(b *testing.B) {
	specialJSON := []byte(`{
		"key-with-dash": "value1",
		"key_with_underscore": "value2",
		"key.with.dot": "value3",
		"key with space": "value4"
	}`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := FromBytes(specialJSON)
		_ = root.Get("key-with-dash")
	}
}
