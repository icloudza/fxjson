# 查询和聚合

FxJSON 提供了强大的数据查询和聚合功能，包括内置的查询构建器和聚合器，以及高性能的遍历和数据访问，让您能够高效地从 JSON 数据中提取和分析信息。

> **提示：** FxJSON 提供了内置的 `node.Query()` 查询构建器和 `node.Aggregate()` 聚合器（详见 [API 文档](/api/)），本指南展示了多种数据查询和聚合的方法。

## 基础数据查询

### 使用内置查询构建器

```go
func builtinQueryExample() {
    productsData := `{
        "products": [
            {"id": 1, "name": "笔记本电脑", "price": 5999, "category": "电子产品", "stock": 10, "rating": 4.5, "active": true},
            {"id": 2, "name": "智能手机", "price": 2999, "category": "电子产品", "stock": 25, "rating": 4.3, "active": true},
            {"id": 3, "name": "蓝牙耳机", "price": 299, "category": "配件", "stock": 50, "rating": 4.1, "active": false},
            {"id": 4, "name": "机械键盘", "price": 699, "category": "配件", "stock": 15, "rating": 4.7, "active": true},
            {"id": 5, "name": "显示器", "price": 1999, "category": "电子产品", "stock": 8, "rating": 4.4, "active": true}
        ]
    }`

    node := fxjson.FromBytes([]byte(productsData))
    products := node.Get("products")

    // 使用查询构建器查找符合条件的产品
    results, err := products.Query().
        Where("price", ">", 1000).
        Where("active", "=", true).
        Where("category", "=", "电子产品").
        SortBy("price", "desc").
        Limit(3).
        ToSlice()
    
    if err != nil {
        fmt.Printf("查询错误: %v\n", err)
        return
    }

    fmt.Printf("找到 %d 个符合条件的产品:\n", len(results))
    for _, product := range results {
        name := product.Get("name").StringOr("")
        price := product.Get("price").FloatOr(0)
        fmt.Printf("- %s: ¥%.0f\n", name, price)
    }

    // 使用聚合器计算统计信息
    stats, err := products.Aggregate().
        Count("total_products").
        Sum("price", "total_value").
        Avg("price", "avg_price").
        Max("rating", "highest_rating").
        Min("stock", "lowest_stock").
        Execute(products)
    
    if err != nil {
        fmt.Printf("聚合错误: %v\n", err)
        return
    }

    fmt.Println("统计信息:")
    for key, value := range stats {
        fmt.Printf("- %s: %v\n", key, value)
    }
}
```

### 简单条件查询

```go
func basicQuery() {
    productsData := `{
        "products": [
            {"id": 1, "name": "笔记本电脑", "price": 5999, "category": "电子产品", "stock": 10, "rating": 4.5},
            {"id": 2, "name": "智能手机", "price": 2999, "category": "电子产品", "stock": 25, "rating": 4.3},
            {"id": 3, "name": "蓝牙耳机", "price": 299, "category": "配件", "stock": 50, "rating": 4.1},
            {"id": 4, "name": "机械键盘", "price": 699, "category": "配件", "stock": 15, "rating": 4.7},
            {"id": 5, "name": "显示器", "price": 1999, "category": "电子产品", "stock": 8, "rating": 4.4}
        ]
    }`

    node := fxjson.FromBytes([]byte(productsData))
    products := node.Get("products")

    // 查找高价商品（价格大于2000）
    var highPriceProducts []fxjson.Node
    products.ArrayForEach(func(index int, product fxjson.Node) bool {
        price := product.Get("price").FloatOr(0)
        if price > 2000 {
            highPriceProducts = append(highPriceProducts, product)
        }
        return true
    })

    fmt.Printf("高价商品数量: %d\n", len(highPriceProducts))
    for _, product := range highPriceProducts {
        name := product.Get("name").StringOr("")
        price := product.Get("price").FloatOr(0)
        fmt.Printf("- %s: ¥%.0f\n", name, price)
    }

    // 查找电子产品中库存充足的商品
    var electronicProducts []fxjson.Node
    products.ArrayForEach(func(index int, product fxjson.Node) bool {
        category := product.Get("category").StringOr("")
        stock := product.Get("stock").IntOr(0)
        if category == "电子产品" && stock > 10 {
            electronicProducts = append(electronicProducts, product)
        }
        return true
    })

    fmt.Printf("库存充足的电子产品: %d\n", len(electronicProducts))

    // 查找名称包含"键盘"的商品
    var keyboardProducts []fxjson.Node
    products.ArrayForEach(func(index int, product fxjson.Node) bool {
        name := product.Get("name").StringOr("")
        if strings.Contains(name, "键盘") {
            keyboardProducts = append(keyboardProducts, product)
        }
        return true
    })

    fmt.Printf("键盘类商品: %d\n", len(keyboardProducts))
}
```

