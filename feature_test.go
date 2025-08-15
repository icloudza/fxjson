package fxjson

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// 测试用的复杂JSON数据
const testComplexJSON = `{
  "code": 0,
  "success": true,
  "msg": "数据获取成功",
  "timestamp": 1692345600,
  "data": {
    "user_profile": {
      "user_id": "user_123",
      "nickname": "创作者小王",
      "level": 5,
      "verified": true,
      "followers_count": 15230,
      "following_count": 892
    },
    "notes": [
      {
        "id": "note_001",
        "title": "美食探店｜这家餐厅太棒了",
        "type": "normal",
        "view_count": 12580,
        "likes": 892,
        "comments_count": 156,
        "shares": 78,
        "collected_count": 245,
        "created_time": "2025-08-14 10:30:00",
        "status": "published",
        "category": "food",
        "tags": ["美食", "探店", "推荐"],
        "images_count": 8,
        "video_duration": 0,
        "location": "上海市黄浦区",
        "engagement_rate": 7.2,
        "revenue": 156.80
      },
      {
        "id": "note_002", 
        "title": "今日穿搭分享",
        "type": "video",
        "view_count": 8920,
        "likes": 445,
        "comments_count": 89,
        "shares": 34,
        "collected_count": 123,
        "created_time": "2025-08-13 15:20:00",
        "status": "published",
        "category": "fashion",
        "tags": ["穿搭", "时尚", "分享"],
        "images_count": 0,
        "video_duration": 45,
        "location": "北京市朝阳区",
        "engagement_rate": 5.8,
        "revenue": 89.50
      },
      {
        "id": "note_003",
        "title": "旅行日记 - 三亚行",
        "type": "normal",
        "view_count": 25670,
        "likes": 1250,
        "comments_count": 234,
        "shares": 156,
        "collected_count": 445,
        "created_time": "2025-08-12 09:15:00",
        "status": "published",
        "category": "travel",
        "tags": ["旅行", "三亚", "度假"],
        "images_count": 12,
        "video_duration": 0,
        "location": "海南省三亚市",
        "engagement_rate": 8.9,
        "revenue": 445.30
      },
      {
        "id": "note_004",
        "title": "护肤心得分享",
        "type": "normal", 
        "view_count": 15230,
        "likes": 678,
        "comments_count": 123,
        "shares": 67,
        "collected_count": 289,
        "created_time": "2025-08-11 20:45:00",
        "status": "published",
        "category": "beauty",
        "tags": ["护肤", "美容", "心得"],
        "images_count": 6,
        "video_duration": 0,
        "location": "广州市天河区",
        "engagement_rate": 6.5,
        "revenue": 234.70
      },
      {
        "id": "note_005",
        "title": "健身日常记录",
        "type": "video",
        "view_count": 6780,
        "likes": 234,
        "comments_count": 45,
        "shares": 23,
        "collected_count": 89,
        "created_time": "2025-08-10 07:30:00",
        "status": "draft",
        "category": "fitness",
        "tags": ["健身", "运动", "日常"],
        "images_count": 0,
        "video_duration": 120,
        "location": "深圳市福田区",
        "engagement_rate": 4.2,
        "revenue": 67.20
      }
    ],
    "analytics": {
      "total_notes": 5,
      "total_views": 68180,
      "total_likes": 3499,
      "total_comments": 647,
      "total_revenue": 993.50,
      "avg_engagement_rate": 6.52,
      "period": "last_7_days"
    }
  }
}`

// 包含空字符串的测试数据
const testEmptyStringJSON = `{
  "code": 0,
  "success": true,
  "msg": "",
  "data": {
    "notes": [
      {
        "id": "note_001",
        "title": "",
        "permission_msg": "",
        "xsec_token": "",
        "xsec_source": "",
        "type": "",
        "images_list": [
          {"url": ""},
          {"url": ""},
          {"url": ""}
        ]
      }
    ]
  }
}`

