package fxjson

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// FieldMapper 字段映射配置
type FieldMapper struct {
	Rules         map[string]string      `json:"rules"`          // 字段映射规则
	DefaultValues map[string]interface{} `json:"default_values"` // 默认值
	TypeCast      map[string]string      `json:"type_cast"`      // 类型转换
}

// QueryBuilder 查询构建器
type QueryBuilder struct {
	node       Node
	conditions []Condition
	sortFields []SortField
	limitCount int
	offsetVal  int
}

// Condition 查询条件
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // =, !=, >, <, >=, <=, in, not_in, contains
	Value    interface{} `json:"value"`
}

// SortField 排序字段
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc, desc
}

// Aggregator 聚合器
type Aggregator struct {
	operations []AggOperation
	groupBy    []string
}

// AggOperation 聚合操作
type AggOperation struct {
	Type  string `json:"type"`  // count, sum, avg, max, min
	Field string `json:"field"` // 操作字段
	Alias string `json:"alias"` // 结果别名
}

// ValidationRule 验证规则
type ValidationRule struct {
	Required  bool                          `json:"required"`
	Type      string                        `json:"type"` // string, number, boolean, array, object
	MinLength int                           `json:"min_length"`
	MaxLength int                           `json:"max_length"`
	Min       float64                       `json:"min"`
	Max       float64                       `json:"max"`
	Pattern   string                        `json:"pattern"`
	Default   interface{}                   `json:"default"`
	Sanitize  func(interface{}) interface{} `json:"-"`
}

// DataValidator 数据验证器
type DataValidator struct {
	Rules map[string]ValidationRule `json:"rules"`
}

// Transform 数据变换
func (n Node) Transform(mapper FieldMapper) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 应用默认值
	for key, value := range mapper.DefaultValues {
		result[key] = value
	}

	// 应用字段映射规则
	for sourceField, targetField := range mapper.Rules {
		sourceNode := n.Get(sourceField)
		if sourceNode.Exists() {
			// 获取原始值
			var value interface{}
			switch sourceNode.Type() {
			case 's':
				value, _ = sourceNode.String()
			case 'n':
				// 检查是否需要类型转换
				if castType, exists := mapper.TypeCast[targetField]; exists {
					switch castType {
					case "int":
						value, _ = sourceNode.Int()
					case "float":
						value, _ = sourceNode.Float()
					default:
						value, _ = sourceNode.Float()
					}
				} else {
					value, _ = sourceNode.Float()
				}
			case 'b':
				value, _ = sourceNode.Bool()
			case 'a', 'o':
				value = sourceNode.Raw()
			}

			result[targetField] = value
		}
	}

	return result, nil
}

// Query 创建查询构建器
func (n Node) Query() *QueryBuilder {
	return &QueryBuilder{
		node:       n,
		conditions: make([]Condition, 0),
		sortFields: make([]SortField, 0),
		limitCount: -1,
		offsetVal:  0,
	}
}

// Where 添加查询条件
func (qb *QueryBuilder) Where(field, operator string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return qb
}

// WhereIn 检查字段值是否在指定列表中
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
	return qb.Where(field, "in", values)
}

// WhereNotIn 检查字段值是否不在指定列表中
func (qb *QueryBuilder) WhereNotIn(field string, values []interface{}) *QueryBuilder {
	return qb.Where(field, "not_in", values)
}

// WhereContains 检查字符串字段是否包含指定内容
func (qb *QueryBuilder) WhereContains(field, substring string) *QueryBuilder {
	return qb.Where(field, "contains", substring)
}

// SortBy 添加排序
func (qb *QueryBuilder) SortBy(field, order string) *QueryBuilder {
	qb.sortFields = append(qb.sortFields, SortField{
		Field: field,
		Order: order,
	})
	return qb
}

// Limit 限制结果数量
func (qb *QueryBuilder) Limit(count int) *QueryBuilder {
	qb.limitCount = count
	return qb
}

// Offset 设置偏移量
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offsetVal = offset
	return qb
}

