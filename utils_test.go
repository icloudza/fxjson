package fxjson

import (
	"testing"
	"time"
)

// TestDefaultValueFunctions 测试默认值函数
func TestDefaultValueFunctions(t *testing.T) {
	jsonData := []byte(`{
		"name": "test",
		"age": 25,
		"score": 3.14,
		"active": true,
		"invalid": "not_a_number"
	}`)

	node := FromBytes(jsonData)

	// 测试StringOr
	if v := node.Get("name").StringOr("default"); v != "test" {
		t.Errorf("StringOr failed: expected 'test', got '%s'", v)
	}
	if v := node.Get("missing").StringOr("default"); v != "default" {
		t.Errorf("StringOr failed: expected 'default', got '%s'", v)
	}

	// 测试IntOr
	if v := node.Get("age").IntOr(0); v != 25 {
		t.Errorf("IntOr failed: expected 25, got %d", v)
	}
	if v := node.Get("invalid").IntOr(99); v != 99 {
		t.Errorf("IntOr failed: expected 99, got %d", v)
	}

	// 测试FloatOr
	if v := node.Get("score").FloatOr(0.0); v != 3.14 {
		t.Errorf("FloatOr failed: expected 3.14, got %f", v)
	}
	if v := node.Get("missing").FloatOr(2.71); v != 2.71 {
		t.Errorf("FloatOr failed: expected 2.71, got %f", v)
	}

	// 测试BoolOr
	if v := node.Get("active").BoolOr(false); v != true {
		t.Errorf("BoolOr failed: expected true, got %v", v)
	}
	if v := node.Get("missing").BoolOr(true); v != true {
		t.Errorf("BoolOr failed: expected true, got %v", v)
	}
}

// TestMultiplePathFunctions 测试批量路径函数
func TestMultiplePathFunctions(t *testing.T) {
	jsonData := []byte(`{
		"user": {
			"name": "Alice",
			"profile": {
				"age": 30,
				"city": "Beijing"
			}
		},
		"status": "active"
	}`)

	node := FromBytes(jsonData)

	// 测试GetMultiple
	nodes := node.GetMultiple("user.name", "user.profile.age", "status", "missing")
	if len(nodes) != 4 {
		t.Errorf("GetMultiple failed: expected 4 nodes, got %d", len(nodes))
	}
	if name, _ := nodes[0].String(); name != "Alice" {
		t.Errorf("GetMultiple failed: expected 'Alice', got '%s'", name)
	}
	if age, _ := nodes[1].Int(); age != 30 {
		t.Errorf("GetMultiple failed: expected 30, got %d", age)
	}
	if !nodes[3].IsNull() && !nodes[3].Exists() {
		// missing路径应该不存在
	}

	// 测试HasAnyPath
	if !node.HasAnyPath("user.name", "missing1", "missing2") {
		t.Error("HasAnyPath failed: expected true")
	}
	if node.HasAnyPath("missing1", "missing2", "missing3") {
		t.Error("HasAnyPath failed: expected false")
	}

	// 测试HasAllPaths
	if !node.HasAllPaths("user.name", "user.profile.age", "status") {
		t.Error("HasAllPaths failed: expected true")
	}
	if node.HasAllPaths("user.name", "missing") {
		t.Error("HasAllPaths failed: expected false")
	}
}

// TestSliceConversions 测试数组转换函数
func TestSliceConversions(t *testing.T) {
	// 测试ToStringSlice
	strArrayData := []byte(`["hello", "world", "test"]`)
	strNode := FromBytes(strArrayData)
	if strSlice, err := strNode.ToStringSlice(); err != nil {
		t.Errorf("ToStringSlice failed: %v", err)
	} else if len(strSlice) != 3 || strSlice[0] != "hello" {
		t.Errorf("ToStringSlice failed: unexpected result %v", strSlice)
	}

	// 测试ToIntSlice
	intArrayData := []byte(`[1, 2, 3, 4, 5]`)
	intNode := FromBytes(intArrayData)
	if intSlice, err := intNode.ToIntSlice(); err != nil {
		t.Errorf("ToIntSlice failed: %v", err)
	} else if len(intSlice) != 5 || intSlice[0] != 1 {
		t.Errorf("ToIntSlice failed: unexpected result %v", intSlice)
	}

	// 测试ToFloatSlice
	floatArrayData := []byte(`[1.1, 2.2, 3.3]`)
	floatNode := FromBytes(floatArrayData)
	if floatSlice, err := floatNode.ToFloatSlice(); err != nil {
		t.Errorf("ToFloatSlice failed: %v", err)
	} else if len(floatSlice) != 3 || floatSlice[0] != 1.1 {
		t.Errorf("ToFloatSlice failed: unexpected result %v", floatSlice)
	}

	// 测试ToBoolSlice
	boolArrayData := []byte(`[true, false, true]`)
	boolNode := FromBytes(boolArrayData)
	if boolSlice, err := boolNode.ToBoolSlice(); err != nil {
		t.Errorf("ToBoolSlice failed: %v", err)
	} else if len(boolSlice) != 3 || boolSlice[0] != true {
		t.Errorf("ToBoolSlice failed: unexpected result %v", boolSlice)
	}
}

