package fxjson

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

const (
	maxInt64U = uint64(9223372036854775807)  // 2^63-1
	minInt64U = uint64(9223372036854775808)  // -(min int64) 的绝对值
	maxUint64 = uint64(18446744073709551615) // 2^64-1
)

type Node struct {
	raw      []byte
	start    int
	end      int
	typ      byte   // 'o' 'a' 's' 'n' 'b' 'l'
	expanded []byte // 存储展开后的JSON数据
}

// JsonParam 用于控制 JSON 输出的格式化参数
type JsonParam struct {
	Indent     int  // 缩进空格数；0 表示紧凑模式（不换行不缩进），>0 表示每层缩进的空格数量
	EscapeHTML bool // 是否转义 HTML 符号（< > &）；true 时会输出 \u003C \u003E \u0026
	Precision  int  // 浮点数精度；-1 表示原样输出，>=0 表示保留的小数位数（四舍五入）
}

type NodeType byte

const (
	TypeInvalid NodeType = 0
	TypeObject  NodeType = 'o'
	TypeArray   NodeType = 'a'
	TypeString  NodeType = 's'
	TypeNumber  NodeType = 'n'
	TypeBool    NodeType = 'b'
	TypeNull    NodeType = 'l'
)

// ===== 数组下标缓存（无锁、全局、键为底层数据指针+范围）=====

type arrKey struct {
	data uintptr
	s, e int
}

var arrIdxCache sync.Map // map[arrKey][]int

func dataPtr(b []byte) uintptr {
	if len(b) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(unsafe.SliceData(b)))
}

func buildArrOffsetsCached(n Node) []int {
	if n.typ != 'a' || n.start >= n.end {
		return nil
	}

	// 使用展开后的数据
	data := n.getWorkingData()
	key := arrKey{data: dataPtr(data), s: n.start, e: n.end}
	if v, ok := arrIdxCache.Load(key); ok {
		return v.([]int)
	}

	pos := n.start + 1 // skip '['
	var offs []int
	for pos < n.end {
		for pos < n.end && data[pos] <= ' ' {
			pos++
		}
		if pos >= n.end || data[pos] == ']' {
			break
		}
		offs = append(offs, pos)
		pos = skipValueFast(data, pos, n.end)
		for pos < n.end && data[pos] <= ' ' {
			pos++
		}
		if pos < n.end && data[pos] == ',' {
			pos++
		}
	}
	arrIdxCache.Store(key, offs)
	return offs
}

// getWorkingData 返回用于工作的数据（优先使用展开后的数据）
func (n Node) getWorkingData() []byte {
	if len(n.expanded) > 0 {
		return n.expanded
	}
	return n.raw
}

// ===== 转义处理相关函数 =====

// unescapeJSON 解转义JSON字符串
func unescapeJSON(s string) string {
	if !strings.Contains(s, "\\") {
		return s
	}

	var result strings.Builder
	result.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] != '\\' {
			result.WriteByte(s[i])
			i++
			continue
		}

		if i+1 >= len(s) {
			result.WriteByte(s[i])
			i++
			continue
		}

		switch s[i+1] {
		case '"':
			result.WriteByte('"')
			i += 2
		case '\\':
			result.WriteByte('\\')
			i += 2
		case '/':
			result.WriteByte('/')
			i += 2
		case 'b':
			result.WriteByte('\b')
			i += 2
		case 'f':
			result.WriteByte('\f')
			i += 2
		case 'n':
			result.WriteByte('\n')
			i += 2
		case 'r':
			result.WriteByte('\r')
			i += 2
		case 't':
			result.WriteByte('\t')
			i += 2
		case 'u':
			if i+5 < len(s) {
				// 简化处理：直接跳过unicode转义
				result.WriteString(s[i : i+6])
				i += 6
			} else {
				result.WriteByte(s[i])
				i++
			}
		default:
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String()
}

// isValidJSON 检查字符串是否为有效的JSON
func isValidJSON(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	return (s[0] == '{' && s[len(s)-1] == '}') || (s[0] == '[' && s[len(s)-1] == ']')
}

// expandNestedJSON 递归展开嵌套的JSON字符串
func expandNestedJSON(data []byte) []byte {
	node := parseRootNode(data)
	if !node.Exists() {
		return data
	}

	expanded, changed := expandNode(node)
	if !changed {
		return data
	}

	return expanded
}

// expandNode 展开单个节点
func expandNode(n Node) ([]byte, bool) {
	data := n.getWorkingData()

	switch n.typ {
	case 'o':
		return expandObject(n, data)
	case 'a':
		return expandArray(n, data)
	case 's':
		return expandString(n, data)
	default:
		return data[n.start:n.end], false
	}
}