### 复杂条件查询

```go
func advancedQuery() {
    salesData := `{
        "sales": [
            {"id": 1, "product": "笔记本", "amount": 5999, "quantity": 2, "date": "2024-01-15", "region": "北京"},
            {"id": 2, "product": "手机", "amount": 2999, "quantity": 1, "date": "2024-01-16", "region": "上海"},
            {"id": 3, "product": "耳机", "amount": 299, "quantity": 5, "date": "2024-01-17", "region": "广州"},
            {"id": 4, "product": "键盘", "amount": 699, "quantity": 3, "date": "2024-01-18", "region": "北京"},
            {"id": 5, "product": "鼠标", "amount": 199, "quantity": 10, "date": "2024-01-19", "region": "深圳"},
            {"id": 6, "product": "显示器", "amount": 1999, "quantity": 1, "date": "2024-01-20", "region": "上海"}
        ]
    }`

    node := fxjson.FromBytes([]byte(salesData))
    sales := node.Get("sales")

    // 查找北京和上海的销售记录
    var beijingShanghaiSales []fxjson.Node
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        region := sale.Get("region").StringOr("")
        if region == "北京" || region == "上海" {
            beijingShanghaiSales = append(beijingShanghaiSales, sale)
        }
        return true
    })
    fmt.Printf("北京和上海的销售记录: %d\n", len(beijingShanghaiSales))

    // 查找非北京的销售记录
    var nonBeijingSales []fxjson.Node
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        region := sale.Get("region").StringOr("")
        if region != "北京" {
            nonBeijingSales = append(nonBeijingSales, sale)
        }
        return true
    })
    fmt.Printf("非北京的销售记录: %d\n", len(nonBeijingSales))

    // 复杂组合查询：金额大于1000或数量大于等于5，且不在深圳
    var complexResults []fxjson.Node
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        amount := sale.Get("amount").FloatOr(0)
        quantity := sale.Get("quantity").IntOr(0)
        region := sale.Get("region").StringOr("")
        
        if (amount > 1000 || quantity >= 5) && region != "深圳" {
            complexResults = append(complexResults, sale)
        }
        return true
    })
    fmt.Printf("复杂查询结果: %d\n", len(complexResults))

    // 日期范围查询
    var dateRangeResults []fxjson.Node
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        date := sale.Get("date").StringOr("")
        if date >= "2024-01-16" && date <= "2024-01-18" {
            dateRangeResults = append(dateRangeResults, sale)
        }
        return true
    })
    fmt.Printf("指定日期范围销售: %d\n", len(dateRangeResults))
}
```

## 数据排序和分页

### 自定义排序

```go
func sortingAndPaging() {
    productsData := `{
        "products": [
            {"id": 1, "name": "笔记本电脑", "price": 5999, "rating": 4.5},
            {"id": 2, "name": "智能手机", "price": 2999, "rating": 4.3},
            {"id": 3, "name": "蓝牙耳机", "price": 299, "rating": 4.1},
            {"id": 4, "name": "机械键盘", "price": 699, "rating": 4.7}
        ]
    }`

    node := fxjson.FromBytes([]byte(productsData))
    products := node.Get("products")

    // 收集所有产品到切片中
    var productList []fxjson.Node
    products.ArrayForEach(func(index int, product fxjson.Node) bool {
        productList = append(productList, product)
        return true
    })

    // 按价格降序排列
    sort.Slice(productList, func(i, j int) bool {
        priceI := productList[i].Get("price").FloatOr(0)
        priceJ := productList[j].Get("price").FloatOr(0)
        return priceI > priceJ
    })

    fmt.Println("按价格降序排列:")
    for i, product := range productList {
        if i >= 3 { // 只显示前3个
            break
        }
        name := product.Get("name").StringOr("")
        price := product.Get("price").FloatOr(0)
        fmt.Printf("%d. %s: ¥%.0f\n", i+1, name, price)
    }

    // 按评分降序排列
    sort.Slice(productList, func(i, j int) bool {
        ratingI := productList[i].Get("rating").FloatOr(0)
        ratingJ := productList[j].Get("rating").FloatOr(0)
        return ratingI > ratingJ
    })

    fmt.Println("\n按评分降序排列:")
    for i, product := range productList {
        name := product.Get("name").StringOr("")
        rating := product.Get("rating").FloatOr(0)
        fmt.Printf("%d. %s: %.1f\n", i+1, name, rating)
    }

    // 分页处理
    pageSize := 2
    totalPages := (len(productList) + pageSize - 1) / pageSize
    
    fmt.Printf("\n分页结果 (每页%d个，共%d页):\n", pageSize, totalPages)
    for page := 0; page < totalPages; page++ {
        start := page * pageSize
        end := start + pageSize
        if end > len(productList) {
            end = len(productList)
        }
        
        fmt.Printf("第%d页:\n", page+1)
        for i := start; i < end; i++ {
            name := productList[i].Get("name").StringOr("")
            fmt.Printf("  - %s\n", name)
        }
    }
}
```

