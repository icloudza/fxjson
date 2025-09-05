package fxjson

import (
	"testing"
	"time"
)

// TestFromStringComprehensive 全面测试FromString方法
func TestFromStringComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		testFunc func(*testing.T, Node)
	}{
		{
			name:    "基本对象解析",
			jsonStr: `{"name": "张三", "age": 30, "active": true}`,
			testFunc: func(t *testing.T, node Node) {
				if !node.Exists() {
					t.Error("节点不存在")
					return
				}

				name := node.Get("name").StringOr("")
				if name != "张三" {
					t.Errorf("期望名字为'张三'，实际为'%s'", name)
				}

				age := node.Get("age").IntOr(0)
				if age != 30 {
					t.Errorf("期望年龄为30，实际为%d", age)
				}

				active := node.Get("active").BoolOr(false)
				if !active {
					t.Error("期望active为true")
				}
			},
		},
		{
			name:    "嵌套对象路径访问",
			jsonStr: `{"user": {"profile": {"city": "北京", "country": "中国"}}}`,
			testFunc: func(t *testing.T, node Node) {
				city := node.GetPath("user.profile.city").StringOr("")
				if city != "北京" {
					t.Errorf("期望城市为'北京'，实际为'%s'", city)
				}

				country := node.GetPath("user.profile.country").StringOr("")
				if country != "中国" {
					t.Errorf("期望国家为'中国'，实际为'%s'", country)
				}
			},
		},
		{
			name:    "数组操作",
			jsonStr: `{"scores": [95, 87, 92, 88], "names": ["张三", "李四", "王五"]}`,
			testFunc: func(t *testing.T, node Node) {
				scores := node.Get("scores")
				if scores.Len() != 4 {
					t.Errorf("期望分数数组长度为4，实际为%d", scores.Len())
				}

				first := scores.Index(0).IntOr(0)
				if first != 95 {
					t.Errorf("期望第一个分数为95，实际为%d", first)
				}

				// 测试ToIntSlice
				scoreSlice, err := scores.ToIntSlice()
				if err != nil {
					t.Errorf("转换为int切片失败: %v", err)
				}
				if len(scoreSlice) != 4 {
					t.Errorf("期望int切片长度为4，实际为%d", len(scoreSlice))
				}

				// 测试ToStringSlice
				names := node.Get("names")
				nameSlice, err := names.ToStringSlice()
				if err != nil {
					t.Errorf("转换为string切片失败: %v", err)
				}
				if len(nameSlice) != 3 {
					t.Errorf("期望string切片长度为3，实际为%d", len(nameSlice))
				}
			},
		},
		{
			name:    "数据验证",
			jsonStr: `{"email": "user@example.com", "website": "https://example.com", "ip": "192.168.1.1", "age": 25, "score": 95.5}`,
			testFunc: func(t *testing.T, node Node) {
				email := node.Get("email")
				if !email.IsValidEmail() {
					t.Error("邮箱格式验证失败")
				}

				website := node.Get("website")
				if !website.IsValidURL() {
					t.Error("URL格式验证失败")
				}

				ip := node.Get("ip")
				if !ip.IsValidIP() {
					t.Error("IP地址格式验证失败")
				}

				age := node.Get("age")
				if !age.InRange(18, 65) {
					t.Error("年龄范围验证失败")
				}

				score := node.Get("score")
				if !score.InRange(0, 100) {
					t.Error("分数范围验证失败")
				}
			},
		},
		{
			name:    "类型检查",
			jsonStr: `{"name": "张三", "age": 30, "active": true, "address": null, "hobbies": ["阅读"], "profile": {"city": "北京"}}`,
			testFunc: func(t *testing.T, node Node) {
				if !node.Get("name").IsString() {
					t.Error("name应该是字符串类型")
				}

				if !node.Get("age").IsNumber() {
					t.Error("age应该是数字类型")
				}

				if !node.Get("active").IsBool() {
					t.Error("active应该是布尔类型")
				}

				if !node.Get("address").IsNull() {
					t.Error("address应该是null类型")
				}

				if !node.Get("hobbies").IsArray() {
					t.Error("hobbies应该是数组类型")
				}

				if !node.Get("profile").IsObject() {
					t.Error("profile应该是对象类型")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := FromString(tt.jsonStr)
			tt.testFunc(t, node)
		})
	}
}

// TestFromStringVsFromBytes 比较FromString和FromBytes的性能
func TestFromStringVsFromBytes(t *testing.T) {
	jsonStr := `{"users": [{"name": "张三", "age": 30, "scores": [95, 87, 92]}, {"name": "李四", "age": 25, "scores": [88, 91, 89]}]}`
	jsonBytes := []byte(jsonStr)

	const iterations = 10000

	// 测试FromString性能
	start := time.Now()
	for i := 0; i < iterations; i++ {
		node := FromString(jsonStr)
		_ = node.Get("users").Index(0).Get("name").StringOr("")
	}
	fromStringDuration := time.Since(start)

	// 测试FromBytes性能
	start = time.Now()
	for i := 0; i < iterations; i++ {
		node := FromBytes(jsonBytes)
		_ = node.Get("users").Index(0).Get("name").StringOr("")
	}
	fromBytesDuration := time.Since(start)

	t.Logf("FromString %d次操作耗时: %v", iterations, fromStringDuration)
	t.Logf("FromBytes %d次操作耗时: %v", iterations, fromBytesDuration)

	// FromString应该只是FromBytes的简单封装，性能差异应该很小
	ratio := float64(fromStringDuration) / float64(fromBytesDuration)
	if ratio > 1.5 {
		t.Errorf("FromString性能比FromBytes差太多，比率: %.2f", ratio)
	}
}