// expandObject 展开对象
func expandObject(n Node, data []byte) ([]byte, bool) {
	var result strings.Builder
	result.WriteByte('{')

	pos := n.start + 1 // skip '{'
	changed := false
	first := true

	for pos < n.end {
		// 跳过空白
		for pos < n.end && data[pos] <= ' ' {
			pos++
		}
		if pos >= n.end || data[pos] == '}' {
			break
		}

		if !first {
			result.WriteByte(',')
		}
		first = false

		// 解析键
		if data[pos] != '"' {
			break
		}

		keyStart := pos
		pos++
		for pos < n.end && data[pos] != '"' {
			if data[pos] == '\\' {
				pos++
			}
			pos++
		}
		pos++ // skip closing quote

		result.Write(data[keyStart:pos])

		// 跳过冒号
		for pos < n.end && data[pos] <= ' ' {
			pos++
		}
		if pos < n.end && data[pos] == ':' {
			pos++
			result.WriteByte(':')
		}
		for pos < n.end && data[pos] <= ' ' {
			pos++
		}

		// 解析值
		valueNode := parseValueAt(data, pos, n.end)
		expandedValue, valueChanged := expandNode(valueNode)
		result.Write(expandedValue)

		if valueChanged {
			changed = true
		}

		pos = valueNode.end

		// 跳过逗号
		for pos < n.end && data[pos] <= ' ' {
			pos++
		}
		if pos < n.end && data[pos] == ',' {
			pos++
		}
	}

	result.WriteByte('}')
	return []byte(result.String()), changed
}

// expandArray 展开数组
func expandArray(n Node, data []byte) ([]byte, bool) {
	var result strings.Builder
	result.WriteByte('[')

	pos := n.start + 1 // skip '['
	changed := false
	first := true

	for pos < n.end {
		// 跳过空白
		for pos < n.end && data[pos] <= ' ' {
			pos++
		}
		if pos >= n.end || data[pos] == ']' {
			break
		}

		if !first {
			result.WriteByte(',')
		}
		first = false

		// 解析值
		valueNode := parseValueAt(data, pos, n.end)
		expandedValue, valueChanged := expandNode(valueNode)
		result.Write(expandedValue)

		if valueChanged {
			changed = true
		}

		pos = valueNode.end

		// 跳过逗号
		for pos < n.end && data[pos] <= ' ' {
			pos++
		}
		if pos < n.end && data[pos] == ',' {
			pos++
		}
	}

	result.WriteByte(']')
	return []byte(result.String()), changed
}

// expandString 展开字符串（如果包含嵌套JSON）
func expandString(n Node, data []byte) ([]byte, bool) {
	if n.start+1 >= n.end {
		return data[n.start:n.end], false
	}

	// 提取字符串内容（不包括引号）
	strContent := string(data[n.start+1 : n.end-1])

	// 解转义
	unescaped := unescapeJSON(strContent)

	// 检查是否为有效的JSON
	if isValidJSON(unescaped) {
		// 递归展开嵌套的JSON
		nestedExpanded := expandNestedJSON([]byte(unescaped))
		return nestedExpanded, true
	}

	return data[n.start:n.end], false
}

// parseRootNode 解析根节点
func parseRootNode(data []byte) Node {
	if len(data) == 0 {
		return Node{}
	}

	start, end := 0, len(data)
	for start < end && data[start] <= ' ' {
		start++
	}
	if start >= end {
		return Node{}
	}

	var typ byte
	switch data[start] {
	case '{':
		typ = 'o'
	case '[':
		typ = 'a'
	case '"':
		typ = 's'
	case 't', 'f':
		typ = 'b'
	case 'n':
		typ = 'l'
	default:
		typ = 'n'
	}
	end = skipValueFast(data, start, end)
	return Node{raw: data, start: start, end: end, typ: typ}
}

// ===== From / 基本访问 =====

// FromBytes 创建节点并智能展开嵌套的转义JSON
func FromBytes(b []byte) Node {
	if len(b) == 0 {
		return Node{}
	}

	// 首先创建原始节点
	originalNode := parseRootNode(b)
	if !originalNode.Exists() {
		return originalNode
	}

	// 尝试展开嵌套的JSON
	expanded := expandNestedJSON(b)

	// 如果展开后有变化，重新解析
	if len(expanded) != len(b) || string(expanded) != string(b) {
		expandedNode := parseRootNode(expanded)
		expandedNode.expanded = expanded
		return expandedNode
	}

	return originalNode
}

func (n Node) Get(path string) Node {
	if len(path) == 0 || len(n.raw) == 0 {
		return Node{}
	}
	for i := 0; i < len(path); i++ {
		if path[i] == '.' || path[i] == '[' {
			return n.GetPath(path)
		}
	}
	if n.typ != 'o' {
		return Node{}
	}

	data := n.getWorkingData()
	keyData := unsafe.StringData(path)
	keyLen := len(path)
	pos := findObjectField(data, n.start+1, n.end, keyData, 0, keyLen)
	if pos < 0 {
		return Node{}
	}
	return parseValueAtWithData(data, pos, n.end, n.expanded)
}