// ToSlice 执行查询并返回结果
func (qb *QueryBuilder) ToSlice() ([]Node, error) {
	if qb.node.Type() != 'a' {
		return nil, fmt.Errorf("node is not an array")
	}

	var results []Node

	// 遍历数组元素
	for i := 0; i < qb.node.Len(); i++ {
		item := qb.node.Index(i)

		// 检查是否满足所有条件
		if qb.matchesConditions(item) {
			results = append(results, item)
		}
	}

	// 排序
	if len(qb.sortFields) > 0 {
		qb.sortResults(results)
	}

	// 应用偏移和限制
	start := qb.offsetVal
	if start < 0 {
		start = 0
	}
	if start >= len(results) {
		return []Node{}, nil
	}

	end := len(results)
	if qb.limitCount > 0 && start+qb.limitCount < end {
		end = start + qb.limitCount
	}

	return results[start:end], nil
}

// Count 计算匹配条件的数量
func (qb *QueryBuilder) Count() (int, error) {
	results, err := qb.ToSlice()
	if err != nil {
		return 0, err
	}
	return len(results), nil
}

// First 返回第一个匹配的元素
func (qb *QueryBuilder) First() (Node, error) {
	qb.limitCount = 1
	results, err := qb.ToSlice()
	if err != nil {
		return Node{}, err
	}
	if len(results) == 0 {
		return Node{}, fmt.Errorf("no matching elements found")
	}
	return results[0], nil
}

// matchesConditions 检查节点是否满足所有条件
func (qb *QueryBuilder) matchesConditions(node Node) bool {
	for _, condition := range qb.conditions {
		if !qb.evaluateCondition(node, condition) {
			return false
		}
	}
	return true
}

// evaluateCondition 评估单个条件
func (qb *QueryBuilder) evaluateCondition(node Node, condition Condition) bool {
	fieldNode := node.Get(condition.Field)
	if !fieldNode.Exists() {
		return condition.Operator == "!=" || condition.Operator == "not_in"
	}

	fieldValue := qb.getNodeValue(fieldNode)

	switch condition.Operator {
	case "=":
		return qb.compareValues(fieldValue, condition.Value) == 0
	case "!=":
		return qb.compareValues(fieldValue, condition.Value) != 0
	case ">":
		return qb.compareValues(fieldValue, condition.Value) > 0
	case "<":
		return qb.compareValues(fieldValue, condition.Value) < 0
	case ">=":
		return qb.compareValues(fieldValue, condition.Value) >= 0
	case "<=":
		return qb.compareValues(fieldValue, condition.Value) <= 0
	case "in":
		if values, ok := condition.Value.([]interface{}); ok {
			for _, v := range values {
				if qb.compareValues(fieldValue, v) == 0 {
					return true
				}
			}
		}
		return false
	case "not_in":
		if values, ok := condition.Value.([]interface{}); ok {
			for _, v := range values {
				if qb.compareValues(fieldValue, v) == 0 {
					return false
				}
			}
		}
		return true
	case "contains":
		if fieldStr, ok := fieldValue.(string); ok {
			if searchStr, ok := condition.Value.(string); ok {
				return strings.Contains(fieldStr, searchStr)
			}
		}
		return false
	}

	return false
}

// getNodeValue 获取节点的值
func (qb *QueryBuilder) getNodeValue(node Node) interface{} {
	switch node.Type() {
	case 's':
		if val, err := node.String(); err == nil {
			return val
		}
	case 'n':
		if val, err := node.Float(); err == nil {
			return val
		}
	case 'b':
		if val, err := node.Bool(); err == nil {
			return val
		}
	}
	return nil
}

// compareValues 比较两个值
func (qb *QueryBuilder) compareValues(a, b interface{}) int {
	// 类型转换和比较逻辑
	aVal := qb.normalizeValue(a)
	bVal := qb.normalizeValue(b)

	// 字符串比较
	if aStr, aOk := aVal.(string); aOk {
		if bStr, bOk := bVal.(string); bOk {
			return strings.Compare(aStr, bStr)
		}
	}

	// 数值比较
	if aNum, aOk := aVal.(float64); aOk {
		if bNum, bOk := bVal.(float64); bOk {
			if aNum < bNum {
				return -1
			} else if aNum > bNum {
				return 1
			}
			return 0
		}
	}

	// 布尔值比较
	if aBool, aOk := aVal.(bool); aOk {
		if bBool, bOk := bVal.(bool); bOk {
			if aBool == bBool {
				return 0
			} else if aBool {
				return 1
			}
			return -1
		}
	}

	return 0
}