// TestDataTransformation 测试数据变换功能
func TestDataTransformation(t *testing.T) {
	fmt.Println("\n🔄 测试数据变换功能")
	fmt.Println(strings.Repeat("-", 50))

	node := FromBytes([]byte(testComplexJSON))

	// 定义字段映射规则
	mapper := FieldMapper{
		Rules: map[string]string{
			"data.user_profile.user_id":         "uid",
			"data.user_profile.nickname":        "name",
			"data.user_profile.followers_count": "fans",
			"data.user_profile.level":           "user_level",
			"data.analytics.total_views":        "total_pv",
			"data.analytics.total_revenue":      "total_income",
		},
		DefaultValues: map[string]interface{}{
			"status":     "active",
			"created_by": "system",
		},
		TypeCast: map[string]string{
			"user_level": "int",
			"fans":       "int",
		},
	}

	// 执行数据变换
	result, err := node.Transform(mapper)
	if err != nil {
		t.Fatalf("数据变换失败: %v", err)
	}

	// 验证结果
	if result["uid"] != "user_123" {
		t.Errorf("用户ID映射错误: 期望 'user_123', 实际 '%v'", result["uid"])
	}

	if result["name"] != "创作者小王" {
		t.Errorf("用户名映射错误: 期望 '创作者小王', 实际 '%v'", result["name"])
	}

	if result["status"] != "active" {
		t.Errorf("默认值设置错误: 期望 'active', 实际 '%v'", result["status"])
	}

	fmt.Printf("✅ 数据变换成功，映射了 %d 个字段\n", len(result))
	for key, value := range result {
		fmt.Printf("   %s: %v\n", key, value)
	}
}

// TestConditionalQueries 测试条件查询功能
func TestConditionalQueries(t *testing.T) {
	fmt.Println("\n🔍 测试条件查询功能")
	fmt.Println(strings.Repeat("-", 50))

	node := FromBytes([]byte(testComplexJSON))
	notesList := node.Get("data.notes")

	// 测试1: 查询高浏览量笔记
	highViewNotes, err := notesList.Query().
		Where("view_count", ">", 10000).
		Where("status", "=", "published").
		SortBy("view_count", "desc").
		ToSlice()

	if err != nil {
		t.Fatalf("高浏览量查询失败: %v", err)
	}

	if len(highViewNotes) != 3 {
		t.Errorf("高浏览量笔记数量错误: 期望 3, 实际 %d", len(highViewNotes))
	}

	fmt.Printf("✅ 高浏览量查询成功，找到 %d 篇笔记\n", len(highViewNotes))
	for i, note := range highViewNotes {
		title, _ := note.Get("title").String()
		viewCount, _ := note.Get("view_count").Int()
		fmt.Printf("   [%d] %s - %d浏览\n", i+1, title, viewCount)
	}

	// 测试2: 查询视频类型笔记
	videoNotes, err := notesList.Query().
		Where("type", "=", "video").
		ToSlice()

	if err != nil {
		t.Fatalf("视频查询失败: %v", err)
	}

	if len(videoNotes) != 2 {
		t.Errorf("视频笔记数量错误: 期望 2, 实际 %d", len(videoNotes))
	}

	fmt.Printf("✅ 视频类型查询成功，找到 %d 篇视频\n", len(videoNotes))

	// 测试3: 统计查询
	count, err := notesList.Query().
		Where("engagement_rate", ">", 6.0).
		Count()

	if err != nil {
		t.Fatalf("统计查询失败: %v", err)
	}

	if count != 3 {
		t.Errorf("高互动率笔记统计错误: 期望 3, 实际 %d", count)
	}

	fmt.Printf("✅ 统计查询成功，高互动率笔记: %d篇\n", count)

	// 测试4: 第一个匹配项查询
	firstNote, err := notesList.Query().
		Where("category", "=", "travel").
		First()

	if err != nil {
		t.Fatalf("第一个匹配项查询失败: %v", err)
	}

	title, _ := firstNote.Get("title").String()
	if !strings.Contains(title, "旅行") {
		t.Errorf("第一个旅行笔记标题错误: %s", title)
	}

	fmt.Printf("✅ 第一个匹配项查询成功: %s\n", title)
}