func (n Node) GetPath(path string) Node {
	if len(n.raw) == 0 || len(path) == 0 {
		return Node{}
	}
	data := n.getWorkingData()
	pos := n.start
	end := n.end

	for pos < end && (data[pos] == ' ' || data[pos] == '\t' || data[pos] == '\n' || data[pos] == '\r') {
		pos++
	}
	if pos < end && data[pos] == '{' {
		pos++
	}

	pathData := unsafe.StringData(path)
	pathLen := len(path)
	pathPos := 0

	for pathPos < pathLen {
		segStart := pathPos
		segLen := 0
		for pathPos < pathLen {
			c := *(*byte)(unsafe.Add(unsafe.Pointer(pathData), pathPos))
			if c == '.' || c == '[' {
				break
			}
			segLen++
			pathPos++
		}

		if segLen > 0 {
			pos = findObjectField(data, pos, end, pathData, segStart, segLen)
			if pos < 0 {
				return Node{}
			}
		}

		for pathPos < pathLen && *(*byte)(unsafe.Add(unsafe.Pointer(pathData), pathPos)) == '[' {
			pathPos++
			idx := 0
			for pathPos < pathLen {
				c := *(*byte)(unsafe.Add(unsafe.Pointer(pathData), pathPos))
				if c == ']' {
					pathPos++
					break
				}
				idx = idx*10 + int(c-'0')
				pathPos++
			}
			pos = findArrayElement(data, pos, end, idx)
			if pos < 0 {
				return Node{}
			}
		}

		if pathPos < pathLen && *(*byte)(unsafe.Add(unsafe.Pointer(pathData), pathPos)) == '.' {
			pathPos++
			if pos < end && data[pos] == '{' {
				pos++
			}
		}
	}

	return parseValueAtWithData(data, pos, end, n.expanded)
}

// parseValueAtWithData 解析指定位置的值，保持expanded数据
func parseValueAtWithData(data []byte, pos int, end int, expanded []byte) Node {
	node := parseValueAt(data, pos, end)
	if len(expanded) > 0 {
		node.expanded = expanded
	}
	return node
}

// ===== 对象/数组定位 =====
func findObjectField(data []byte, start int, end int, keyData *byte, keyStart int, keyLen int) int {
	pos := start
	for pos < end {
		for pos < end && (data[pos] <= ' ') {
			pos++
		}
		if pos >= end || data[pos] == '}' {
			return -1
		}
		if data[pos] != '"' {
			return -1
		}
		pos++
		fieldStart := pos
		match := true
		if pos+keyLen <= end && data[pos+keyLen] == '"' {
			for i := 0; i < keyLen; i++ {
				if data[fieldStart+i] != *(*byte)(unsafe.Add(unsafe.Pointer(keyData), keyStart+i)) {
					match = false
					break
				}
			}
			if match {
				pos += keyLen + 1
				for pos < end && data[pos] <= ' ' {
					pos++
				}
				if pos >= end || data[pos] != ':' {
					return -1
				}
				pos++
				for pos < end && data[pos] <= ' ' {
					pos++
				}
				return pos
			}
		}
		for pos < end && data[pos] != '"' {
			if data[pos] == '\\' {
				pos++
			}
			pos++
		}
		pos++
		for pos < end && data[pos] != ':' {
			pos++
		}
		pos++
		for pos < end && data[pos] <= ' ' {
			pos++
		}
		pos = skipValueFast(data, pos, end)
		if pos < end && data[pos] == ',' {
			pos++
		}
	}
	return -1
}

func findArrayElement(data []byte, start int, end int, index int) int {
	pos := start
	for pos < end && data[pos] <= ' ' {
		pos++
	}
	if pos >= end || data[pos] != '[' {
		return -1
	}
	pos++
	currentIndex := 0
	for pos < end {
		for pos < end && data[pos] <= ' ' {
			pos++
		}
		if pos >= end || data[pos] == ']' {
			return -1
		}
		if currentIndex == index {
			return pos
		}
		pos = skipValueFast(data, pos, end)
		currentIndex++
		for pos < end && data[pos] <= ' ' {
			pos++
		}
		if pos < end && data[pos] == ',' {
			pos++
		}
	}
	return -1
}

// Index 借助全局缓存，O(1) 取第 i 个元素起点；保持值接收器以支持链式
func (n Node) Index(i int) Node {
	offs := buildArrOffsetsCached(n)
	if i < 0 || i >= len(offs) {
		return Node{}
	}
	data := n.getWorkingData()
	pos := offs[i]
	end := skipValueFast(data, pos, n.end)
	node := Node{raw: n.raw, start: pos, end: end, typ: detectType(data[pos])}
	if len(n.expanded) > 0 {
		node.expanded = n.expanded
	}
	return node
}