// TestValidationFunctions 测试验证函数
func TestValidationFunctions(t *testing.T) {
	jsonData := []byte(`{"email":"test@example.com","invalid_email":"not_an_email","url":"https://example.com","invalid_url":"not a url","phone":"+1234567890","uuid":"550e8400-e29b-41d4-a716-446655440000","ipv4":"192.168.1.1","ipv6":"2001:0db8:85a3:0000:0000:8a2e:0370:7334"}`)

	node := FromBytes(jsonData)

	// 测试电子邮件验证
	if !node.Get("email").IsValidEmail() {
		t.Error("IsValidEmail failed for valid email")
	}
	if node.Get("invalid_email").IsValidEmail() {
		t.Error("IsValidEmail failed for invalid email")
	}

	// 测试URL验证
	if !node.Get("url").IsValidURL() {
		t.Error("IsValidURL failed for valid URL")
	}
	if node.Get("invalid_url").IsValidURL() {
		t.Error("IsValidURL failed for invalid URL")
	}

	// 测试电话号码验证
	if !node.Get("phone").IsValidPhone() {
		t.Error("IsValidPhone failed for valid phone")
	}

	// 测试UUID验证
	if !node.Get("uuid").IsValidUUID() {
		t.Error("IsValidUUID failed for valid UUID")
	}

	// 测试IP地址验证
	if !node.Get("ipv4").IsValidIPv4() {
		t.Error("IsValidIPv4 failed for valid IPv4")
	}
	if !node.Get("ipv6").IsValidIPv6() {
		t.Error("IsValidIPv6 failed for valid IPv6")
	}
	if !node.Get("ipv4").IsValidIP() {
		t.Error("IsValidIP failed for valid IPv4")
	}
}

// TestStringOperations 测试字符串操作函数
func TestStringOperations(t *testing.T) {
	jsonData := []byte(`{
		"text": "Hello World",
		"lower": "lowercase",
		"upper": "UPPERCASE",
		"spaces": "  trimmed  "
	}`)

	node := FromBytes(jsonData)

	// 测试Contains
	if !node.Get("text").Contains("World") {
		t.Error("Contains failed")
	}
	if node.Get("text").Contains("NotFound") {
		t.Error("Contains failed for non-existent substring")
	}

	// 测试StartsWith和EndsWith
	if !node.Get("text").StartsWith("Hello") {
		t.Error("StartsWith failed")
	}
	if !node.Get("text").EndsWith("World") {
		t.Error("EndsWith failed")
	}

	// 测试大小写转换
	if lower, err := node.Get("upper").ToLower(); err != nil || lower != "uppercase" {
		t.Errorf("ToLower failed: %v, %s", err, lower)
	}
	if upper, err := node.Get("lower").ToUpper(); err != nil || upper != "LOWERCASE" {
		t.Errorf("ToUpper failed: %v, %s", err, upper)
	}

	// 测试Trim
	if trimmed, err := node.Get("spaces").Trim(); err != nil || trimmed != "trimmed" {
		t.Errorf("Trim failed: %v, '%s'", err, trimmed)
	}
}

// TestArrayOperations 测试数组操作函数
func TestArrayOperations(t *testing.T) {
	jsonData := []byte(`[1, 2, 3, 4, 5]`)
	node := FromBytes(jsonData)

	// 测试First和Last
	if first := node.First(); !first.Exists() {
		t.Error("First failed")
	} else if val, _ := first.Int(); val != 1 {
		t.Errorf("First failed: expected 1, got %d", val)
	}

	if last := node.Last(); !last.Exists() {
		t.Error("Last failed")
	} else if val, _ := last.Int(); val != 5 {
		t.Errorf("Last failed: expected 5, got %d", val)
	}

	// 测试Slice
	sliced := node.Slice(1, 4)
	if len(sliced) != 3 {
		t.Errorf("Slice failed: expected 3 elements, got %d", len(sliced))
	}
	if val, _ := sliced[0].Int(); val != 2 {
		t.Errorf("Slice failed: expected first element to be 2, got %d", val)
	}

	// 测试Reverse
	reversed := node.Reverse()
	if len(reversed) != 5 {
		t.Errorf("Reverse failed: expected 5 elements, got %d", len(reversed))
	}
	if val, _ := reversed[0].Int(); val != 5 {
		t.Errorf("Reverse failed: expected first element to be 5, got %d", val)
	}
}