## 数据聚合分析

### 基础聚合计算

```go
func basicAggregation() {
    salesData := `{
        "sales": [
            {"product": "笔记本", "amount": 5999, "quantity": 2, "region": "北京"},
            {"product": "手机", "amount": 2999, "quantity": 5, "region": "上海"},
            {"product": "耳机", "amount": 299, "quantity": 10, "region": "广州"},
            {"product": "键盘", "amount": 699, "quantity": 3, "region": "北京"},
            {"product": "鼠标", "amount": 199, "quantity": 8, "region": "深圳"}
        ]
    }`

    node := fxjson.FromBytes([]byte(salesData))
    sales := node.Get("sales")

    // 计算总销售额
    totalAmount := 0.0
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        amount := sale.Get("amount").FloatOr(0)
        totalAmount += amount
        return true
    })

    // 计算平均销售额
    recordCount := sales.Len()
    avgAmount := totalAmount / float64(recordCount)

    // 找出最大和最小销售额
    maxAmount := 0.0
    minAmount := math.Inf(1)
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        amount := sale.Get("amount").FloatOr(0)
        if amount > maxAmount {
            maxAmount = amount
        }
        if amount < minAmount {
            minAmount = amount
        }
        return true
    })

    fmt.Printf("总销售额: ¥%.0f\n", totalAmount)
    fmt.Printf("平均销售额: ¥%.2f\n", avgAmount)
    fmt.Printf("最大单笔: ¥%.0f\n", maxAmount)
    fmt.Printf("最小单笔: ¥%.0f\n", minAmount)
    fmt.Printf("销售记录数: %d\n", recordCount)

    // 计算总销售数量
    totalQuantity := 0.0
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        quantity := sale.Get("quantity").FloatOr(0)
        totalQuantity += quantity
        return true
    })

    avgQuantity := totalQuantity / float64(recordCount)
    fmt.Printf("总销售数量: %.0f\n", totalQuantity)
    fmt.Printf("平均销售数量: %.2f\n", avgQuantity)
}
```

### 分组聚合分析