// ===== 跳值 / 解析 =====
// 替换原函数：更稳健的越界处理与转义跳过
func skipValueFast(data []byte, pos int, end int) int {
	if pos >= end {
		return pos
	}
	switch data[pos] {
	case '"':
		pos++
		for pos < end {
			switch data[pos] {
			case '"':
				return pos + 1
			case '\\':
				// 处理转义，确保不越界；对 \uXXXX 做快速跳过
				if pos+1 >= end {
					return end
				}
				if data[pos+1] == 'u' && pos+5 < end {
					pos += 6 // \uXXXX
				} else {
					pos += 2 // \x
				}
			default:
				pos++
			}
		}
		return end

	case '{':
		pos++
		depth := 1
		for pos < end && depth > 0 {
			switch data[pos] {
			case '"':
				// 跳过字符串（含转义与 \uXXXX）
				pos++
				for pos < end {
					switch data[pos] {
					case '"':
						pos++
						goto contObj // 结束当前字符串，继续对象扫描
					case '\\':
						if pos+1 >= end {
							return end
						}
						if data[pos+1] == 'u' && pos+5 < end {
							pos += 6
						} else {
							pos += 2
						}
					default:
						pos++
					}
				}
				return end
			case '{':
				depth++
				pos++
			case '}':
				depth--
				pos++
			default:
				pos++
			}
		}
		return pos
	contObj:
		// 继续对象扫描
		for pos < end && depth > 0 {
			switch data[pos] {
			case '"':
				pos++
				for pos < end {
					switch data[pos] {
					case '"':
						pos++
						goto contObj
					case '\\':
						if pos+1 >= end {
							return end
						}
						if data[pos+1] == 'u' && pos+5 < end {
							pos += 6
						} else {
							pos += 2
						}
					default:
						pos++
					}
				}
				return end
			case '{':
				depth++
				pos++
			case '}':
				depth--
				pos++
			default:
				pos++
			}
		}
		return pos

	case '[':
		pos++
		depth := 1
		for pos < end && depth > 0 {
			switch data[pos] {
			case '"':
				// 跳过字符串（含转义与 \uXXXX）
				pos++
				for pos < end {
					switch data[pos] {
					case '"':
						pos++
						goto contArr
					case '\\':
						if pos+1 >= end {
							return end
						}
						if data[pos+1] == 'u' && pos+5 < end {
							pos += 6
						} else {
							pos += 2
						}
					default:
						pos++
					}
				}
				return end
			case '[':
				depth++
				pos++
			case ']':
				depth--
				pos++
			default:
				pos++
			}
		}
		return pos
	contArr:
		for pos < end && depth > 0 {
			switch data[pos] {
			case '"':
				pos++
				for pos < end {
					switch data[pos] {
					case '"':
						pos++
						goto contArr
					case '\\':
						if pos+1 >= end {
							return end
						}
						if data[pos+1] == 'u' && pos+5 < end {
							pos += 6
						} else {
							pos += 2
						}
					default:
						pos++
					}
				}
				return end
			case '[':
				depth++
				pos++
			case ']':
				depth--
				pos++
			default:
				pos++
			}
		}
		return pos

	case 't':
		if pos+4 <= end {
			return pos + 4
		}
		return end
	case 'f':
		if pos+5 <= end {
			return pos + 5
		}
		return end
	case 'n':
		if pos+4 <= end {
			return pos + 4
		}
		return end
	default:
		// number: [-] digits [ . digits ] [ e[+/-]digits ]
		if data[pos] == '-' {
			pos++
			if pos >= end {
				return end
			}
		}
		for pos < end && data[pos] >= '0' && data[pos] <= '9' {
			pos++
		}
		if pos < end && data[pos] == '.' {
			pos++
			for pos < end && data[pos] >= '0' && data[pos] <= '9' {
				pos++
			}
		}
		if pos < end && (data[pos] == 'e' || data[pos] == 'E') {
			pos++
			if pos < end && (data[pos] == '+' || data[pos] == '-') {
				pos++
			}
			for pos < end && data[pos] >= '0' && data[pos] <= '9' {
				pos++
			}
		}
		return pos
	}
}

func parseValueAt(data []byte, pos int, end int) Node {
	if pos < 0 || pos >= end {
		return Node{}
	}
	var typ byte
	valStart := pos
	valEnd := pos
	switch data[pos] {
	case '"':
		typ = 's'
		valEnd = pos + 1
		for valEnd < end && data[valEnd] != '"' {
			if data[valEnd] == '\\' {
				valEnd++
			}
			valEnd++
		}
		if valEnd < end {
			valEnd++
		}
	case '{':
		typ = 'o'
		valEnd = skipValueFast(data, pos, end)
	case '[':
		typ = 'a'
		valEnd = skipValueFast(data, pos, end)
	case 't':
		typ = 'b'
		valEnd = pos + 4
	case 'f':
		typ = 'b'
		valEnd = pos + 5
	case 'n':
		typ = 'l'
		valEnd = pos + 4
	default:
		if data[pos] == '-' || (data[pos] >= '0' && data[pos] <= '9') {
			typ = 'n'
			valEnd = skipValueFast(data, pos, end)
		}
	}
	return Node{raw: data, start: valStart, end: valEnd, typ: typ}
}

// ===== 字面量取值 =====

// String 返回节点的字符串值
// 如果节点类型不是 JSON 字符串，或内容为空，则返回错误
func (n Node) String() (string, error) {
	if n.typ != 's' {
		return "", fmt.Errorf("node is not a string type (got type=%q)", n.Kind())
	}
	data := n.getWorkingData()
	if n.start+1 >= n.end {
		return "", fmt.Errorf("invalid string bounds: start=%d end=%d", n.start, n.end)
	}

	bytes := data[n.start+1 : n.end-1]
	if len(bytes) == 0 {
		return "", nil // 空字符串正常返回
	}

	return unsafe.String(&bytes[0], len(bytes)), nil
}

