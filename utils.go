package fxjson

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ==================== 默认值支持函数 ====================

// StringOr 获取字符串值，如果失败返回默认值
func (n Node) StringOr(defaultValue string) string {
	if str, err := n.String(); err == nil {
		return str
	}
	return defaultValue
}

// IntOr 获取整数值，如果失败返回默认值
func (n Node) IntOr(defaultValue int64) int64 {
	if val, err := n.Int(); err == nil {
		return val
	}
	return defaultValue
}

// FloatOr 获取浮点数值，如果失败返回默认值
func (n Node) FloatOr(defaultValue float64) float64 {
	if val, err := n.Float(); err == nil {
		return val
	}
	return defaultValue
}

// BoolOr 获取布尔值，如果失败返回默认值
func (n Node) BoolOr(defaultValue bool) bool {
	if val, err := n.Bool(); err == nil {
		return val
	}
	return defaultValue
}

// UintOr 获取无符号整数值，如果失败返回默认值
func (n Node) UintOr(defaultValue uint64) uint64 {
	if val, err := n.Uint(); err == nil {
		return val
	}
	return defaultValue
}

// ==================== 批量获取函数 ====================

// GetMultiple 同时获取多个路径的值
func (n Node) GetMultiple(paths ...string) []Node {
	results := make([]Node, len(paths))
	for i, path := range paths {
		results[i] = n.GetPath(path)
	}
	return results
}

// HasAnyPath 检查是否存在任意一个路径
func (n Node) HasAnyPath(paths ...string) bool {
	for _, path := range paths {
		if n.GetPath(path).Exists() {
			return true
		}
	}
	return false
}

// HasAllPaths 检查是否存在所有路径
func (n Node) HasAllPaths(paths ...string) bool {
	for _, path := range paths {
		if !n.GetPath(path).Exists() {
			return false
		}
	}
	return true
}

// ==================== 数据转换工具 ====================

// ToStringSlice 将数组转换为字符串切片
func (n Node) ToStringSlice() ([]string, error) {
	if !n.IsArray() {
		return nil, &FxJSONError{
			Type:    ErrorTypeTypeMismatch,
			Message: "node is not an array",
		}
	}

	result := make([]string, 0, n.Len())
	n.ArrayForEach(func(index int, value Node) bool {
		if str, err := value.String(); err == nil {
			result = append(result, str)
		}
		return true
	})

	return result, nil
}

// ToIntSlice 将数组转换为整数切片
func (n Node) ToIntSlice() ([]int64, error) {
	if !n.IsArray() {
		return nil, &FxJSONError{
			Type:    ErrorTypeTypeMismatch,
			Message: "node is not an array",
		}
	}

	result := make([]int64, 0, n.Len())
	n.ArrayForEach(func(index int, value Node) bool {
		if val, err := value.Int(); err == nil {
			result = append(result, val)
		}
		return true
	})

	return result, nil
}

// ToFloatSlice 将数组转换为浮点数切片
func (n Node) ToFloatSlice() ([]float64, error) {
	if !n.IsArray() {
		return nil, &FxJSONError{
			Type:    ErrorTypeTypeMismatch,
			Message: "node is not an array",
		}
	}

	result := make([]float64, 0, n.Len())
	n.ArrayForEach(func(index int, value Node) bool {
		if val, err := value.Float(); err == nil {
			result = append(result, val)
		}
		return true
	})

	return result, nil
}

// ToBoolSlice 将数组转换为布尔值切片
func (n Node) ToBoolSlice() ([]bool, error) {
	if !n.IsArray() {
		return nil, &FxJSONError{
			Type:    ErrorTypeTypeMismatch,
			Message: "node is not an array",
		}
	}

	result := make([]bool, 0, n.Len())
	n.ArrayForEach(func(index int, value Node) bool {
		if val, err := value.Bool(); err == nil {
			result = append(result, val)
		}
		return true
	})

	return result, nil
}

// ==================== 数据验证工具 ====================

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	uuidRegex  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// IsValidEmail 检查字符串是否为有效的电子邮件地址
func (n Node) IsValidEmail() bool {
	if str, err := n.String(); err == nil {
		return emailRegex.MatchString(str)
	}
	return false
}

// IsValidURL 检查字符串是否为有效的URL
func (n Node) IsValidURL() bool {
	if str, err := n.String(); err == nil {
		_, err := url.Parse(str)
		if err != nil {
			return false
		}
		// 额外检查是否有scheme和host
		u, _ := url.Parse(str)
		return u.Scheme != "" && u.Host != ""
	}
	return false
}