// TestDataAggregation 测试数据聚合功能
func TestDataAggregation(t *testing.T) {
	fmt.Println("\n📈 测试数据聚合功能")
	fmt.Println(strings.Repeat("-", 50))

	node := FromBytes([]byte(testComplexJSON))
	notesList := node.Get("data.notes")

	// 测试1: 按类型分组统计
	categoryStats, err := notesList.Aggregate().
		GroupBy("category").
		Count("total_notes").
		Sum("view_count", "total_views").
		Sum("likes", "total_likes").
		Sum("revenue", "total_revenue").
		Avg("engagement_rate", "avg_engagement").
		Max("view_count", "max_views").
		Execute(notesList)

	if err != nil {
		t.Fatalf("分类聚合失败: %v", err)
	}

	// 验证聚合结果
	if len(categoryStats) != 5 { // food, fashion, travel, beauty, fitness
		t.Errorf("分类数量错误: 期望 5, 实际 %d", len(categoryStats))
	}

	fmt.Printf("✅ 按类型分组聚合成功，%d个分类:\n", len(categoryStats))
	for category, stats := range categoryStats {
		if statsMap, ok := stats.(map[string]interface{}); ok {
			fmt.Printf("   📁 %s类:\n", category)
			fmt.Printf("      笔记数量: %.0f\n", statsMap["total_notes"])
			fmt.Printf("      总浏览量: %.0f\n", statsMap["total_views"])
			fmt.Printf("      总点赞数: %.0f\n", statsMap["total_likes"])
			fmt.Printf("      总收入: %.2f元\n", statsMap["total_revenue"])
			fmt.Printf("      平均互动率: %.2f%%\n", statsMap["avg_engagement"])
			fmt.Printf("      最高浏览量: %.0f\n", statsMap["max_views"])
		}
	}

	// 测试2: 全局统计（无分组）
	globalStats, err := notesList.Aggregate().
		Count("total_count").
		Sum("view_count", "total_views").
		Sum("revenue", "total_revenue").
		Avg("engagement_rate", "avg_engagement").
		Min("engagement_rate", "min_engagement").
		Max("engagement_rate", "max_engagement").
		Execute(notesList)

	if err != nil {
		t.Fatalf("全局聚合失败: %v", err)
	}

	fmt.Printf("\n✅ 全局统计聚合成功:\n")
	for _, stats := range globalStats {
		if statsMap, ok := stats.(map[string]interface{}); ok {
			fmt.Printf("   📝 总笔记数: %.0f\n", statsMap["total_count"])
			fmt.Printf("   👀 总浏览量: %.0f\n", statsMap["total_views"])
			fmt.Printf("   💰 总收入: %.2f元\n", statsMap["total_revenue"])
			fmt.Printf("   📊 平均互动率: %.2f%%\n", statsMap["avg_engagement"])
			fmt.Printf("   📉 最低互动率: %.2f%%\n", statsMap["min_engagement"])
			fmt.Printf("   📈 最高互动率: %.2f%%\n", statsMap["max_engagement"])
		}
		break // 只有一个结果
	}
}

// TestCachePerformance 测试缓存性能功能
func TestCachePerformance(t *testing.T) {
	fmt.Println("\n⚡ 测试缓存性能功能")
	fmt.Println(strings.Repeat("-", 50))

	// 创建缓存
	cache := NewMemoryCache(10)
	EnableCaching(cache)

	// 第一次解析（无缓存）
	start := time.Now()
	node1 := FromBytesWithCache([]byte(testComplexJSON), 5*time.Minute)
	firstParseTime := time.Since(start)

	// 第二次解析（使用缓存）
	start = time.Now()
	node2 := FromBytesWithCache([]byte(testComplexJSON), 5*time.Minute)
	secondParseTime := time.Since(start)

	// 验证缓存效果
	if secondParseTime >= firstParseTime {
		t.Logf("警告: 缓存性能提升不明显 - 第一次: %v, 第二次: %v", firstParseTime, secondParseTime)
	}

	// 验证结果一致性
	title1, _ := node1.Get("data.notes[0].title").String()
	title2, _ := node2.Get("data.notes[0].title").String()
	if title1 != title2 {
		t.Errorf("缓存结果不一致: %s != %s", title1, title2)
	}

	// 检查缓存统计
	cacheStats := cache.Stats()
	if cacheStats.Hits < 1 {
		t.Errorf("缓存命中次数应该 >= 1, 实际: %d", cacheStats.Hits)
	}

	fmt.Printf("✅ 缓存性能测试成功:\n")
	fmt.Printf("   第一次解析: %v\n", firstParseTime)
	fmt.Printf("   缓存解析: %v\n", secondParseTime)
	if secondParseTime < firstParseTime {
		fmt.Printf("   性能提升: %.2fx\n", float64(firstParseTime)/float64(secondParseTime))
	}
	fmt.Printf("   缓存命中: %d次\n", cacheStats.Hits)
	fmt.Printf("   缓存未命中: %d次\n", cacheStats.Misses)
	fmt.Printf("   命中率: %.2f%%\n", cacheStats.HitRate*100)

	// 测试批处理
	fmt.Printf("\n✅ 测试批处理功能:\n")
	processedCount := 0
	processor := NewBatchProcessor(2, func(nodes []Node) error {
		processedCount += len(nodes)
		fmt.Printf("   处理批次: %d个节点\n", len(nodes))
		return nil
	})

	notesList := node1.Get("data.notes")
	notesList.ArrayForEach(func(index int, note Node) bool {
		processor.Add(note)
		return true
	})
	processor.Flush()

	if processedCount != notesList.Len() {
		t.Errorf("批处理数量错误: 期望 %d, 实际 %d", notesList.Len(), processedCount)
	}
	fmt.Printf("   批处理完成，共处理 %d个节点\n", processedCount)
}