// Int 返回节点的 int64 整数值
// 如果节点类型不是 JSON 数字、为空、包含非整数字符，或超出 int64 范围，则返回错误
func (n Node) Int() (int64, error) {
	if n.typ != 'n' || n.start >= n.end {
		return 0, fmt.Errorf("node is not a number type (got type=%q)", n.Kind())
	}
	data := n.getWorkingData()[n.start:n.end]
	if len(data) == 0 {
		return 0, fmt.Errorf("empty number at [%d:%d] (type=%q)", n.start, n.end, n.Kind())
	}
	i := 0
	neg := false
	if data[0] == '-' {
		neg = true
		i++
		if i >= len(data) {
			return 0, fmt.Errorf("invalid number: only a minus sign at [%d:%d] (type=%q)", n.start, n.end, n.Kind())
		}
	}
	var val uint64
	for ; i < len(data); i++ {
		c := data[i]
		if c < '0' || c > '9' {
			return 0, fmt.Errorf(
				"not an integer: found char %q at pos=%d (abs=%d, type=%q)",
				c, i, n.start+i, n.Kind(),
			)
		}
		d := uint64(c - '0')
		if !neg {
			if val > (maxInt64U-d)/10 {
				return 0, fmt.Errorf(
					"int64 overflow: positive value exceeds %d at pos=%d (abs=%d, current=%d)",
					math.MaxInt64, i, n.start+i, val,
				)
			}
		} else {
			if val > (minInt64U-d)/10 {
				return 0, fmt.Errorf(
					"int64 overflow: negative value exceeds %d at pos=%d (abs=%d, current=%d)",
					math.MinInt64, i, n.start+i, val,
				)
			}
		}
		val = val*10 + d
	}
	if neg {
		return -int64(val), nil
	}
	return int64(val), nil
}

// 其他方法保持原有实现...
// [省略大量重复代码，这里只展示关键修改部分]

// ===== Predicates =====

// Exists 判断节点是否存在。
// 若原始数据非空且起止位置有效，则返回 true，否则返回 false
func (n Node) Exists() bool {
	return (len(n.raw) > 0 || len(n.expanded) > 0) && n.start >= 0 && n.end > n.start && n.typ != 0
}

// IsObject 判断节点是否为 JSON 对象
func (n Node) IsObject() bool { return n.typ == 'o' }

// IsArray 判断节点是否为 JSON 数组
func (n Node) IsArray() bool { return n.typ == 'a' }

// IsString 判断节点是否为 JSON 字符串
func (n Node) IsString() bool { return n.typ == 's' }

// IsNumber 判断节点是否为 JSON 数字
func (n Node) IsNumber() bool { return n.typ == 'n' }

// IsBool 判断节点是否为 JSON 布尔值
func (n Node) IsBool() bool { return n.typ == 'b' }

// IsNull 判断节点是否为 JSON null
func (n Node) IsNull() bool { return n.typ == 'l' }

// IsScalar 判断节点是否为标量类型（字符串、数字、布尔值或 null）
func (n Node) IsScalar() bool {
	return n.typ == 's' || n.typ == 'n' || n.typ == 'b' || n.typ == 'l'
}

// IsContainer 判断节点是否为容器类型（对象或数组）
func (n Node) IsContainer() bool {
	return n.typ == 'o' || n.typ == 'a'
}

// ===== 其他方法的实现 =====

// Uint 返回节点的 uint64 无符号整数值
func (n Node) Uint() (uint64, error) {
	if n.typ != 'n' || n.start >= n.end {
		return 0, fmt.Errorf("not a number: got type=%q at range [%d:%d]", n.Kind(), n.start, n.end)
	}
	data := n.getWorkingData()[n.start:n.end]
	if len(data) == 0 {
		return 0, fmt.Errorf("empty number at range [%d:%d] (type=%q)", n.start, n.end, n.Kind())
	}
	if data[0] == '-' {
		return 0, fmt.Errorf("negative to uint at pos=%d (abs=%d, type=%q)", 0, n.start, n.Kind())
	}
	var val uint64
	for i := 0; i < len(data); i++ {
		c := data[i]
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not an unsigned integer: found char %q at pos=%d (abs=%d, type=%q)", c, i, n.start+i, n.Kind())
		}
		d := uint64(c - '0')
		if val > (maxUint64-d)/10 {
			return 0, fmt.Errorf("uint64 overflow: value exceeds %d at pos=%d (abs=%d, current=%d)", maxUint64, i, n.start+i, val)
		}
		val = val*10 + d
	}
	return val, nil
}