```go
func groupAggregation() {
    salesData := `{
        "sales": [
            {"product": "笔记本", "amount": 5999, "quantity": 2, "region": "北京", "category": "电子产品"},
            {"product": "手机", "amount": 2999, "quantity": 5, "region": "上海", "category": "电子产品"},
            {"product": "耳机", "amount": 299, "quantity": 10, "region": "广州", "category": "配件"},
            {"product": "键盘", "amount": 699, "quantity": 3, "region": "北京", "category": "配件"},
            {"product": "鼠标", "amount": 199, "quantity": 8, "region": "深圳", "category": "配件"},
            {"product": "显示器", "amount": 1999, "quantity": 1, "region": "上海", "category": "电子产品"}
        ]
    }`

    node := fxjson.FromBytes([]byte(salesData))
    sales := node.Get("sales")

    // 按地区分组聚合
    regionStats := make(map[string]map[string]float64)
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        region := sale.Get("region").StringOr("")
        amount := sale.Get("amount").FloatOr(0)
        quantity := sale.Get("quantity").FloatOr(0)

        if regionStats[region] == nil {
            regionStats[region] = make(map[string]float64)
        }

        regionStats[region]["total_amount"] += amount
        regionStats[region]["total_quantity"] += quantity
        regionStats[region]["count"] += 1

        return true
    })

    fmt.Println("按地区分组统计:")
    for region, stats := range regionStats {
        fmt.Printf("%s:\n", region)
        fmt.Printf("  总销售额: ¥%.0f\n", stats["total_amount"])
        fmt.Printf("  总数量: %.0f\n", stats["total_quantity"])
        fmt.Printf("  订单数: %.0f\n", stats["count"])
        fmt.Printf("  平均单价: ¥%.2f\n", stats["total_amount"]/stats["count"])
        fmt.Println()
    }

    // 按类别分组聚合
    categoryStats := groupBy(sales, "category", []string{"amount", "quantity"})
    fmt.Println("按类别分组统计:")
    for category, stats := range categoryStats {
        fmt.Printf("%s: 总额¥%.0f, 总量%.0f, 订单%.0f\n", 
            category, stats["amount_sum"], stats["quantity_sum"], stats["count"])
    }
}

func groupBy(node fxjson.Node, groupField string, aggregateFields []string) map[string]map[string]float64 {
    groups := make(map[string]map[string]float64)

    node.ArrayForEach(func(index int, item fxjson.Node) bool {
        groupValue := item.Get(groupField).StringOr("unknown")

        if groups[groupValue] == nil {
            groups[groupValue] = make(map[string]float64)
        }

        // 计数
        groups[groupValue]["count"] += 1

        // 聚合指定字段
        for _, field := range aggregateFields {
            value := item.Get(field).FloatOr(0)
            groups[groupValue][field+"_sum"] += value
            groups[groupValue][field+"_avg"] = groups[groupValue][field+"_sum"] / groups[groupValue]["count"]
            
            if groups[groupValue][field+"_max"] < value || groups[groupValue]["count"] == 1 {
                groups[groupValue][field+"_max"] = value
            }
            
            if groups[groupValue][field+"_min"] > value || groups[groupValue]["count"] == 1 {
                groups[groupValue][field+"_min"] = value
            }
        }

        return true
    })

    return groups
}
```

## 数据分析示例

### 销售数据分析

```go
func salesAnalysis() {
    // 生成模拟销售数据
    salesData := generateSalesData()
    node := fxjson.FromBytes([]byte(salesData))
    sales := node.Get("sales")

    // 时间序列分析
    monthlyStats := analyzeMonthlyTrends(sales)
    fmt.Println("月度趋势分析:")
    for month, stats := range monthlyStats {
        fmt.Printf("%s: 销售额¥%.0f, 增长率%.2f%%\n", 
            month, stats["amount"], stats["growth_rate"])
    }

    // 产品分析
    productPerformance := analyzeProductPerformance(sales)
    fmt.Println("\n产品性能分析:")
    for _, product := range productPerformance {
        fmt.Printf("产品: %s, 销售额: ¥%.0f, 利润率: %.2f%%\n",
            product["name"], product["revenue"], product["profit_margin"])
    }

    // 地区分析
    regionAnalysis := analyzeRegionalPerformance(sales)
    fmt.Println("\n区域分析:")
    for region, metrics := range regionAnalysis {
        fmt.Printf("%s: 市场份额%.2f%%, 平均订单¥%.2f\n",
            region, metrics["market_share"], metrics["avg_order"])
    }
}

func analyzeMonthlyTrends(sales fxjson.Node) map[string]map[string]float64 {
    monthlyData := make(map[string][]float64)
    
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        date := sale.Get("date").StringOr("")
        amount := sale.Get("amount").FloatOr(0)
        
        if len(date) >= 7 {
            month := date[:7] // YYYY-MM
            monthlyData[month] = append(monthlyData[month], amount)
        }
        return true
    })

    results := make(map[string]map[string]float64)
    var previousAmount float64
    
    for month, amounts := range monthlyData {
        total := 0.0
        for _, amount := range amounts {
            total += amount
        }
        
        growthRate := 0.0
        if previousAmount > 0 {
            growthRate = (total - previousAmount) / previousAmount * 100
        }
        
        results[month] = map[string]float64{
            "amount":      total,
            "count":       float64(len(amounts)),
            "avg_order":   total / float64(len(amounts)),
            "growth_rate": growthRate,
        }
        
        previousAmount = total
    }
    
    return results
}

func analyzeProductPerformance(sales fxjson.Node) []map[string]interface{} {
    productStats := make(map[string]map[string]float64)
    
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        product := sale.Get("product").StringOr("")
        amount := sale.Get("amount").FloatOr(0)
        cost := sale.Get("cost").FloatOr(amount * 0.7) // 假设成本为70%
        
        if productStats[product] == nil {
            productStats[product] = make(map[string]float64)
        }
        
        productStats[product]["revenue"] += amount
        productStats[product]["cost"] += cost
        productStats[product]["orders"] += 1
        
        return true
    })

    var results []map[string]interface{}
    for product, stats := range productStats {
        profit := stats["revenue"] - stats["cost"]
        profitMargin := 0.0
        if stats["revenue"] > 0 {
            profitMargin = profit / stats["revenue"] * 100
        }
        
        results = append(results, map[string]interface{}{
            "name":          product,
            "revenue":       stats["revenue"],
            "profit":        profit,
            "profit_margin": profitMargin,
            "orders":        stats["orders"],
            "avg_order":     stats["revenue"] / stats["orders"],
        })
    }
    
    return results
}

func analyzeRegionalPerformance(sales fxjson.Node) map[string]map[string]float64 {
    regionStats := make(map[string]map[string]float64)
    totalRevenue := 0.0
    
    // 第一遍：收集基础数据
    sales.ArrayForEach(func(index int, sale fxjson.Node) bool {
        region := sale.Get("region").StringOr("")
        amount := sale.Get("amount").FloatOr(0)
        
        if regionStats[region] == nil {
            regionStats[region] = make(map[string]float64)
        }
        
        regionStats[region]["revenue"] += amount
        regionStats[region]["orders"] += 1
        totalRevenue += amount
        
        return true
    })
    
    // 第二遍：计算比率和平均值
    for region, stats := range regionStats {
        stats["market_share"] = (stats["revenue"] / totalRevenue) * 100
        stats["avg_order"] = stats["revenue"] / stats["orders"]
    }
    
    return regionStats
}

func generateSalesData() string {
    return `{
        "sales": [
            {"product": "笔记本", "amount": 5999, "date": "2024-01-15", "region": "北京", "cost": 4199},
            {"product": "手机", "amount": 2999, "date": "2024-01-16", "region": "上海", "cost": 2099},
            {"product": "耳机", "amount": 299, "date": "2024-02-17", "region": "广州", "cost": 209},
            {"product": "键盘", "amount": 699, "date": "2024-02-18", "region": "北京", "cost": 489},
            {"product": "鼠标", "amount": 199, "date": "2024-03-19", "region": "深圳", "cost": 139}
        ]
    }`
}
```