// TestDebugFeatures 测试调试功能
func TestDebugFeatures(t *testing.T) {
	fmt.Println("\n🔍 测试调试功能")
	fmt.Println(strings.Repeat("-", 50))

	// 启用调试模式
	EnableDebugMode()
	defer DisableDebugMode()

	// 带调试信息的解析
	node, debugInfo := FromBytesWithDebug([]byte(testComplexJSON))

	// 验证调试信息
	if debugInfo.NodeCount < 10 {
		t.Errorf("节点数量过少: %d", debugInfo.NodeCount)
	}

	if debugInfo.MaxDepth < 3 {
		t.Errorf("最大深度过浅: %d", debugInfo.MaxDepth)
	}

	fmt.Printf("✅ 调试信息收集成功:\n")
	fmt.Printf("   解析时间: %v\n", debugInfo.ParseTime)
	fmt.Printf("   内存使用: %d bytes\n", debugInfo.MemoryUsage)
	fmt.Printf("   节点数量: %d\n", debugInfo.NodeCount)
	fmt.Printf("   最大深度: %d\n", debugInfo.MaxDepth)

	if len(debugInfo.PerformanceHints) > 0 {
		fmt.Printf("   性能建议: %d条\n", len(debugInfo.PerformanceHints))
		for _, hint := range debugInfo.PerformanceHints {
			fmt.Printf("     - %s\n", hint)
		}
	}

	// 测试节点检查
	userProfile := node.Get("data.user_profile")
	inspection := userProfile.Inspect()

	fmt.Printf("\n✅ 节点检查功能:\n")
	fmt.Printf("   类型: %v\n", inspection["type"])
	fmt.Printf("   存在: %v\n", inspection["exists"])
	fmt.Printf("   键数量: %v\n", inspection["key_count"])

	// 测试美化打印
	prettyJSON := userProfile.PrettyPrint()
	if len(prettyJSON) < 50 {
		t.Errorf("美化打印结果过短: %d", len(prettyJSON))
	}

	fmt.Printf("\n✅ 美化打印测试:\n")
	fmt.Printf("   输出长度: %d字符\n", len(prettyJSON))
	fmt.Printf("   前100字符: %s...\n", prettyJSON[:min(100, len(prettyJSON))])

	// 测试JSON差异对比
	modifiedJSON := strings.Replace(testComplexJSON, `"level": 5`, `"level": 6`, 1)
	modifiedNode := FromBytes([]byte(modifiedJSON))

	diffs := node.Get("data.user_profile").Diff(modifiedNode.Get("data.user_profile"))

	fmt.Printf("\n✅ JSON差异对比:\n")
	fmt.Printf("   发现差异: %d处\n", len(diffs))
	for _, diff := range diffs {
		fmt.Printf("     %s: %s %v -> %v\n", diff.Path, diff.Type, diff.OldValue, diff.NewValue)
	}
}

// TestDataValidation 测试数据验证功能
func TestDataValidation(t *testing.T) {
	fmt.Println("\n✅ 测试数据验证功能")
	fmt.Println(strings.Repeat("-", 50))

	node := FromBytes([]byte(testComplexJSON))
	notesList := node.Get("data.notes")
	firstNote := notesList.Index(0)

	// 定义验证规则
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
			"engagement_rate": {
				Required: true,
				Type:     "number",
				Min:      0,
				Max:      100,
			},
		},
	}

	// 验证数据
	result, errors := firstNote.Validate(validator)

	if len(errors) > 0 {
		t.Errorf("验证失败: %v", errors)
	}

	// 验证结果包含必要字段
	if result["title"] == nil {
		t.Error("验证结果缺少标题字段")
	}

	if result["view_count"] == nil {
		t.Error("验证结果缺少浏览量字段")
	}

	fmt.Printf("✅ 数据验证成功:\n")
	fmt.Printf("   验证字段数: %d\n", len(result))
	fmt.Printf("   错误数量: %d\n", len(errors))
	for key, value := range result {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// 测试验证失败情况
	invalidValidator := &DataValidator{
		Rules: map[string]ValidationRule{
			"view_count": {
				Required: true,
				Type:     "number",
				Min:      100000, // 设置一个过高的最小值
			},
		},
	}

	_, invalidErrors := firstNote.Validate(invalidValidator)
	if len(invalidErrors) == 0 {
		t.Error("期望验证失败，但验证通过了")
	}

	fmt.Printf("✅ 验证失败测试成功，产生 %d个错误\n", len(invalidErrors))
}