// Float 返回节点的 float64 浮点值
func (n Node) Float() (float64, error) {
	if n.typ != 'n' || n.start >= n.end {
		return 0, fmt.Errorf("not a number: got type=%q at range [%d:%d] (len=%d)", n.Kind(), n.start, n.end, n.end-n.start)
	}
	data := n.getWorkingData()[n.start:n.end]
	if len(data) == 0 {
		return 0, fmt.Errorf("empty number at range [%d:%d] (type=%q)", n.start, n.end, n.Kind())
	}
	i := 0
	neg := false
	if data[i] == '-' {
		neg = true
		i++
		if i >= len(data) {
			return 0, fmt.Errorf("invalid number: lone '-' at pos=%d (abs=%d, type=%q)", i-1, n.start+(i-1), n.Kind())
		}
	}
	var mant uint64
	sawDigit := false
	const maxMantDigits = 19
	digits := 0
	for i < len(data) {
		c := data[i]
		if c < '0' || c > '9' {
			break
		}
		sawDigit = true
		if digits < maxMantDigits {
			mant = mant*10 + uint64(c-'0')
			digits++
		}
		i++
	}
	decExp := 0
	if i < len(data) && data[i] == '.' {
		i++
		for i < len(data) {
			c := data[i]
			if c < '0' || c > '9' {
				break
			}
			sawDigit = true
			if digits < maxMantDigits {
				mant = mant*10 + uint64(c-'0')
				digits++
				decExp--
			} else {
				decExp--
			}
			i++
		}
	}
	if i < len(data) && (data[i] == 'e' || data[i] == 'E') {
		i++
		if i >= len(data) {
			return 0, fmt.Errorf("invalid number: unexpected end after '-' at pos=%d (abs=%d, type=%q)", i-1, n.start+(i-1), n.Kind())
		}
		expNeg := false
		if data[i] == '+' || data[i] == '-' {
			expNeg = data[i] == '-'
			i++
		}
		if i >= len(data) || data[i] < '0' || data[i] > '9' {
			var got byte
			if i < len(data) {
				got = data[i]
			}
			return 0, fmt.Errorf("invalid number: expected digit but got %q at pos=%d (abs=%d, type=%q)", got, i, n.start+i, n.Kind())
		}
		exp := 0
		const maxExp = 1000
		for i < len(data) {
			c := data[i]
			if c < '0' || c > '9' {
				break
			}
			if exp < maxExp {
				exp = exp*10 + int(c-'0')
			}
			i++
		}
		if expNeg {
			decExp -= exp
		} else {
			decExp += exp
		}
	}
	if !sawDigit {
		return 0, fmt.Errorf("invalid number: no digits found at range [%d:%d] (type=%q)", n.start, n.end, n.Kind())
	}
	f := float64(mant)
	if decExp != 0 {
		f = scaleByPow10(f, decExp)
	}
	if neg {
		f = -f
	}
	return f, nil
}

func scaleByPow10(x float64, k int) float64 {
	if x == 0 {
		return 0
	}
	var p = [...]float64{1e1, 1e2, 1e4, 1e8, 1e16, 1e32, 1e64, 1e128, 1e256}
	kk := k
	if kk < 0 {
		kk = -kk
	}
	if kk > 350 {
		if k > 0 {
			return x * 1.0e308 * 1.0e308
		}
		return 0
	}
	if k > 0 {
		i := 0
		for kk != 0 {
			if kk&1 == 1 {
				x *= p[i]
			}
			kk >>= 1
			i++
		}
	} else if k < 0 {
		i := 0
		for kk != 0 {
			if kk&1 == 1 {
				x /= p[i]
			}
			kk >>= 1
			i++
		}
	}
	return x
}

// Bool 返回节点的布尔值
func (n Node) Bool() (bool, error) {
	if n.typ != 'b' || n.start >= n.end {
		return false, fmt.Errorf("not a bool: got type=%q at range [%d:%d]", n.Kind(), n.start, n.end)
	}
	data := n.getWorkingData()[n.start:n.end]
	if len(data) == 4 && data[0] == 't' && data[1] == 'r' && data[2] == 'u' && data[3] == 'e' {
		return true, nil
	}
	if len(data) == 5 && data[0] == 'f' && data[1] == 'a' && data[2] == 'l' && data[3] == 's' && data[4] == 'e' {
		return false, nil
	}
	return false, fmt.Errorf("invalid bool: value=%q at range [%d:%d] (type=%q)",
		unsafe.String(&data[0], len(data)), n.start, n.end, n.Kind())
}

// NumStr 返回节点的数字原始字符串表示
func (n Node) NumStr() (string, error) {
	if n.typ != 'n' || n.start >= n.end {
		return "", fmt.Errorf("not a number: got type=%q at range [%d:%d]", n.Kind(), n.start, n.end)
	}
	data := n.getWorkingData()
	return unsafe.String(&data[n.start], n.end-n.start), nil
}

// Raw 返回节点的原始 JSON 字节切片
func (n Node) Raw() []byte {
	data := n.getWorkingData()
	if n.start >= 0 && n.end <= len(data) && n.start < n.end {
		return data[n.start:n.end]
	}
	return nil
}

// Json 返回节点的 JSON 表示（仅 object 和 array 可用）
func (n Node) Json() (string, error) {
	if !n.Exists() || n.start < 0 || n.start >= n.end {
		return "", fmt.Errorf("invalid node: exists=%v, type=%q, range=[%d:%d]", n.Exists(), n.Kind(), n.start, n.end)
	}
	// 类型安全
	if n.typ != 'o' && n.typ != 'a' {
		return "", fmt.Errorf("json() only valid for object/array, got type=%q at range [%d:%d]", n.Kind(), n.start, n.end)
	}
	data := n.getWorkingData()
	if n.end > len(data) {
		return "", fmt.Errorf("invalid range: end=%d > len(data)=%d", n.end, len(data))
	}
	return unsafe.String(&data[n.start], n.end-n.start), nil
}