// normalizeValue 标准化值类型
func (qb *QueryBuilder) normalizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int())
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint())
	case float32:
		return float64(v)
	case string:
		// 尝试转换为数字
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num
		}
		return v
	default:
		return value
	}
}

// sortResults 对结果进行排序
func (qb *QueryBuilder) sortResults(results []Node) {
	sort.Slice(results, func(i, j int) bool {
		for _, sortField := range qb.sortFields {
			iVal := qb.getNodeValue(results[i].Get(sortField.Field))
			jVal := qb.getNodeValue(results[j].Get(sortField.Field))

			cmp := qb.compareValues(iVal, jVal)
			if cmp != 0 {
				if sortField.Order == "desc" {
					return cmp > 0
				}
				return cmp < 0
			}
		}
		return false
	})
}

// Aggregate 创建聚合器
func (n Node) Aggregate() *Aggregator {
	return &Aggregator{
		operations: make([]AggOperation, 0),
		groupBy:    make([]string, 0),
	}
}

// Count 计数聚合
func (agg *Aggregator) Count(alias string) *Aggregator {
	agg.operations = append(agg.operations, AggOperation{
		Type:  "count",
		Alias: alias,
	})
	return agg
}

// Sum 求和聚合
func (agg *Aggregator) Sum(field, alias string) *Aggregator {
	agg.operations = append(agg.operations, AggOperation{
		Type:  "sum",
		Field: field,
		Alias: alias,
	})
	return agg
}

// Avg 平均值聚合
func (agg *Aggregator) Avg(field, alias string) *Aggregator {
	agg.operations = append(agg.operations, AggOperation{
		Type:  "avg",
		Field: field,
		Alias: alias,
	})
	return agg
}

// Max 最大值聚合
func (agg *Aggregator) Max(field, alias string) *Aggregator {
	agg.operations = append(agg.operations, AggOperation{
		Type:  "max",
		Field: field,
		Alias: alias,
	})
	return agg
}

// Min 最小值聚合
func (agg *Aggregator) Min(field, alias string) *Aggregator {
	agg.operations = append(agg.operations, AggOperation{
		Type:  "min",
		Field: field,
		Alias: alias,
	})
	return agg
}

// GroupBy 分组
func (agg *Aggregator) GroupBy(fields ...string) *Aggregator {
	agg.groupBy = append(agg.groupBy, fields...)
	return agg
}

// Execute 执行聚合操作
func (agg *Aggregator) Execute(node Node) (map[string]interface{}, error) {
	if node.Type() != 'a' {
		return nil, fmt.Errorf("node must be an array for aggregation")
	}

	result := make(map[string]interface{})

	// 如果没有分组，直接对所有数据聚合
	if len(agg.groupBy) == 0 {
		return agg.executeSimpleAggregation(node)
	}

	// 分组聚合
	groups := make(map[string][]Node)

	for i := 0; i < node.Len(); i++ {
		item := node.Index(i)
		groupKey := agg.buildGroupKey(item)
		groups[groupKey] = append(groups[groupKey], item)
	}

	// 对每个分组执行聚合
	for groupKey, groupItems := range groups {
		groupResult := make(map[string]interface{})

		for _, op := range agg.operations {
			value, err := agg.executeOperation(op, groupItems)
			if err != nil {
				return nil, err
			}
			groupResult[op.Alias] = value
		}

		result[groupKey] = groupResult
	}

	return result, nil
}

// executeSimpleAggregation 执行简单聚合（无分组）
func (agg *Aggregator) executeSimpleAggregation(node Node) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 转换为Node切片
	items := make([]Node, node.Len())
	for i := 0; i < node.Len(); i++ {
		items[i] = node.Index(i)
	}

	for _, op := range agg.operations {
		value, err := agg.executeOperation(op, items)
		if err != nil {
			return nil, err
		}
		result[op.Alias] = value
	}

	return result, nil
}

