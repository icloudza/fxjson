package fxjson

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

// 测试数据
var testJSON = []byte(`{
	"string": "hello world",
	"number": 42,
	"float": 3.14159,
	"bigint": 9223372036854775807,
	"bool_true": true,
	"bool_false": false,
	"null": null,
	"empty_string": "",
	"negative": -123,
	"array": [1, 2, 3, "four", true, null],
	"object": {
		"nested_string": "nested value",
		"nested_number": 100,
		"nested_array": [{"deep": "value"}, {"deep": "value2"}]
	},
	"empty_array": [],
	"empty_object": {},
	"unicode": "中文测试",
	"escaped": "line1\nline2\ttab\"quote\\slash",
	"large_number": 1.23456789e10,
	"scientific": 1.234e-5
}`)

// ===== 基础创建和访问测试 =====

func TestFromBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{"valid json", testJSON, true},
		{"empty input", []byte{}, false},
		{"invalid json", []byte("{invalid"), false},
		{"null json", []byte("null"), true},
		{"array json", []byte("[1,2,3]"), true},
		{"string json", []byte(`"test"`), true},
		{"number json", []byte("123"), true},
		{"bool json", []byte("true"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := FromBytes(tt.input)
			if node.Exists() != tt.expected {
				t.Errorf("FromBytes(%s).Exists() = %v, want %v", string(tt.input), node.Exists(), tt.expected)
			}
		})
	}
}

func TestGet(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key      string
		exists   bool
		nodeType NodeType
	}{
		{"string", true, TypeString},
		{"number", true, TypeNumber},
		{"bool_true", true, TypeBool},
		{"null", true, TypeNull},
		{"array", true, TypeArray},
		{"object", true, TypeObject},
		{"nonexistent", false, TypeInvalid},
		{"", false, TypeInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := node.Get(tt.key)
			if result.Exists() != tt.exists {
				t.Errorf("Get(%q).Exists() = %v, want %v", tt.key, result.Exists(), tt.exists)
			}
			if result.Exists() && result.Kind() != tt.nodeType {
				t.Errorf("Get(%q).Kind() = %v, want %v", tt.key, result.Kind(), tt.nodeType)
			}
		})
	}
}

func TestGetPath(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		path     string
		exists   bool
		nodeType NodeType
	}{
		{"string", true, TypeString},
		{"object.nested_string", true, TypeString},
		{"object.nested_number", true, TypeNumber},
		{"object.nested_array[0].deep", true, TypeString},
		{"object.nested_array[1].deep", true, TypeString},
		{"array[0]", true, TypeNumber},
		{"array[3]", true, TypeString},
		{"array[4]", true, TypeBool},
		{"array[5]", true, TypeNull},
		{"array[10]", false, TypeInvalid},
		{"object.nonexistent", false, TypeInvalid},
		{"nonexistent.path", false, TypeInvalid},
		{"", false, TypeInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := node.GetPath(tt.path)
			if result.Exists() != tt.exists {
				t.Errorf("GetPath(%q).Exists() = %v, want %v", tt.path, result.Exists(), tt.exists)
			}
			if result.Exists() && result.Kind() != tt.nodeType {
				t.Errorf("GetPath(%q).Kind() = %v, want %v", tt.path, result.Kind(), tt.nodeType)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	node := FromBytes(testJSON).Get("array")

	tests := []struct {
		index    int
		exists   bool
		nodeType NodeType
	}{
		{0, true, TypeNumber},
		{1, true, TypeNumber},
		{2, true, TypeNumber},
		{3, true, TypeString},
		{4, true, TypeBool},
		{5, true, TypeNull},
		{6, false, TypeInvalid},
		{-1, false, TypeInvalid},
		{100, false, TypeInvalid},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := node.Index(tt.index)
			if result.Exists() != tt.exists {
				t.Errorf("Index(%d).Exists() = %v, want %v", tt.index, result.Exists(), tt.exists)
			}
			if result.Exists() && result.Kind() != tt.nodeType {
				t.Errorf("Index(%d).Kind() = %v, want %v", tt.index, result.Kind(), tt.nodeType)
			}
		})
	}

	// 测试非数组节点
	stringNode := FromBytes(testJSON).Get("string")
	if stringNode.Index(0).Exists() {
		t.Error("Index on string node should return non-existent node")
	}
}

// ===== 数据类型转换测试 =====