// TestObjectOperations 测试对象操作函数
func TestObjectOperations(t *testing.T) {
	jsonData1 := []byte(`{"a": 1, "b": 2, "c": 3}`)
	jsonData2 := []byte(`{"b": 20, "d": 4}`)
	node1 := FromBytes(jsonData1)
	node2 := FromBytes(jsonData2)

	// 测试Merge
	merged := node1.Merge(node2)
	if len(merged) != 4 {
		t.Errorf("Merge failed: expected 4 keys, got %d", len(merged))
	}
	if val, _ := merged["b"].Int(); val != 20 {
		t.Errorf("Merge failed: expected b=20, got %d", val)
	}

	// 测试Pick
	picked := node1.Pick("a", "c")
	if len(picked) != 2 {
		t.Errorf("Pick failed: expected 2 keys, got %d", len(picked))
	}
	if _, exists := picked["b"]; exists {
		t.Error("Pick failed: 'b' should not be in picked result")
	}

	// 测试Omit
	omitted := node1.Omit("b")
	if len(omitted) != 2 {
		t.Errorf("Omit failed: expected 2 keys, got %d", len(omitted))
	}
	if _, exists := omitted["b"]; exists {
		t.Error("Omit failed: 'b' should not be in omitted result")
	}
}

// TestComparisonFunctions 测试比较函数
func TestComparisonFunctions(t *testing.T) {
	jsonData := []byte(`{
		"empty_string": "",
		"empty_array": [],
		"empty_object": {},
		"null_value": null,
		"positive": 5,
		"negative": -3,
		"zero": 0,
		"integer": 10,
		"float": 10.5
	}`)

	node := FromBytes(jsonData)

	// 测试IsEmpty
	if !node.Get("empty_string").IsEmpty() {
		t.Error("IsEmpty failed for empty string")
	}
	if !node.Get("empty_array").IsEmpty() {
		t.Error("IsEmpty failed for empty array")
	}
	if !node.Get("empty_object").IsEmpty() {
		t.Error("IsEmpty failed for empty object")
	}
	if !node.Get("null_value").IsEmpty() {
		t.Error("IsEmpty failed for null value")
	}

	// 测试数字比较函数
	if !node.Get("positive").IsPositive() {
		t.Error("IsPositive failed")
	}
	if !node.Get("negative").IsNegative() {
		t.Error("IsNegative failed")
	}
	if !node.Get("zero").IsZero() {
		t.Error("IsZero failed")
	}
	if !node.Get("integer").IsInteger() {
		t.Error("IsInteger failed for integer value")
	}
	if node.Get("float").IsInteger() {
		t.Error("IsInteger failed for float value")
	}

	// 测试InRange
	if !node.Get("positive").InRange(0, 10) {
		t.Error("InRange failed")
	}
	if node.Get("negative").InRange(0, 10) {
		t.Error("InRange failed for out of range value")
	}
}

// TestEquals 测试Equals函数
func TestEquals(t *testing.T) {
	jsonData1 := []byte(`{"name": "test", "value": 123}`)
	jsonData2 := []byte(`{"name": "test", "value": 123}`)
	jsonData3 := []byte(`{"name": "test", "value": 456}`)

	node1 := FromBytes(jsonData1)
	node2 := FromBytes(jsonData2)
	node3 := FromBytes(jsonData3)

	// 相同值的字段应该相等
	if !node1.Get("name").Equals(node2.Get("name")) {
		t.Error("Equals failed for equal values")
	}

	// 不同值的字段不应该相等
	if node1.Get("value").Equals(node3.Get("value")) {
		t.Error("Equals failed for different values")
	}
}

// TestCacheDisabling 测试缓存禁用功能
func TestCacheDisabling(t *testing.T) {
	// 先启用缓存
	cache := NewMemoryCache(10)
	EnableCaching(cache)

	// 使用缓存解析
	jsonData := []byte(`{"test": "data"}`)
	node1 := FromBytesWithCache(jsonData, 5*time.Minute)
	if !node1.Exists() {
		t.Error("Failed to parse with cache enabled")
	}

	// 禁用缓存
	DisableCaching()

	// 再次解析应该不使用缓存
	node2 := FromBytesWithCache(jsonData, 5*time.Minute)
	if !node2.Exists() {
		t.Error("Failed to parse with cache disabled")
	}

	// 缓存应该已被禁用，统计信息应该不变
	// 重新启用缓存以便后续测试
	EnableCaching(NewMemoryCache(100))
}

// BenchmarkCacheKeyGeneration 基准测试缓存键生成性能
func BenchmarkCacheKeyGeneration(b *testing.B) {
	data := []byte(`{"name": "test", "value": 123, "nested": {"key": "value"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateCacheKey(data)
	}
}

// BenchmarkDefaultValueFunctions 基准测试默认值函数性能
func BenchmarkDefaultValueFunctions(b *testing.B) {
	jsonData := []byte(`{"name": "test", "age": 25, "score": 3.14}`)
	node := FromBytes(jsonData)

	b.Run("StringOr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = node.Get("name").StringOr("default")
		}
	})

	b.Run("IntOr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = node.Get("age").IntOr(0)
		}
	})

	b.Run("FloatOr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = node.Get("score").FloatOr(0.0)
		}
	})
}