// IsValidPhone 检查字符串是否为有效的电话号码（E.164格式）
func (n Node) IsValidPhone() bool {
	if str, err := n.String(); err == nil {
		return phoneRegex.MatchString(str)
	}
	return false
}

// IsValidUUID 检查字符串是否为有效的UUID
func (n Node) IsValidUUID() bool {
	if str, err := n.String(); err == nil {
		return uuidRegex.MatchString(str)
	}
	return false
}

// IsValidIPv4 检查字符串是否为有效的IPv4地址
func (n Node) IsValidIPv4() bool {
	if str, err := n.String(); err == nil {
		// 简单的IPv4验证
		parts := strings.Split(str, ".")
		if len(parts) != 4 {
			return false
		}
		for _, part := range parts {
			if len(part) == 0 || len(part) > 3 {
				return false
			}
			num := 0
			for _, ch := range part {
				if ch < '0' || ch > '9' {
					return false
				}
				num = num*10 + int(ch-'0')
			}
			if num > 255 {
				return false
			}
		}
		return true
	}
	return false
}

// IsValidIPv6 检查字符串是否为有效的IPv6地址
func (n Node) IsValidIPv6() bool {
	if str, err := n.String(); err == nil {
		// 简单的IPv6验证 - 检查是否包含冒号和十六进制字符
		if !strings.Contains(str, ":") {
			return false
		}
		// 移除IPv6中可能的zone信息
		if idx := strings.Index(str, "%"); idx != -1 {
			str = str[:idx]
		}
		// 展开 :: 缩写
		if strings.Contains(str, "::") {
			// 简单验证：确保只有一个 ::
			if strings.Count(str, "::") > 1 {
				return false
			}
		}
		// 分割并验证每个部分
		parts := strings.Split(str, ":")
		if len(parts) > 8 {
			return false
		}
		for _, part := range parts {
			if len(part) == 0 {
				continue // 允许空部分（::的情况）
			}
			if len(part) > 4 {
				return false
			}
			for _, ch := range part {
				if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
					return false
				}
			}
		}
		return true
	}
	return false
}

// IsValidIP 检查字符串是否为有效的IP地址（IPv4或IPv6）
func (n Node) IsValidIP() bool {
	return n.IsValidIPv4() || n.IsValidIPv6()
}

// IsValidJSON 验证 JSON 格式
func (n Node) IsValidJSON() bool {
	// 如果节点本身就是有效的JSON结构（对象、数组等），则直接返回true
	if n.IsObject() || n.IsArray() || n.IsBool() || n.IsNumber() || n.IsNull() {
		return true
	}

	// 对于字符串节点，检查其内容是否为有效JSON
	if n.IsString() {
		str, err := n.String()
		if err != nil {
			return false
		}

		// 空字符串不是有效的JSON
		if len(str) == 0 {
			return false
		}

		// 尝试解析JSON来验证格式
		testNode := FromString(str)
		return testNode.Exists()
	}

	return false
}

// isNumericString 检查字符串是否为有效数字
func isNumericString(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// ==================== 分组类型检查 ====================
// 注意：IsScalar和IsContainer方法已存在于fxjson.go中

// ==================== 字符串操作工具 ====================

// Contains 检查字符串是否包含子串
func (n Node) Contains(substr string) bool {
	if str, err := n.String(); err == nil {
		return strings.Contains(str, substr)
	}
	return false
}

// StartsWith 检查字符串是否以指定前缀开始
func (n Node) StartsWith(prefix string) bool {
	if str, err := n.String(); err == nil {
		return strings.HasPrefix(str, prefix)
	}
	return false
}

// EndsWith 检查字符串是否以指定后缀结束
func (n Node) EndsWith(suffix string) bool {
	if str, err := n.String(); err == nil {
		return strings.HasSuffix(str, suffix)
	}
	return false
}

// ToLower 将字符串转换为小写
func (n Node) ToLower() (string, error) {
	if str, err := n.String(); err == nil {
		return strings.ToLower(str), nil
	} else {
		return "", err
	}
}

// ToUpper 将字符串转换为大写
func (n Node) ToUpper() (string, error) {
	if str, err := n.String(); err == nil {
		return strings.ToUpper(str), nil
	} else {
		return "", err
	}
}

// Trim 去除字符串两端的空白字符
func (n Node) Trim() (string, error) {
	if str, err := n.String(); err == nil {
		return strings.TrimSpace(str), nil
	} else {
		return "", err
	}
}

// ==================== 数组操作工具 ====================

// First 获取数组的第一个元素
func (n Node) First() Node {
	if n.IsArray() && n.Len() > 0 {
		return n.Index(0)
	}
	return Node{}
}

// Last 获取数组的最后一个元素
func (n Node) Last() Node {
	if n.IsArray() && n.Len() > 0 {
		return n.Index(n.Len() - 1)
	}
	return Node{}
}

// Slice 获取数组的切片（包含start，不包含end）
func (n Node) Slice(start, end int) []Node {
	if !n.IsArray() {
		return nil
	}

	length := n.Len()
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start >= end {
		return []Node{}
	}

	result := make([]Node, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, n.Index(i))
	}
	return result
}

