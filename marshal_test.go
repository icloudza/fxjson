package fxjson

import (
	"fmt"
	"testing"
	"time"
)

// 测试结构体
type Person struct {
	Name     string                 `json:"name"`
	Age      int                    `json:"age"`
	Email    string                 `json:"email,omitempty"`
	IsActive bool                   `json:"is_active"`
	Height   float64                `json:"height"`
	Address  *Address               `json:"address,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
	Created  time.Time              `json:"created"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
	ZipCode string `json:"zip_code,omitempty"`
}

type Company struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Employees   []Person  `json:"employees"`
	Founded     time.Time `json:"founded"`
	IsPublic    bool      `json:"is_public"`
	Revenue     float64   `json:"revenue,omitempty"`
	Departments []string  `json:"departments,omitempty"`
}

// TestBasicMarshal 测试基本序列化
func TestBasicMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: "null",
		},
		{
			name:     "bool_true",
			input:    true,
			expected: "true",
		},
		{
			name:     "bool_false",
			input:    false,
			expected: "false",
		},
		{
			name:     "int",
			input:    42,
			expected: "42",
		},
		{
			name:     "float",
			input:    3.14159,
			expected: "3.14159",
		},
		{
			name:     "string",
			input:    "hello world",
			expected: `"hello world"`,
		},
		{
			name:     "string_with_escape",
			input:    "hello\nworld",
			expected: `"hello\nworld"`,
		},
		{
			name:     "empty_slice",
			input:    []int{},
			expected: "[]",
		},
		{
			name:     "int_slice",
			input:    []int{1, 2, 3},
			expected: "[1,2,3]",
		},
		{
			name:     "string_slice",
			input:    []string{"a", "b", "c"},
			expected: `["a","b","c"]`,
		},
		{
			name:     "empty_map",
			input:    map[string]int{},
			expected: "{}",
		},
		{
			name:     "simple_map",
			input:    map[string]int{"x": 1, "y": 2},
			expected: `{"x":1,"y":2}`, // 注意：map的顺序可能不同
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			got := string(result)
			if tt.name == "simple_map" {
				// 对于map，检查是否包含预期的键值对
				if !contains(got, `"x":1`) || !contains(got, `"y":2`) {
					t.Errorf("Marshal() = %v, want map containing x:1 and y:2", got)
				}
			} else {
				if got != tt.expected {
					t.Errorf("Marshal() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

// TestStructMarshal 测试结构体序列化
func TestStructMarshal(t *testing.T) {
	person := Person{
		Name:     "John Doe",
		Age:      30,
		Email:    "john@example.com",
		IsActive: true,
		Height:   5.9,
		Address: &Address{
			Street:  "123 Main St",
			City:    "New York",
			Country: "USA",
			ZipCode: "10001",
		},
		Tags: []string{"developer", "golang"},
		Meta: map[string]interface{}{
			"level":      "senior",
			"experience": 5,
		},
		Created: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	result, err := MarshalStruct(person)
	if err != nil {
		t.Fatalf("MarshalStruct failed: %v", err)
	}

	// 验证结果是有效的JSON
	if !ValidateJSON(result) {
		t.Errorf("Generated JSON is invalid: %s", string(result))
	}

	// 解析回来验证数据正确性
	node := FromBytes(result)
	if !node.Exists() {
		t.Errorf("Failed to parse generated JSON")
	}

	// 检查一些关键字段
	if name, _ := node.Get("name").String(); name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", name)
	}

	if age, _ := node.Get("age").Int(); age != 30 {
		t.Errorf("Expected age 30, got %d", age)
	}

	if active, _ := node.Get("is_active").Bool(); !active {
		t.Errorf("Expected is_active true, got %v", active)
	}
}

// TestFastMarshal 测试快速序列化
func TestFastMarshal(t *testing.T) {
	data := map[string]interface{}{
		"name":    "test",
		"age":     25,
		"active":  true,
		"scores":  []int{90, 85, 92},
		"details": map[string]string{"city": "Beijing"},
	}

	result := FastMarshal(data)
	if len(result) == 0 {
		t.Errorf("FastMarshal returned empty result")
	}

	// 验证结果是有效的JSON
	if !ValidateJSON(result) {
		t.Errorf("FastMarshal generated invalid JSON: %s", string(result))
	}
}

// TestMarshalWithOptions 测试带选项的序列化
func TestMarshalWithOptions(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"age":  25,
		"tags": []string{"a", "b"},
	}

	// 测试压缩模式
	compact, err := MarshalWithOptions(data, DefaultSerializeOptions)
	if err != nil {
		t.Fatalf("MarshalWithOptions failed: %v", err)
	}

	// 测试美化模式
	pretty, err := MarshalWithOptions(data, PrettySerializeOptions)
	if err != nil {
		t.Fatalf("MarshalWithOptions failed: %v", err)
	}

	// 美化模式应该更长（包含换行和缩进）
	if len(pretty) <= len(compact) {
		t.Errorf("Pretty format should be longer than compact format")
	}

	// 两种模式都应该是有效的JSON
	if !ValidateJSON(compact) {
		t.Errorf("Compact JSON is invalid: %s", string(compact))
	}

	if !ValidateJSON(pretty) {
		t.Errorf("Pretty JSON is invalid: %s", string(pretty))
	}
}

// TestNodeToJSON 测试Node到JSON的序列化
func TestNodeToJSON(t *testing.T) {
	jsonStr := `{"name":"John","age":30,"tags":["a","b"],"address":{"city":"NYC"}}`
	node := FromBytes([]byte(jsonStr))

	if !node.Exists() {
		t.Fatalf("Failed to parse JSON")
	}

	// 测试整个节点序列化
	result, err := node.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 验证结果是有效的JSON
	if !ValidateJSON([]byte(result)) {
		t.Errorf("Node ToJSON generated invalid JSON: %s", result)
	}

	// 测试子节点序列化
	addressNode := node.Get("address")
	addressJSON, err := addressNode.ToJSON()
	if err != nil {
		t.Fatalf("Address ToJSON failed: %v", err)
	}

	if !contains(addressJSON, "NYC") {
		t.Errorf("Address JSON doesn't contain expected data: %s", addressJSON)
	}
}

// TestBatchMarshal 测试批量序列化
func TestBatchMarshal(t *testing.T) {
	persons := []interface{}{
		Person{Name: "Alice", Age: 25},
		Person{Name: "Bob", Age: 30},
		Person{Name: "Charlie", Age: 35},
	}

	results, err := BatchMarshalStructs(persons)
	if err != nil {
		t.Fatalf("BatchMarshalStructs failed: %v", err)
	}

	if len(results) != len(persons) {
		t.Errorf("Expected %d results, got %d", len(persons), len(results))
	}

	// 验证每个结果都是有效的JSON
	for i, result := range results {
		if !ValidateJSON(result) {
			t.Errorf("Result %d is invalid JSON: %s", i, string(result))
		}
	}
}

// TestStreamMarshal 测试流式序列化
func TestStreamMarshal(t *testing.T) {
	var output []byte
	writer := func(data []byte) error {
		output = append(output, data...)
		return nil
	}

	marshaler := NewStreamMarshaler(writer, DefaultSerializeOptions)
	defer marshaler.Close()

	// 序列化一个数组
	if err := marshaler.StartArray(); err != nil {
		t.Fatalf("StartArray failed: %v", err)
	}

	if err := marshaler.WriteValue(1); err != nil {
		t.Fatalf("WriteValue failed: %v", err)
	}

	if err := marshaler.WriteValue("test"); err != nil {
		t.Fatalf("WriteValue failed: %v", err)
	}

	if err := marshaler.WriteValue(true); err != nil {
		t.Fatalf("WriteValue failed: %v", err)
	}

	if err := marshaler.EndArray(); err != nil {
		t.Fatalf("EndArray failed: %v", err)
	}

	expected := `[1,"test",true]`
	if string(output) != expected {
		t.Errorf("StreamMarshal result = %s, want %s", string(output), expected)
	}
}

// TestPerformance 性能测试
func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// 创建测试数据
	company := Company{
		ID:          1,
		Name:        "Tech Corp",
		Founded:     time.Now(),
		IsPublic:    true,
		Revenue:     1000000.50,
		Departments: []string{"Engineering", "Sales", "Marketing"},
		Employees:   make([]Person, 1000),
	}

	// 填充员工数据
	for i := 0; i < 1000; i++ {
		company.Employees[i] = Person{
			Name:     fmt.Sprintf("Employee %d", i),
			Age:      25 + i%40,
			Email:    fmt.Sprintf("emp%d@company.com", i),
			IsActive: i%2 == 0,
			Height:   5.0 + float64(i%20)/10,
			Created:  time.Now().AddDate(0, 0, -i),
		}
	}

	// 性能测试
	start := time.Now()
	result, err := Marshal(company)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Performance test failed: %v", err)
	}

	t.Logf("Marshaled %d employees in %v", len(company.Employees), duration)
	t.Logf("Result size: %d bytes", len(result))

	// 验证结果
	if !ValidateJSON(result) {
		t.Errorf("Performance test generated invalid JSON")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// BenchmarkMarshal 基准测试
func BenchmarkMarshal(b *testing.B) {
	person := Person{
		Name:     "John Doe",
		Age:      30,
		Email:    "john@example.com",
		IsActive: true,
		Height:   5.9,
		Tags:     []string{"developer", "golang", "performance"},
		Meta: map[string]interface{}{
			"level":      "senior",
			"experience": 5,
			"skills":     []string{"go", "rust", "python"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(person)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFastMarshal 快速序列化基准测试
func BenchmarkFastMarshal(b *testing.B) {
	person := Person{
		Name:     "John Doe",
		Age:      30,
		IsActive: true,
		Height:   5.9,
		Tags:     []string{"developer", "golang"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FastMarshal(person)
	}
}

// BenchmarkStructMarshal 结构体序列化基准测试
func BenchmarkStructMarshal(b *testing.B) {
	person := Person{
		Name:     "John Doe",
		Age:      30,
		Email:    "john@example.com",
		IsActive: true,
		Height:   5.9,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := MarshalStruct(person)
		if err != nil {
			b.Fatal(err)
		}
	}
}