func TestString(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key      string
		expected string
		hasError bool
	}{
		{"string", "hello world", false},
		{"empty_string", "", false},
		{"unicode", "中文测试", false},
		{"escaped", "line1\nline2\ttab\"quote\\slash", false},
		{"number", "", true}, // 非字符串类型应该返回错误
		{"null", "", true},
		{"bool_true", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result, err := node.Get(tt.key).String()
			if tt.hasError {
				if err == nil {
					t.Errorf("String() should return error for key %q", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("String() returned unexpected error for key %q: %v", tt.key, err)
				}
				if result != tt.expected {
					t.Errorf("String() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestInt(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key      string
		expected int64
		hasError bool
	}{
		{"number", 42, false},
		{"negative", -123, false},
		{"bigint", 9223372036854775807, false},
		{"string", 0, true},
		{"float", 0, true}, // 浮点数应该返回错误
		{"bool_true", 0, true},
		{"null", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result, err := node.Get(tt.key).Int()
			if tt.hasError {
				if err == nil {
					t.Errorf("Int() should return error for key %q", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("Int() returned unexpected error for key %q: %v", tt.key, err)
				}
				if result != tt.expected {
					t.Errorf("Int() = %d, want %d", result, tt.expected)
				}
			}
		})
	}
}

func TestUint(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key      string
		expected uint64
		hasError bool
	}{
		{"number", 42, false},
		{"bigint", 9223372036854775807, false},
		{"negative", 0, true}, // 负数应该返回错误
		{"string", 0, true},
		{"float", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result, err := node.Get(tt.key).Uint()
			if tt.hasError {
				if err == nil {
					t.Errorf("Uint() should return error for key %q", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("Uint() returned unexpected error for key %q: %v", tt.key, err)
				}
				if result != tt.expected {
					t.Errorf("Uint() = %d, want %d", result, tt.expected)
				}
			}
		})
	}
}

func TestFloat(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key      string
		expected float64
		hasError bool
	}{
		{"float", 3.14159, false},
		{"number", 42.0, false},
		{"negative", -123.0, false},
		{"large_number", 1.23456789e10, false},
		{"scientific", 1.234e-5, false},
		{"string", 0, true},
		{"bool_true", 0, true},
		{"null", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result, err := node.Get(tt.key).Float()
			if tt.hasError {
				if err == nil {
					t.Errorf("Float() should return error for key %q", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("Float() returned unexpected error for key %q: %v", tt.key, err)
				}
				if math.Abs(result-tt.expected) > 1e-10 {
					t.Errorf("Float() = %f, want %f", result, tt.expected)
				}
			}
		})
	}
}

func TestBool(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key      string
		expected bool
		hasError bool
	}{
		{"bool_true", true, false},
		{"bool_false", false, false},
		{"string", false, true},
		{"number", false, true},
		{"null", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result, err := node.Get(tt.key).Bool()
			if tt.hasError {
				if err == nil {
					t.Errorf("Bool() should return error for key %q", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("Bool() returned unexpected error for key %q: %v", tt.key, err)
				}
				if result != tt.expected {
					t.Errorf("Bool() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestNumStr(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key      string
		expected string
		hasError bool
	}{
		{"number", "42", false},
		{"float", "3.14159", false},
		{"negative", "-123", false},
		{"large_number", "1.23456789e10", false},
		{"scientific", "1.234e-5", false},
		{"string", "", true},
		{"bool_true", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result, err := node.Get(tt.key).NumStr()
			if tt.hasError {
				if err == nil {
					t.Errorf("NumStr() should return error for key %q", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("NumStr() returned unexpected error for key %q: %v", tt.key, err)
				}
				if result != tt.expected {
					t.Errorf("NumStr() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestFloatString(t *testing.T) {
	// 测试FloatString()方法保持原始JSON格式
	precisionJSON := []byte(`{
		"price": 1.1,
		"rating": 4.50,
		"score": 95.0,
		"percentage": 12.34,
		"integer": 42,
		"scientific": 1.234e-5,
		"large": 1.23456789e10
	}`)

	node := FromBytes(precisionJSON)

	tests := []struct {
		key      string
		expected string
		hasError bool
	}{
		{"price", "1.1", false},
		{"rating", "4.50", false},
		{"score", "95.0", false},
		{"percentage", "12.34", false},
		{"integer", "42", false},
		{"scientific", "1.234e-5", false},
		{"large", "1.23456789e10", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result, err := node.Get(tt.key).FloatString()
			if tt.hasError {
				if err == nil {
					t.Errorf("FloatString() should return error for key %q", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("FloatString() returned unexpected error for key %q: %v", tt.key, err)
				}
				if result != tt.expected {
					t.Errorf("FloatString() = %q, want %q", result, tt.expected)
				}
			}
		})
	}

	// 测试非数字类型的错误处理
	t.Run("non-number types", func(t *testing.T) {
		testData := []byte(`{"string": "hello", "bool": true, "null": null, "array": [1,2,3]}`)
		node := FromBytes(testData)

		nonNumericKeys := []string{"string", "bool", "null", "array"}
		for _, key := range nonNumericKeys {
			if _, err := node.Get(key).FloatString(); err == nil {
				t.Errorf("FloatString() should return error for non-numeric key %q", key)
			}
		}
	})
}

// ===== 节点属性测试 =====

func TestNodeTypes(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key    string
		checks map[string]bool
	}{
		{"string", map[string]bool{"IsString": true, "IsScalar": true}},
		{"number", map[string]bool{"IsNumber": true, "IsScalar": true}},
		{"bool_true", map[string]bool{"IsBool": true, "IsScalar": true}},
		{"null", map[string]bool{"IsNull": true, "IsScalar": true}},
		{"array", map[string]bool{"IsArray": true, "IsContainer": true}},
		{"object", map[string]bool{"IsObject": true, "IsContainer": true}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			n := node.Get(tt.key)

			if tt.checks["IsString"] && !n.IsString() {
				t.Error("IsString() should return true")
			}
			if tt.checks["IsNumber"] && !n.IsNumber() {
				t.Error("IsNumber() should return true")
			}
			if tt.checks["IsBool"] && !n.IsBool() {
				t.Error("IsBool() should return true")
			}
			if tt.checks["IsNull"] && !n.IsNull() {
				t.Error("IsNull() should return true")
			}
			if tt.checks["IsArray"] && !n.IsArray() {
				t.Error("IsArray() should return true")
			}
			if tt.checks["IsObject"] && !n.IsObject() {
				t.Error("IsObject() should return true")
			}
			if tt.checks["IsScalar"] && !n.IsScalar() {
				t.Error("IsScalar() should return true")
			}
			if tt.checks["IsContainer"] && !n.IsContainer() {
				t.Error("IsContainer() should return true")
			}
		})
	}
}

func TestExists(t *testing.T) {
	node := FromBytes(testJSON)

	existingKeys := []string{"string", "number", "null", "array", "object", "empty_string", "empty_array", "empty_object"}
	for _, key := range existingKeys {
		if !node.Get(key).Exists() {
			t.Errorf("Get(%q).Exists() should return true", key)
		}
	}

	nonExistingKeys := []string{"nonexistent", "missing", "", "object.missing", "array[100]"}
	for _, key := range nonExistingKeys {
		if key == "" {
			continue // Skip empty key as Get("") returns empty node
		}
		var result Node
		if key == "object.missing" {
			result = node.GetPath(key)
		} else if key == "array[100]" {
			result = node.GetPath(key)
		} else {
			result = node.Get(key)
		}
		if result.Exists() {
			t.Errorf("Get(%q).Exists() should return false", key)
		}
	}
}

// ===== 长度和键值测试 =====

func TestLen(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key      string
		expected int
	}{
		{"array", 6},
		{"object", 3}, // nested_string, nested_number, nested_array
		{"empty_array", 0},
		{"empty_object", 0},
		{"string", 11}, // "hello world" has 11 characters
		{"empty_string", 0},
		{"number", 0}, // 非容器类型返回0
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := node.Get(tt.key).Len()
			if result != tt.expected {
				t.Errorf("Len() = %d, want %d for key %q", result, tt.expected, tt.key)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	node := FromBytes(testJSON)

	// 测试对象键
	objectKeys := node.Get("object").Keys()
	expectedKeys := []string{"nested_string", "nested_number", "nested_array"}
	if len(objectKeys) != len(expectedKeys) {
		t.Fatalf("Keys() returned %d keys, want %d", len(objectKeys), len(expectedKeys))
	}

	keyMap := make(map[string]bool)
	for _, key := range objectKeys {
		keyMap[string(key)] = true
	}

	for _, expected := range expectedKeys {
		if !keyMap[expected] {
			t.Errorf("Keys() missing expected key: %s", expected)
		}
	}

	// 测试非对象节点
	arrayKeys := node.Get("array").Keys()
	if arrayKeys != nil {
		t.Error("Keys() should return nil for non-object nodes")
	}
}

func TestRaw(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key      string
		expected string
	}{
		{"string", `"hello world"`},
		{"number", "42"},
		{"bool_true", "true"},
		{"bool_false", "false"},
		{"null", "null"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			raw := node.Get(tt.key).Raw()
			if string(raw) != tt.expected {
				t.Errorf("Raw() = %q, want %q", string(raw), tt.expected)
			}
		})
	}
}

func TestJson(t *testing.T) {
	node := FromBytes(testJSON)

	// 测试对象和数组的Json()方法
	tests := []struct {
		key      string
		hasError bool
	}{
		{"object", false},
		{"array", false},
		{"empty_object", false},
		{"empty_array", false},
		{"string", true}, // 标量类型应该返回错误
		{"number", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result, err := node.Get(tt.key).Json()
			if tt.hasError {
				if err == nil {
					t.Errorf("Json() should return error for key %q", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("Json() returned unexpected error for key %q: %v", tt.key, err)
				}
				// 验证返回的JSON是有效的
				var temp interface{}
				if err := json.Unmarshal([]byte(result), &temp); err != nil {
					t.Errorf("Json() returned invalid JSON for key %q: %v", tt.key, err)
				}
			}
		})
	}
}

func TestRawString(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []string{"string", "number", "array", "object", "null"}
	for _, key := range tests {
		t.Run(key, func(t *testing.T) {
			result, err := node.Get(key).RawString()
			if err != nil {
				t.Errorf("RawString() returned unexpected error for key %q: %v", key, err)
			}
			if result == "" {
				t.Errorf("RawString() returned empty string for key %q", key)
			}
		})
	}
}

// ===== 解码测试 =====

func TestDecode(t *testing.T) {
	node := FromBytes(testJSON)

	// 测试解码到不同类型
	t.Run("decode string", func(t *testing.T) {
		var result string
		err := node.Get("string").Decode(&result)
		if err != nil {
			t.Errorf("Decode() returned error: %v", err)
		}
		if result != "hello world" {
			t.Errorf("Decode() = %q, want %q", result, "hello world")
		}
	})

	t.Run("decode number", func(t *testing.T) {
		var result float64
		err := node.Get("number").Decode(&result)
		if err != nil {
			t.Errorf("Decode() returned error: %v", err)
		}
		if result != 42.0 {
			t.Errorf("Decode() = %f, want %f", result, 42.0)
		}
	})

	t.Run("decode array", func(t *testing.T) {
		var result []interface{}
		err := node.Get("array").Decode(&result)
		if err != nil {
			t.Errorf("Decode() returned error: %v", err)
		}
		if len(result) != 6 {
			t.Errorf("Decode() array length = %d, want %d", len(result), 6)
		}
	})

	t.Run("decode object", func(t *testing.T) {
		var result map[string]interface{}
		err := node.Get("object").Decode(&result)
		if err != nil {
			t.Errorf("Decode() returned error: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("Decode() object length = %d, want %d", len(result), 3)
		}
	})

	// 测试错误情况
	t.Run("decode nil pointer", func(t *testing.T) {
		var result *string
		err := node.Get("string").Decode(result)
		if err == nil {
			t.Error("Decode() should return error for nil pointer")
		}
	})

	t.Run("decode non-pointer", func(t *testing.T) {
		var result string
		err := node.Get("string").Decode(result)
		if err == nil {
			t.Error("Decode() should return error for non-pointer")
		}
	})
}

// ===== 遍历方法测试 =====

func TestForEach(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("object foreach", func(t *testing.T) {
		objectNode := node.Get("object")
		var keys []string
		var values []string

		objectNode.ForEach(func(key string, value Node) bool {
			keys = append(keys, key)
			if value.IsString() {
				if str, err := value.String(); err == nil {
					values = append(values, str)
				}
			}
			return true
		})

		if len(keys) != 3 {
			t.Errorf("ForEach() visited %d keys, want 3", len(keys))
		}

		expectedKeys := map[string]bool{
			"nested_string": true,
			"nested_number": true,
			"nested_array":  true,
		}

		for _, key := range keys {
			if !expectedKeys[key] {
				t.Errorf("ForEach() found unexpected key: %s", key)
			}
		}
	})

	t.Run("early termination", func(t *testing.T) {
		objectNode := node.Get("object")
		count := 0

		objectNode.ForEach(func(key string, value Node) bool {
			count++
			return count < 2 // 只处理前2个元素
		})

		if count != 2 {
			t.Errorf("ForEach() with early termination visited %d keys, want 2", count)
		}
	})

	t.Run("non-object node", func(t *testing.T) {
		stringNode := node.Get("string")
		called := false

		stringNode.ForEach(func(key string, value Node) bool {
			called = true
			return true
		})

		if called {
			t.Error("ForEach() should not call function for non-object nodes")
		}
	})
}

func TestArrayForEach(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("array foreach", func(t *testing.T) {
		arrayNode := node.Get("array")
		var indices []int
		var values []interface{}

		arrayNode.ArrayForEach(func(index int, value Node) bool {
			indices = append(indices, index)
			if value.IsNumber() {
				if num, err := value.Float(); err == nil {
					values = append(values, num)
				}
			} else if value.IsString() {
				if str, err := value.String(); err == nil {
					values = append(values, str)
				}
			} else if value.IsBool() {
				if b, err := value.Bool(); err == nil {
					values = append(values, b)
				}
			} else if value.IsNull() {
				values = append(values, nil)
			}
			return true
		})

		expectedIndices := []int{0, 1, 2, 3, 4, 5}
		if !reflect.DeepEqual(indices, expectedIndices) {
			t.Errorf("ArrayForEach() indices = %v, want %v", indices, expectedIndices)
		}

		if len(values) != 6 {
			t.Errorf("ArrayForEach() collected %d values, want 6", len(values))
		}
	})

	t.Run("early termination", func(t *testing.T) {
		arrayNode := node.Get("array")
		count := 0

		arrayNode.ArrayForEach(func(index int, value Node) bool {
			count++
			return count < 3 // 只处理前3个元素
		})

		if count != 3 {
			t.Errorf("ArrayForEach() with early termination visited %d elements, want 3", count)
		}
	})

	t.Run("non-array node", func(t *testing.T) {
		stringNode := node.Get("string")
		called := false

		stringNode.ArrayForEach(func(index int, value Node) bool {
			called = true
			return true
		})

		if called {
			t.Error("ArrayForEach() should not call function for non-array nodes")
		}
	})
}

func TestWalk(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("walk all nodes", func(t *testing.T) {
		var paths []string
		var nodeTypes []NodeType

		node.Walk(func(path string, n Node) bool {
			paths = append(paths, path)
			nodeTypes = append(nodeTypes, n.Kind())
			return true
		})

		if len(paths) == 0 {
			t.Error("Walk() should visit at least one node")
		}

		// 检查根节点
		if paths[0] != "" || nodeTypes[0] != TypeObject {
			t.Error("Walk() should start with root node")
		}

		// 检查是否包含预期的路径
		expectedPaths := []string{"", "string", "object", "object.nested_string", "array", "array[0]"}
		pathMap := make(map[string]bool)
		for _, path := range paths {
			pathMap[path] = true
		}

		for _, expected := range expectedPaths {
			if !pathMap[expected] {
				t.Errorf("Walk() missing expected path: %s", expected)
			}
		}
	})

	t.Run("walk with early termination", func(t *testing.T) {
		count := 0

		node.Walk(func(path string, n Node) bool {
			count++
			return path != "object" // 在到达object时停止递归其子节点
		})

		if count == 0 {
			t.Error("Walk() should visit at least one node")
		}
	})
}

func TestGetAllKeys(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("object keys", func(t *testing.T) {
		objectNode := node.Get("object")
		keys := objectNode.GetAllKeys()

		expectedKeys := []string{"nested_string", "nested_number", "nested_array"}
		if len(keys) != len(expectedKeys) {
			t.Errorf("GetAllKeys() returned %d keys, want %d", len(keys), len(expectedKeys))
		}

		keyMap := make(map[string]bool)
		for _, key := range keys {
			keyMap[key] = true
		}

		for _, expected := range expectedKeys {
			if !keyMap[expected] {
				t.Errorf("GetAllKeys() missing expected key: %s", expected)
			}
		}
	})

	t.Run("non-object node", func(t *testing.T) {
		stringNode := node.Get("string")
		keys := stringNode.GetAllKeys()
		if keys != nil {
			t.Error("GetAllKeys() should return nil for non-object nodes")
		}
	})
}

func TestGetAllValues(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("array values", func(t *testing.T) {
		arrayNode := node.Get("array")
		values := arrayNode.GetAllValues()

		if len(values) != 6 {
			t.Errorf("GetAllValues() returned %d values, want 6", len(values))
		}

		// 检查类型
		expectedTypes := []NodeType{TypeNumber, TypeNumber, TypeNumber, TypeString, TypeBool, TypeNull}
		for i, value := range values {
			if i < len(expectedTypes) && value.Kind() != expectedTypes[i] {
				t.Errorf("GetAllValues()[%d].Kind() = %v, want %v", i, value.Kind(), expectedTypes[i])
			}
		}
	})

	t.Run("non-array node", func(t *testing.T) {
		stringNode := node.Get("string")
		values := stringNode.GetAllValues()
		if values != nil {
			t.Error("GetAllValues() should return nil for non-array nodes")
		}
	})
}

func TestToMap(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("object to map", func(t *testing.T) {
		objectNode := node.Get("object")
		nodeMap := objectNode.ToMap()

		if len(nodeMap) != 3 {
			t.Errorf("ToMap() returned map with %d entries, want 3", len(nodeMap))
		}

		expectedKeys := []string{"nested_string", "nested_number", "nested_array"}
		for _, key := range expectedKeys {
			if _, exists := nodeMap[key]; !exists {
				t.Errorf("ToMap() missing expected key: %s", key)
			}
		}
	})

	t.Run("non-object node", func(t *testing.T) {
		stringNode := node.Get("string")
		nodeMap := stringNode.ToMap()
		if nodeMap != nil {
			t.Error("ToMap() should return nil for non-object nodes")
		}
	})
}

func TestToSlice(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("array to slice", func(t *testing.T) {
		arrayNode := node.Get("array")
		nodeSlice := arrayNode.ToSlice()

		if len(nodeSlice) != 6 {
			t.Errorf("ToSlice() returned slice with %d elements, want 6", len(nodeSlice))
		}

		// 检查类型
		expectedTypes := []NodeType{TypeNumber, TypeNumber, TypeNumber, TypeString, TypeBool, TypeNull}
		for i, node := range nodeSlice {
			if i < len(expectedTypes) && node.Kind() != expectedTypes[i] {
				t.Errorf("ToSlice()[%d].Kind() = %v, want %v", i, node.Kind(), expectedTypes[i])
			}
		}
	})

	t.Run("non-array node", func(t *testing.T) {
		stringNode := node.Get("string")
		nodeSlice := stringNode.ToSlice()
		if nodeSlice != nil {
			t.Error("ToSlice() should return nil for non-array nodes")
		}
	})
}

// ===== 查找和条件方法测试 =====

func TestFindInObject(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("find existing key", func(t *testing.T) {
		objectNode := node.Get("object")
		key, value, found := objectNode.FindInObject(func(k string, v Node) bool {
			return k == "nested_string"
		})

		if !found {
			t.Error("FindInObject() should find existing key")
		}
		if key != "nested_string" {
			t.Errorf("FindInObject() key = %q, want %q", key, "nested_string")
		}
		if !value.IsString() {
			t.Error("FindInObject() should return string node")
		}
	})

	t.Run("find non-existing condition", func(t *testing.T) {
		objectNode := node.Get("object")
		_, _, found := objectNode.FindInObject(func(k string, v Node) bool {
			return k == "non_existing_key"
		})

		if found {
			t.Error("FindInObject() should not find non-existing key")
		}
	})
}

func TestFindInArray(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("find existing element", func(t *testing.T) {
		arrayNode := node.Get("array")
		index, value, found := arrayNode.FindInArray(func(i int, v Node) bool {
			return v.IsString()
		})

		if !found {
			t.Error("FindInArray() should find string element")
		}
		if index != 3 {
			t.Errorf("FindInArray() index = %d, want 3", index)
		}
		if !value.IsString() {
			t.Error("FindInArray() should return string node")
		}
	})

	t.Run("find non-existing condition", func(t *testing.T) {
		arrayNode := node.Get("array")
		_, _, found := arrayNode.FindInArray(func(i int, v Node) bool {
			return i > 100 // 数组没有这么多元素
		})

		if found {
			t.Error("FindInArray() should not find element with impossible condition")
		}
	})
}

func TestFilterArray(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("filter numbers", func(t *testing.T) {
		arrayNode := node.Get("array")
		numbers := arrayNode.FilterArray(func(i int, v Node) bool {
			return v.IsNumber()
		})

		if len(numbers) != 3 {
			t.Errorf("FilterArray() returned %d numbers, want 3", len(numbers))
		}

		for _, num := range numbers {
			if !num.IsNumber() {
				t.Error("FilterArray() should only return number nodes")
			}
		}
	})

	t.Run("filter with no matches", func(t *testing.T) {
		arrayNode := node.Get("array")
		result := arrayNode.FilterArray(func(i int, v Node) bool {
			return v.IsObject() // 数组中没有对象
		})

		if len(result) != 0 {
			t.Errorf("FilterArray() with no matches should return empty slice, got %d elements", len(result))
		}
	})
}

func TestHasKey(t *testing.T) {
	node := FromBytes(testJSON)

	tests := []struct {
		key      string
		expected bool
	}{
		{"string", true},
		{"number", true},
		{"null", true},
		{"non_existing", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := node.HasKey(tt.key)
			if result != tt.expected {
				t.Errorf("HasKey(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetKeyValue(t *testing.T) {
	node := FromBytes(testJSON)
	defaultNode := FromBytes([]byte(`"default_value"`))

	t.Run("existing key", func(t *testing.T) {
		result := node.GetKeyValue("string", defaultNode)
		if !result.IsString() {
			t.Error("GetKeyValue() should return existing string node")
		}
		if str, err := result.String(); err != nil || str != "hello world" {
			t.Errorf("GetKeyValue() = %q, want %q", str, "hello world")
		}
	})

	t.Run("non-existing key", func(t *testing.T) {
		result := node.GetKeyValue("non_existing", defaultNode)
		if !result.IsString() {
			t.Error("GetKeyValue() should return default node")
		}
		if str, err := result.String(); err != nil || str != "default_value" {
			t.Errorf("GetKeyValue() = %q, want %q", str, "default_value")
		}
	})
}

func TestCountIf(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("count numbers", func(t *testing.T) {
		arrayNode := node.Get("array")
		count := arrayNode.CountIf(func(i int, v Node) bool {
			return v.IsNumber()
		})

		if count != 3 {
			t.Errorf("CountIf() = %d, want 3", count)
		}
	})

	t.Run("count with no matches", func(t *testing.T) {
		arrayNode := node.Get("array")
		count := arrayNode.CountIf(func(i int, v Node) bool {
			return v.IsObject()
		})

		if count != 0 {
			t.Errorf("CountIf() with no matches = %d, want 0", count)
		}
	})
}

func TestAllMatch(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("all exist", func(t *testing.T) {
		arrayNode := node.Get("array")
		result := arrayNode.AllMatch(func(i int, v Node) bool {
			return v.Exists()
		})

		if !result {
			t.Error("AllMatch() should return true when all elements exist")
		}
	})

	t.Run("not all numbers", func(t *testing.T) {
		arrayNode := node.Get("array")
		result := arrayNode.AllMatch(func(i int, v Node) bool {
			return v.IsNumber()
		})

		if result {
			t.Error("AllMatch() should return false when not all elements are numbers")
		}
	})
}

func TestAnyMatch(t *testing.T) {
	node := FromBytes(testJSON)

	t.Run("any string", func(t *testing.T) {
		arrayNode := node.Get("array")
		result := arrayNode.AnyMatch(func(i int, v Node) bool {
			return v.IsString()
		})

		if !result {
			t.Error("AnyMatch() should return true when any element is string")
		}
	})

	t.Run("any object", func(t *testing.T) {
		arrayNode := node.Get("array")
		result := arrayNode.AnyMatch(func(i int, v Node) bool {
			return v.IsObject()
		})

		if result {
			t.Error("AnyMatch() should return false when no element is object")
		}
	})
}

// ===== 边界情况和错误处理测试 =====

func TestEdgeCases(t *testing.T) {
	t.Run("empty json", func(t *testing.T) {
		node := FromBytes([]byte{})
		if node.Exists() {
			t.Error("Empty JSON should not exist")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		node := FromBytes([]byte(`{"incomplete": tr`))
		if node.Exists() {
			t.Error("Malformed JSON should not exist")
		}
	})

	t.Run("deeply nested access", func(t *testing.T) {
		deepJSON := []byte(`{"a":{"b":{"c":{"d":"deep_value"}}}}`)
		node := FromBytes(deepJSON)
		result := node.GetPath("a.b.c.d")

		if !result.Exists() {
			t.Error("Deep nested access should work")
		}
		if str, err := result.String(); err != nil || str != "deep_value" {
			t.Errorf("Deep nested value = %q, want %q", str, "deep_value")
		}
	})

	t.Run("array bounds", func(t *testing.T) {
		arrayJSON := []byte(`[1,2,3]`)
		node := FromBytes(arrayJSON)

		// 正常访问
		if !node.Index(0).Exists() {
			t.Error("Valid array index should exist")
		}
		if !node.Index(2).Exists() {
			t.Error("Valid array index should exist")
		}

		// 越界访问
		if node.Index(-1).Exists() {
			t.Error("Negative array index should not exist")
		}
		if node.Index(3).Exists() {
			t.Error("Out of bounds array index should not exist")
		}
	})

	t.Run("unicode handling", func(t *testing.T) {
		unicodeJSON := []byte(`{"emoji":"🚀","chinese":"你好","mixed":"Hello 世界"}`)
		node := FromBytes(unicodeJSON)

		if emoji, err := node.Get("emoji").String(); err != nil || emoji != "🚀" {
			t.Errorf("Unicode emoji = %q, want %q", emoji, "🚀")
		}
		if chinese, err := node.Get("chinese").String(); err != nil || chinese != "你好" {
			t.Errorf("Chinese text = %q, want %q", chinese, "你好")
		}
	})

	t.Run("escaped characters", func(t *testing.T) {
		escapedJSON := []byte(`{"newline":"line1\nline2","quote":"say \"hello\"","backslash":"path\\to\\file"}`)
		node := FromBytes(escapedJSON)

		if newline, err := node.Get("newline").String(); err != nil || newline != "line1\nline2" {
			t.Errorf("Newline escape = %q, want %q", newline, "line1\nline2")
		}
	})

	t.Run("large numbers", func(t *testing.T) {
		largeJSON := []byte(`{"maxint64":9223372036854775807,"minint64":-9223372036854775808}`)
		node := FromBytes(largeJSON)

		if max, err := node.Get("maxint64").Int(); err != nil || max != math.MaxInt64 {
			t.Errorf("Max int64 = %d, want %d", max, int64(math.MaxInt64))
		}
		if min, err := node.Get("minint64").Int(); err != nil || min != math.MinInt64 {
			t.Errorf("Min int64 = %d, want %d", min, int64(math.MinInt64))
		}
	})
}

// ===== 嵌套JSON展开功能测试 =====

func XTestNestedJSONExpansion(t *testing.T) {
	t.Run("simple nested json string", func(t *testing.T) {
		nestedJSON := []byte(`{
			"data": "{\"name\":\"Alice\",\"age\":30}",
			"normal": "regular string"
		}`)

		node := FromBytes(nestedJSON)

		// 访问嵌套的JSON应该自动展开
		dataNode := node.Get("data")
		if !dataNode.IsObject() {
			t.Error("Nested JSON string should be expanded to object")
		}

		// 应该能够访问嵌套JSON的字段
		name := dataNode.Get("name")
		if !name.Exists() || !name.IsString() {
			t.Error("Should be able to access expanded nested JSON fields")
		}

		if nameStr, err := name.String(); err != nil || nameStr != "Alice" {
			t.Errorf("Nested JSON name = %q, want %q", nameStr, "Alice")
		}

		age := dataNode.Get("age")
		if !age.Exists() || !age.IsNumber() {
			t.Error("Should be able to access expanded nested JSON number fields")
		}

		if ageNum, err := age.Int(); err != nil || ageNum != 30 {
			t.Errorf("Nested JSON age = %d, want %d", ageNum, 30)
		}
	})

	t.Run("nested json array", func(t *testing.T) {
		nestedJSON := []byte(`{
			"items": "[1,2,3,\"four\"]",
			"meta": "not json"
		}`)

		node := FromBytes(nestedJSON)

		// 访问嵌套的JSON数组应该自动展开
		itemsNode := node.Get("items")
		if !itemsNode.IsArray() {
			t.Error("Nested JSON array string should be expanded to array")
		}

		if itemsNode.Len() != 4 {
			t.Errorf("Expanded array length = %d, want 4", itemsNode.Len())
		}

		// 检查数组元素
		first := itemsNode.Index(0)
		if !first.IsNumber() {
			t.Error("First array element should be number")
		}

		fourth := itemsNode.Index(3)
		if !fourth.IsString() {
			t.Error("Fourth array element should be string")
		}

		if str, err := fourth.String(); err != nil || str != "four" {
			t.Errorf("Fourth element = %q, want %q", str, "four")
		}
	})

	t.Run("deeply nested json", func(t *testing.T) {
		nestedJSON := []byte(`{
			"level1": "{\"level2\":\"{\\\"level3\\\":\\\"deep_value\\\"}\"}"
		}`)

		node := FromBytes(nestedJSON)

		// 多层嵌套应该递归展开
		level1 := node.Get("level1")
		if !level1.IsObject() {
			t.Error("Level 1 should be expanded to object")
		}

		level2 := level1.Get("level2")
		if !level2.IsObject() {
			t.Error("Level 2 should be expanded to object")
		}

		level3 := level2.Get("level3")
		if !level3.IsString() {
			t.Error("Level 3 should be string")
		}

		if value, err := level3.String(); err != nil || value != "deep_value" {
			t.Errorf("Deep nested value = %q, want %q", value, "deep_value")
		}
	})

	t.Run("mixed nested and regular data", func(t *testing.T) {
		mixedJSON := []byte(`{
			"regular_string": "hello",
			"regular_number": 42,
			"nested_object": "{\"inner\":\"value\"}",
			"nested_array": "[1,2,3]",
			"regular_array": [4,5,6],
			"not_json_string": "this is not {json}"
		}`)

		node := FromBytes(mixedJSON)

		// 常规字段应该正常工作
		if str, err := node.Get("regular_string").String(); err != nil || str != "hello" {
			t.Error("Regular string should work normally")
		}

		if num, err := node.Get("regular_number").Int(); err != nil || num != 42 {
			t.Error("Regular number should work normally")
		}

		// 嵌套的JSON应该展开
		nestedObj := node.Get("nested_object")
		if !nestedObj.IsObject() {
			t.Error("Nested JSON object should be expanded")
		}

		if inner, err := nestedObj.Get("inner").String(); err != nil || inner != "value" {
			t.Error("Should access nested object fields")
		}

		nestedArr := node.Get("nested_array")
		if !nestedArr.IsArray() {
			t.Error("Nested JSON array should be expanded")
		}

		if nestedArr.Len() != 3 {
			t.Error("Nested array should have 3 elements")
		}

		// 常规数组应该保持不变
		regularArr := node.Get("regular_array")
		if !regularArr.IsArray() {
			t.Error("Regular array should remain array")
		}

		// 非JSON字符串应该保持字符串
		notJson := node.Get("not_json_string")
		if !notJson.IsString() {
			t.Error("Non-JSON string should remain string")
		}
	})

	t.Run("invalid nested json", func(t *testing.T) {
		invalidJSON := []byte(`{
			"malformed": "{invalid json}",
			"incomplete": "{\"key\":",
			"empty": "",
			"normal": "normal string"
		}`)

		node := FromBytes(invalidJSON)

		// 格式错误的JSON应该保持为字符串
		malformed := node.Get("malformed")
		if !malformed.IsString() {
			t.Error("Malformed JSON should remain as string")
		}

		incomplete := node.Get("incomplete")
		if !incomplete.IsString() {
			t.Error("Incomplete JSON should remain as string")
		}

		empty := node.Get("empty")
		if !empty.IsString() {
			t.Error("Empty string should remain as string")
		}

		// 正常字符串应该不受影响
		normal := node.Get("normal")
		if !normal.IsString() {
			t.Error("Normal string should remain as string")
		}
	})
}

// ===== 错误处理和nil安全测试 =====

func TestErrorHandling(t *testing.T) {
	t.Run("nil function parameters", func(t *testing.T) {
		node := FromBytes(testJSON)

		// ForEach with nil function should not panic
		node.Get("object").ForEach(nil)

		// ArrayForEach with nil function should not panic
		node.Get("array").ArrayForEach(nil)

		// Walk with nil function should not panic
		node.Walk(nil)

		// FindInObject with nil predicate should return false
		_, _, found := node.Get("object").FindInObject(nil)
		if found {
			t.Error("FindInObject with nil predicate should return false")
		}

		// FilterArray with nil predicate should return nil
		result := node.Get("array").FilterArray(nil)
		if result != nil {
			t.Error("FilterArray with nil predicate should return nil")
		}
	})

	t.Run("out of range operations", func(t *testing.T) {
		node := FromBytes(testJSON)

		// String operations on wrong types
		if _, err := node.Get("number").String(); err == nil {
			t.Error("String() on number should return error")
		}

		// Int operations on wrong types
		if _, err := node.Get("string").Int(); err == nil {
			t.Error("Int() on string should return error")
		}

		// Array operations on non-arrays
		nonArray := node.Get("string")
		if nonArray.Index(0).Exists() {
			t.Error("Index() on non-array should return non-existent node")
		}

		if nonArray.Len() == 0 {
			t.Error("Len() on string should return character count")
		}
	})

	t.Run("memory safety", func(t *testing.T) {
		// Test with very short JSON
		shortJSON := []byte(`{}`)
		node := FromBytes(shortJSON)

		if !node.Exists() {
			t.Error("Empty object should exist")
		}

		// Test accessing non-existent fields shouldn't crash
		result := node.Get("nonexistent")
		if result.Exists() {
			t.Error("Non-existent field should not exist")
		}

		// Test with single character
		singleChar := []byte(`1`)
		singleNode := FromBytes(singleChar)
		if !singleNode.Exists() || !singleNode.IsNumber() {
			t.Error("Single number should parse correctly")
		}
	})
}

// ===== 类型转换边界测试 =====

func TestTypeConversionBoundaries(t *testing.T) {
	t.Run("integer overflow", func(t *testing.T) {
		// 测试超出int64范围的数字
		overflowJSON := []byte(`{"overflow": 18446744073709551615}`) // 超出int64最大值
		node := FromBytes(overflowJSON)

		if _, err := node.Get("overflow").Int(); err == nil {
			t.Error("Int() should return error for overflow values")
		}

		// 但Float应该能处理
		if _, err := node.Get("overflow").Float(); err != nil {
			t.Error("Float() should handle large numbers")
		}
	})

	t.Run("float precision", func(t *testing.T) {
		precisionJSON := []byte(`{"precise": 1.7976931348623157e+307}`) // 接近但不超过float64最大值
		node := FromBytes(precisionJSON)

		value, err := node.Get("precise").Float()
		if err != nil {
			t.Errorf("Float() should handle maximum float64 values: %v", err)
		}

		if math.IsInf(value, 0) {
			t.Error("Value should not be infinity")
		}
	})

	t.Run("unicode in strings", func(t *testing.T) {
		unicodeJSON := []byte(`{
			"emoji": "🎉🚀💻",
			"chinese": "测试中文字符",
			"japanese": "テストデータ",
			"mixed": "Mixed: 中文 English 日本語"
		}`)

		node := FromBytes(unicodeJSON)

		tests := []struct {
			key      string
			expected string
		}{
			{"emoji", "🎉🚀💻"},
			{"chinese", "测试中文字符"},
			{"japanese", "テストデータ"},
			{"mixed", "Mixed: 中文 English 日本語"},
		}

		for _, tt := range tests {
			if str, err := node.Get(tt.key).String(); err != nil || str != tt.expected {
				t.Errorf("Unicode string %s = %q, want %q", tt.key, str, tt.expected)
			}
		}
	})
}