// Reverse 返回反转后的数组节点
func (n Node) Reverse() []Node {
	if !n.IsArray() {
		return nil
	}

	length := n.Len()
	result := make([]Node, length)
	for i := 0; i < length; i++ {
		result[length-1-i] = n.Index(i)
	}
	return result
}

// ==================== 对象操作工具 ====================

// Merge 合并两个对象节点（浅合并）
func (n Node) Merge(other Node) map[string]Node {
	result := make(map[string]Node)

	// 添加当前节点的所有键值对
	if n.IsObject() {
		n.ForEach(func(key string, value Node) bool {
			result[key] = value
			return true
		})
	}

	// 添加或覆盖other节点的键值对
	if other.IsObject() {
		other.ForEach(func(key string, value Node) bool {
			result[key] = value
			return true
		})
	}

	return result
}

// Pick 从对象中选择指定的键
func (n Node) Pick(keys ...string) map[string]Node {
	result := make(map[string]Node)
	if !n.IsObject() {
		return result
	}

	keySet := make(map[string]bool)
	for _, key := range keys {
		keySet[key] = true
	}

	n.ForEach(func(key string, value Node) bool {
		if keySet[key] {
			result[key] = value
		}
		return true
	})

	return result
}

// Omit 从对象中排除指定的键
func (n Node) Omit(keys ...string) map[string]Node {
	result := make(map[string]Node)
	if !n.IsObject() {
		return result
	}

	keySet := make(map[string]bool)
	for _, key := range keys {
		keySet[key] = true
	}

	n.ForEach(func(key string, value Node) bool {
		if !keySet[key] {
			result[key] = value
		}
		return true
	})

	return result
}

// ==================== 比较工具 ====================

// Equals 检查两个节点是否相等
func (n Node) Equals(other Node) bool {
	// 比较类型
	if n.typ != other.typ {
		return false
	}

	// 比较原始数据
	nData := n.getWorkingData()
	oData := other.getWorkingData()

	if len(nData[n.start:n.end]) != len(oData[other.start:other.end]) {
		return false
	}

	for i := 0; i < n.end-n.start; i++ {
		if nData[n.start+i] != oData[other.start+i] {
			return false
		}
	}

	return true
}

// IsEmpty 检查节点是否为空（空字符串、空数组、空对象、null）
func (n Node) IsEmpty() bool {
	switch n.typ {
	case 's':
		str, _ := n.String()
		return str == ""
	case 'a':
		return n.Len() == 0
	case 'o':
		return n.Len() == 0
	case 'l':
		return true
	default:
		return !n.Exists()
	}
}

// ==================== 数字操作工具 ====================

// IsPositive 检查数字是否为正数
func (n Node) IsPositive() bool {
	if n.IsNumber() {
		if val, err := n.Float(); err == nil {
			return val > 0
		}
	}
	return false
}

// IsNegative 检查数字是否为负数
func (n Node) IsNegative() bool {
	if n.IsNumber() {
		if val, err := n.Float(); err == nil {
			return val < 0
		}
	}
	return false
}

// IsZero 检查数字是否为零
func (n Node) IsZero() bool {
	if n.IsNumber() {
		if val, err := n.Float(); err == nil {
			return val == 0
		}
	}
	return false
}

// IsInteger 检查数字是否为整数
func (n Node) IsInteger() bool {
	if n.IsNumber() {
		if val, err := n.Float(); err == nil {
			return val == float64(int64(val))
		}
	}
	return false
}

// InRange 检查数字是否在指定范围内（包含边界）
func (n Node) InRange(min, max float64) bool {
	if n.IsNumber() {
		if val, err := n.Float(); err == nil {
			return val >= min && val <= max
		}
	}
	return false
}