// TestHighPerformanceTraversal 测试高性能遍历
func TestHighPerformanceTraversal(t *testing.T) {
	jsonStr := `{
		"users": {
			"admin": {"name": "管理员", "level": 5, "permissions": ["read", "write", "delete"]},
			"guest": {"name": "访客", "level": 1, "permissions": ["read"]}
		},
		"stats": [100, 200, 300, 400, 500]
	}`

	node := FromString(jsonStr)

	// 测试ForEach遍历对象
	userCount := 0
	node.Get("users").ForEach(func(userType string, userInfo Node) bool {
		userCount++
		name := userInfo.Get("name").StringOr("")
		level := userInfo.Get("level").IntOr(0)

		if userType == "admin" && name != "管理员" {
			t.Errorf("期望管理员名字为'管理员'，实际为'%s'", name)
		}
		if userType == "admin" && level != 5 {
			t.Errorf("期望管理员等级为5，实际为%d", level)
		}

		return true
	})

	if userCount != 2 {
		t.Errorf("期望用户数量为2，实际为%d", userCount)
	}

	// 测试ArrayForEach遍历数组
	total := int64(0)
	itemCount := 0
	node.Get("stats").ArrayForEach(func(index int, item Node) bool {
		itemCount++
		value := item.IntOr(0)
		total += value
		return true
	})

	if itemCount != 5 {
		t.Errorf("期望统计项数量为5，实际为%d", itemCount)
	}

	expectedTotal := int64(1500) // 100+200+300+400+500
	if total != expectedTotal {
		t.Errorf("期望总和为%d，实际为%d", expectedTotal, total)
	}
}

// TestWalkFunctionality 测试Walk深度遍历功能
func TestWalkFunctionality(t *testing.T) {
	jsonStr := `{
		"company": "Tech Corp",
		"departments": {
			"engineering": {
				"count": 25,
				"teams": ["backend", "frontend"],
				"manager": {"name": "张经理", "level": 8}
			},
			"sales": {
				"count": 15,
				"teams": ["enterprise", "retail"]
			}
		}
	}`

	node := FromString(jsonStr)

	pathCount := 0
	stringCount := 0
	numberCount := 0
	arrayCount := 0

	node.Walk(func(path string, n Node) bool {
		pathCount++

		if n.IsString() {
			stringCount++
		} else if n.IsNumber() {
			numberCount++
		} else if n.IsArray() {
			arrayCount++
		}

		return true
	})

	if pathCount == 0 {
		t.Error("Walk应该遍历到至少一个路径")
	}

	if stringCount == 0 {
		t.Error("应该遍历到至少一个字符串类型的值")
	}

	if numberCount == 0 {
		t.Error("应该遍历到至少一个数字类型的值")
	}

	if arrayCount == 0 {
		t.Error("应该遍历到至少一个数组类型的值")
	}

	t.Logf("Walk遍历统计 - 总路径: %d, 字符串: %d, 数字: %d, 数组: %d",
		pathCount, stringCount, numberCount, arrayCount)
}

// TestKeysAndGetAllKeys 测试Keys方法
func TestKeysAndGetAllKeys(t *testing.T) {
	jsonStr := `{"name": "张三", "age": 30, "city": "北京", "active": true}`
	node := FromString(jsonStr)

	// 测试GetAllKeys
	keys := node.GetAllKeys()
	if len(keys) != 4 {
		t.Errorf("期望键数量为4，实际为%d", len(keys))
	}

	expectedKeys := map[string]bool{
		"name": true, "age": true, "city": true, "active": true,
	}

	for _, key := range keys {
		if !expectedKeys[key] {
			t.Errorf("意外的键: %s", key)
		}
	}

	// 测试Keys (新的API兼容方法)
	keysStr := node.Keys()
	if len(keysStr) != len(keys) {
		t.Errorf("Keys结果长度与GetAllKeys不一致")
	}
}

// BenchmarkFromStringPerformance FromString性能基准测试
func BenchmarkFromStringPerformance(b *testing.B) {
	jsonStr := `{"name": "张三", "age": 30, "scores": [95, 87, 92, 88], "active": true, "profile": {"city": "北京", "country": "中国"}}`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		node := FromString(jsonStr)
		_ = node.Get("name").StringOr("")
		_ = node.Get("age").IntOr(0)
		_ = node.GetPath("profile.city").StringOr("")
		_ = node.Get("scores").Len()
	}
}

// BenchmarkArrayForEachVsTraditionalLoop 比较ArrayForEach和传统循环的性能
func BenchmarkArrayForEachVsTraditionalLoop(b *testing.B) {
	jsonStr := `{"data": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20]}`
	node := FromString(jsonStr)
	data := node.Get("data")

	b.Run("ArrayForEach", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			total := int64(0)
			data.ArrayForEach(func(index int, item Node) bool {
				total += item.IntOr(0)
				return true
			})
			_ = total
		}
	})

	b.Run("TraditionalLoop", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			total := int64(0)
			for j := 0; j < data.Len(); j++ {
				total += data.Index(j).IntOr(0)
			}
			_ = total
		}
	})
}