// ===== 统计 / Keys =====

func (n Node) Len() int {
	data := n.getWorkingData()
	// 数组
	if n.typ == 'a' {
		pos := n.start
		end := n.end
		for pos < end && data[pos] != '[' {
			pos++
		}
		if pos >= end {
			return 0
		}
		pos++
		count := 0
		for pos < end {
			for pos < end && data[pos] <= ' ' {
				pos++
			}
			if pos >= end || data[pos] == ']' {
				break
			}
			count++
			pos = skipValueFast(data, pos, end)
			for pos < end && data[pos] <= ' ' {
				pos++
			}
			if pos < end && data[pos] == ',' {
				pos++
			}
		}
		return count
	}
	// 对象
	if n.typ == 'o' {
		pos := n.start
		end := n.end
		for pos < end && data[pos] != '{' {
			pos++
		}
		if pos >= end {
			return 0
		}
		pos++
		count := 0
		for pos < end {
			for pos < end && data[pos] <= ' ' {
				pos++
			}
			if pos >= end || data[pos] == '}' {
				break
			}
			if data[pos] != '"' {
				return count
			}
			pos++
			for pos < end && data[pos] != '"' {
				if data[pos] == '\\' {
					pos++
				}
				pos++
			}
			pos++
			for pos < end && data[pos] <= ' ' {
				pos++
			}
			if pos >= end || data[pos] != ':' {
				return count
			}
			pos++
			for pos < end && data[pos] <= ' ' {
				pos++
			}
			pos = skipValueFast(data, pos, end)
			count++
			for pos < end && data[pos] <= ' ' {
				pos++
			}
			if pos < end && data[pos] == ',' {
				pos++
			}
		}
		return count
	}
	// 字符串
	if n.typ == 's' {
		start := n.start
		end := n.end
		// 找到第一个引号
		for start < end && data[start] != '"' {
			start++
		}
		if start >= end {
			return 0
		}
		start++
		length := 0
		for start < end && data[start] != '"' {
			if data[start] == '\\' {
				start++ // 跳过转义符
			}
			start++
			length++
		}
		return length
	}
	return 0
}

func (n Node) Keys() [][]byte {
	if n.typ != 'o' {
		return nil
	}
	var keys [][]byte
	data := n.getWorkingData()
	pos := n.start
	end := n.end
	for pos < end && data[pos] != '{' {
		pos++
	}
	if pos >= end {
		return nil
	}
	pos++
	for pos < end {
		for pos < end && data[pos] <= ' ' {
			pos++
		}
		if pos >= end || data[pos] == '}' {
			break
		}
		if data[pos] != '"' {
			return keys
		}
		pos++
		keyStart := pos
		for pos < end && data[pos] != '"' {
			if data[pos] == '\\' {
				pos++
			}
			pos++
		}
		keyEnd := pos
		keys = append(keys, data[keyStart:keyEnd])
		pos++
		for pos < end && data[pos] <= ' ' {
			pos++
		}
		if pos >= end || data[pos] != ':' {
			return keys
		}
		pos++
		for pos < end && data[pos] <= ' ' {
			pos++
		}
		pos = skipValueFast(data, pos, end)
		for pos < end && data[pos] <= ' ' {
			pos++
		}
		if pos < end && data[pos] == ',' {
			pos++
		}
	}
	return keys
}

// ===== 解码 =====

// RawString 返回节点的原始 JSON 字符串形式。
func (n Node) RawString() (string, error) {
	data := n.getWorkingData()
	if n.start >= 0 && n.end <= len(data) && n.start < n.end {
		return unsafe.String(&data[n.start], n.end-n.start), nil
	}
	return "", fmt.Errorf("invalid node range: start=%d, end=%d, len(data)=%d, type=%q", n.start, n.end, len(data), n.Kind())
}

// Decode 将节点的 JSON 值解码到提供的变量 v 中
func (n Node) Decode(v any) error {
	if !n.Exists() {
		return fmt.Errorf("node does not exist: start=%d, end=%d, type=%q", n.start, n.end, n.Kind())
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer: got kind=%s, isNil=%v, type=%T", rv.Kind(), rv.IsNil(), v)
	}
	data := n.getWorkingData()
	val, _, err := fastDecode(data, n.start, n.end)
	if err != nil {
		return err
	}
	rv.Elem().Set(reflect.ValueOf(val))
	return nil
}