// buildGroupKey 构建分组键
func (agg *Aggregator) buildGroupKey(item Node) string {
	var keyParts []string
	for _, field := range agg.groupBy {
		value := item.Get(field)
		if valueStr, err := value.String(); err == nil {
			keyParts = append(keyParts, valueStr)
		} else {
			keyParts = append(keyParts, "null")
		}
	}
	return strings.Join(keyParts, "|")
}

// executeOperation 执行单个聚合操作
func (agg *Aggregator) executeOperation(op AggOperation, items []Node) (interface{}, error) {
	switch op.Type {
	case "count":
		return len(items), nil

	case "sum":
		var sum float64
		for _, item := range items {
			if val, err := item.Get(op.Field).Float(); err == nil {
				sum += val
			}
		}
		return sum, nil

	case "avg":
		var sum float64
		var count int
		for _, item := range items {
			if val, err := item.Get(op.Field).Float(); err == nil {
				sum += val
				count++
			}
		}
		if count == 0 {
			return 0, nil
		}
		return sum / float64(count), nil

	case "max":
		var max float64
		var hasValue bool
		for _, item := range items {
			if val, err := item.Get(op.Field).Float(); err == nil {
				if !hasValue || val > max {
					max = val
					hasValue = true
				}
			}
		}
		if !hasValue {
			return nil, nil
		}
		return max, nil

	case "min":
		var min float64
		var hasValue bool
		for _, item := range items {
			if val, err := item.Get(op.Field).Float(); err == nil {
				if !hasValue || val < min {
					min = val
					hasValue = true
				}
			}
		}
		if !hasValue {
			return nil, nil
		}
		return min, nil

	default:
		return nil, fmt.Errorf("unknown aggregation operation: %s", op.Type)
	}
}

// Validate 数据验证
func (n Node) Validate(validator *DataValidator) (map[string]interface{}, []error) {
	result := make(map[string]interface{})
	var errors []error

	for fieldName, rule := range validator.Rules {
		fieldNode := n.Get(fieldName)

		// 检查必填字段
		if rule.Required && !fieldNode.Exists() {
			errors = append(errors, fmt.Errorf("field '%s' is required", fieldName))
			continue
		}

		// 应用默认值
		if !fieldNode.Exists() && rule.Default != nil {
			result[fieldName] = rule.Default
			continue
		}

		if !fieldNode.Exists() {
			continue
		}

		// 验证和转换值
		value, err := validateAndConvertField(fieldNode, rule)
		if err != nil {
			errors = append(errors, fmt.Errorf("field '%s': %w", fieldName, err))
			continue
		}

		// 应用清理函数
		if rule.Sanitize != nil {
			value = rule.Sanitize(value)
		}

		result[fieldName] = value
	}

	return result, errors
}

// validateAndConvertField 验证和转换字段值
func validateAndConvertField(node Node, rule ValidationRule) (interface{}, error) {
	switch rule.Type {
	case "string":
		value, err := node.String()
		if err != nil {
			return nil, err
		}

		if rule.MinLength > 0 && len(value) < rule.MinLength {
			return nil, fmt.Errorf("string too short, minimum length is %d", rule.MinLength)
		}

		if rule.MaxLength > 0 && len(value) > rule.MaxLength {
			return nil, fmt.Errorf("string too long, maximum length is %d", rule.MaxLength)
		}

		return value, nil

	case "number":
		value, err := node.Float()
		if err != nil {
			return nil, err
		}

		if rule.Min != 0 && value < rule.Min {
			return nil, fmt.Errorf("number too small, minimum is %f", rule.Min)
		}

		if rule.Max != 0 && value > rule.Max {
			return nil, fmt.Errorf("number too large, maximum is %f", rule.Max)
		}

		return value, nil

	case "boolean":
		return node.Bool()

	default:
		// 原样返回
		switch node.Type() {
		case 's':
			return node.String()
		case 'n':
			return node.Float()
		case 'b':
			return node.Bool()
		default:
			return node.Raw(), nil
		}
	}
}

// Stream 流式处理
func (n Node) Stream(processor func(Node, int) bool) error {
	if n.Type() != 'a' {
		return fmt.Errorf("node must be an array for streaming")
	}

	for i := 0; i < n.Len(); i++ {
		item := n.Index(i)
		if !processor(item, i) {
			break
		}
	}

	return nil
}