// TestEmptyStringHandling 测试空字符串处理
func TestEmptyStringHandling(t *testing.T) {
	fmt.Println("\n🔧 测试空字符串处理功能")
	fmt.Println(strings.Repeat("-", 50))

	// 解析包含空字符串的JSON（这以前会panic）
	node := FromBytes([]byte(testEmptyStringJSON))

	notesList := node.Get("data.notes")
	if notesList.Len() != 1 {
		t.Fatalf("笔记数量错误: 期望 1, 实际 %d", notesList.Len())
	}

	firstNote := notesList.Index(0)

	// 获取所有空字符串字段
	title, err := firstNote.Get("title").String()
	if err != nil {
		t.Errorf("获取空标题失败: %v", err)
	}
	if title != "" {
		t.Errorf("空标题应该是空字符串, 实际: '%s'", title)
	}

	permissionMsg, err := firstNote.Get("permission_msg").String()
	if err != nil {
		t.Errorf("获取空权限消息失败: %v", err)
	}
	if permissionMsg != "" {
		t.Errorf("空权限消息应该是空字符串, 实际: '%s'", permissionMsg)
	}

	// 测试结构体解码（这以前会panic）
	type TestNote struct {
		Id            string `json:"id"`
		Title         string `json:"title"`
		PermissionMsg string `json:"permission_msg"`
		XsecToken     string `json:"xsec_token"`
		XsecSource    string `json:"xsec_source"`
		Type          string `json:"type"`
	}

	var note TestNote
	err = firstNote.Decode(&note)
	if err != nil {
		t.Fatalf("空字符串结构体解码失败: %v", err)
	}

	// 验证所有字段都是空字符串（除了id）
	if note.Id != "note_001" {
		t.Errorf("ID字段错误: 期望 'note_001', 实际 '%s'", note.Id)
	}

	if note.Title != "" {
		t.Errorf("标题应该是空字符串, 实际: '%s'", note.Title)
	}

	fmt.Printf("✅ 空字符串处理测试成功:\n")
	fmt.Printf("   ID: '%s'\n", note.Id)
	fmt.Printf("   标题: '%s'\n", note.Title)
	fmt.Printf("   权限消息: '%s'\n", note.PermissionMsg)
	fmt.Printf("   XSec Token: '%s'\n", note.XsecToken)
	fmt.Printf("   XSec Source: '%s'\n", note.XsecSource)
	fmt.Printf("   类型: '%s'\n", note.Type)

	// 测试图片列表中的空字符串
	imagesList := firstNote.Get("images_list")
	if imagesList.Len() != 3 {
		t.Errorf("图片列表长度错误: 期望 3, 实际 %d", imagesList.Len())
	}

	for i := 0; i < imagesList.Len(); i++ {
		url, err := imagesList.Index(i).Get("url").String()
		if err != nil {
			t.Errorf("获取图片URL失败: %v", err)
		}
		if url != "" {
			t.Errorf("图片URL应该是空字符串, 实际: '%s'", url)
		}
	}

	fmt.Printf("✅ 图片列表空字符串处理成功，%d个空URL\n", imagesList.Len())
}

// TestStreamProcessing 测试流式处理功能
func TestStreamProcessing(t *testing.T) {
	fmt.Println("\n📊 测试流式处理功能")
	fmt.Println(strings.Repeat("-", 50))

	node := FromBytes([]byte(testComplexJSON))
	notesList := node.Get("data.notes")

	// 测试流式处理
	processedCount := 0
	totalViews := int64(0)

	err := notesList.Stream(func(note Node, index int) bool {
		processedCount++
		viewCount, _ := note.Get("view_count").Int()
		totalViews += viewCount

		title, _ := note.Get("title").String()
		fmt.Printf("   [%d] 处理笔记: %s (浏览: %d)\n", index+1, title, viewCount)

		// 测试提前终止
		if index == 2 {
			fmt.Printf("   提前终止处理\n")
			return false
		}
		return true
	})

	if err != nil {
		t.Fatalf("流式处理失败: %v", err)
	}

	// 验证处理了3个笔记（在第3个时提前终止）
	if processedCount != 3 {
		t.Errorf("处理数量错误: 期望 3, 实际 %d", processedCount)
	}

	fmt.Printf("✅ 流式处理成功:\n")
	fmt.Printf("   处理数量: %d篇笔记\n", processedCount)
	fmt.Printf("   累计浏览: %d次\n", totalViews)
}

// 辅助函数：获取最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