## 性能优化

### 大数据量查询优化

```go
func optimizedQuerying() {
    // 生成大量数据进行性能测试
    largeDataset := generateLargeDataset(10000)
    node := fxjson.FromBytes([]byte(largeDataset))
    products := node.Get("products")
    
    // 使用早期终止优化查询
    start := time.Now()
    var expensiveProducts []fxjson.Node
    products.ArrayForEach(func(index int, product fxjson.Node) bool {
        price := product.Get("price").FloatOr(0)
        if price > 5000 {
            expensiveProducts = append(expensiveProducts, product)
            // 如果找到足够的结果，可以提前终止
            if len(expensiveProducts) >= 10 {
                return false // 终止遍历
            }
        }
        return true
    })
    duration := time.Since(start)
    
    fmt.Printf("找到 %d 个昂贵商品，耗时: %v\n", len(expensiveProducts), duration)
    
    // 使用批量处理优化大数据查询
    batchSize := 1000
    var results []fxjson.Node
    
    start = time.Now()
    totalCount := products.Len()
    for i := 0; i < totalCount; i += batchSize {
        end := i + batchSize
        if end > totalCount {
            end = totalCount
        }
        
        // 处理当前批次
        for j := i; j < end; j++ {
            product := products.Index(j)
            category := product.Get("category").StringOr("")
            if category == "电子产品" {
                results = append(results, product)
            }
        }
    }
    batchDuration := time.Since(start)
    
    fmt.Printf("批量处理找到 %d 个电子产品，耗时: %v\n", len(results), batchDuration)
}

func generateLargeDataset(count int) string {
    var products []string
    categories := []string{"电子产品", "配件", "服装", "食品", "图书"}
    
    for i := 0; i < count; i++ {
        price := 100 + (i % 10000)
        category := categories[i%len(categories)]
        
        product := fmt.Sprintf(`{
            "id": %d,
            "name": "商品%d",
            "price": %d,
            "category": "%s",
            "stock": %d
        }`, i, i, price, category, 10+(i%100))
        
        products = append(products, product)
    }
    
    return fmt.Sprintf(`{"products": [%s]}`, strings.Join(products, ","))
}
```

通过这些查询和聚合功能，FxJSON 让您能够高效地从 JSON 数据中提取有价值的信息和洞察，满足各种数据分析需求。