func fastDecode(buf []byte, start, end int) (any, int, error) {
	if start >= end {
		return nil, start, fmt.Errorf("empty node: start=%d, end=%d, len(buf)=%d", start, end, len(buf))
	}
	switch buf[start] {
	case '{':
		m := make(map[string]any)
		i := start + 1
		for {
			for i < end && (buf[i] <= ' ' || buf[i] == ',') {
				i++
			}
			if i >= end || buf[i] == '}' {
				return m, i + 1, nil
			}
			if buf[i] != '"' {
				return nil, i, fmt.Errorf("invalid object key: expected '\"' at pos=%d, got=%q (byte=%d)", i, buf[i], buf[i])
			}
			keyStart := i + 1
			i++
			for i < end && buf[i] != '"' {
				i++
			}
			key := unsafe.String(&buf[keyStart], i-keyStart)
			i++
			for i < end && (buf[i] <= ' ' || buf[i] == ':') {
				i++
			}
			val, ni, err := fastDecode(buf, i, end)
			if err != nil {
				return nil, ni, err
			}
			m[key] = val
			i = ni
		}
	case '[':
		arr := make([]any, 0)
		i := start + 1
		for {
			for i < end && (buf[i] <= ' ' || buf[i] == ',') {
				i++
			}
			if i >= end || buf[i] == ']' {
				return arr, i + 1, nil
			}
			val, ni, err := fastDecode(buf, i, end)
			if err != nil {
				return nil, ni, err
			}
			arr = append(arr, val)
			i = ni
		}
	case '"':
		str := unsafe.String(&buf[start+1], end-start-2)
		return str, end, nil
	case 't':
		return true, start + 4, nil
	case 'f':
		return false, start + 5, nil
	case 'n':
		return nil, start + 4, nil
	default:
		numEnd := start
		for numEnd < end && (buf[numEnd] == '.' || buf[numEnd] == '-' || buf[numEnd] == '+' ||
			(buf[numEnd] >= '0' && buf[numEnd] <= '9') || buf[numEnd] == 'e' || buf[numEnd] == 'E') {
			numEnd++
		}
		if iv, err := parseIntFast(buf[start:numEnd]); err == nil {
			return iv, numEnd, nil
		}
		fv := parseFloatFast(buf[start:numEnd])
		return fv, numEnd, nil
	}
}

func parseIntFast(data []byte) (int64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty number: len(data)=%d, data=%q", len(data), data)
	}
	i := 0
	neg := false
	if data[0] == '-' {
		neg = true
		i++
		if i >= len(data) {
			return 0, fmt.Errorf("invalid number: only '-' found, no digits after, len(data)=%d, data=%q", len(data), data)
		}
	}
	var val uint64
	for ; i < len(data); i++ {
		c := data[i]
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not an integer: found %q (byte=%d) at pos=%d in %q", c, c, i, data)
		}
		d := uint64(c - '0')
		if !neg {
			if val > (maxInt64U-d)/10 {
				return 0, fmt.Errorf("int64 overflow: value=%d digit=%d pos=%d data=%q", val, d, i, data)
			}
		} else {
			if val > (minInt64U-d)/10 {
				return 0, fmt.Errorf("int64 overflow: negative value=%d digit=%d pos=%d data=%q", val, d, i, data)
			}
		}
		val = val*10 + d
	}
	if neg {
		return -int64(val), nil
	}
	return int64(val), nil
}

func parseFloatFast(data []byte) float64 {
	i := 0
	neg := false
	if len(data) > 0 && data[i] == '-' {
		neg = true
		i++
	}
	var mant uint64
	sawDigit := false
	const maxMantDigits = 19
	digits := 0
	for i < len(data) {
		c := data[i]
		if c < '0' || c > '9' {
			break
		}
		sawDigit = true
		if digits < maxMantDigits {
			mant = mant*10 + uint64(c-'0')
			digits++
		}
		i++
	}
	decExp := 0
	if i < len(data) && data[i] == '.' {
		i++
		for i < len(data) {
			c := data[i]
			if c < '0' || c > '9' {
				break
			}
			sawDigit = true
			if digits < maxMantDigits {
				mant = mant*10 + uint64(c-'0')
				digits++
				decExp--
			} else {
				decExp--
			}
			i++
		}
	}
	if i < len(data) && (data[i] == 'e' || data[i] == 'E') {
		i++
		expNeg := false
		if i < len(data) && (data[i] == '+' || data[i] == '-') {
			expNeg = data[i] == '-'
			i++
		}
		exp := 0
		for i < len(data) {
			c := data[i]
			if c < '0' || c > '9' {
				break
			}
			exp = exp*10 + int(c-'0')
			i++
		}
		if expNeg {
			decExp -= exp
		} else {
			decExp += exp
		}
	}
	if !sawDigit {
		return 0
	}
	f := float64(mant)
	if decExp != 0 {
		f = scaleByPow10(f, decExp)
	}
	if neg {
		f = -f
	}
	return f
}

// ===== TypeInfo =====

func (n Node) Type() byte { return n.typ }

func (n Node) Kind() NodeType { return NodeType(n.typ) }

func (t NodeType) String() string {
	switch t {
	case TypeObject:
		return "object"
	case TypeArray:
		return "array"
	case TypeString:
		return "string"
	case TypeNumber:
		return "number"
	case TypeBool:
		return "bool"
	case TypeNull:
		return "null"
	default:
		return "invalid"
	}
}

func detectType(c byte) byte {
	switch c {
	case '{':
		return 'o'
	case '[':
		return 'a'
	case '"':
		return 's'
	case 't', 'f':
		return 'b'
	case 'n':
		return 'l'
	default:
		return 'n'
	}
}
