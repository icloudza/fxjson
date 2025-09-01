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

// ParseOptions 用于控制 JSON 解析行为和安全限制
type ParseOptions struct {
	MaxDepth      int  // 最大嵌套深度，0 表示无限制
	MaxStringLen  int  // 最大字符串长度，0 表示无限制
	MaxObjectKeys int  // 最大对象键数量，0 表示无限制
	MaxArrayItems int  // 最大数组项数量，0 表示无限制
	StrictMode    bool // 严格模式：拒绝格式错误的 JSON
}

// DefaultParseOptions 默认解析选项
var DefaultParseOptions = ParseOptions{
	MaxDepth:      1000,        // 默认最大1000层嵌套
	MaxStringLen:  1024 * 1024, // 默认最大1MB字符串
	MaxObjectKeys: 10000,       // 默认最大10000个键
	MaxArrayItems: 100000,      // 默认最大100000个数组项
	StrictMode:    false,       // 默认非严格模式
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
	// 只有以 {, [, " 开头的才可能是嵌套的 JSON
	// 数字、布尔值和 null 不应该被展开
	firstChar := s[0]
	if firstChar != '{' && firstChar != '[' && firstChar != '"' {
		return false
	}
	// 使用简化的验证，避免循环依赖
	return isValidJSONSimple([]byte(s))
}

// isValidJSONSimple 简单的JSON格式验证
func isValidJSONSimple(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	start, end := 0, len(data)
	for start < end && data[start] <= ' ' {
		start++
	}
	if start >= end {
		return false
	}

	// 使用简化的skipValue来检查
	valueEnd := skipValueSimple(data, start, end)

	// 检查是否读取了整个输入
	pos := valueEnd
	for pos < end && data[pos] <= ' ' {
		pos++
	}
	return pos == end && valueEnd > start
}

// skipValueSimple 简化的值跳过函数，不会调用其他可能导致循环的函数
func skipValueSimple(data []byte, pos int, end int) int {
	if pos >= end {
		return pos
	}

	switch data[pos] {
	case '{':
		pos++
		depth := 1
		for pos < end && depth > 0 {
			if data[pos] == '"' {
				pos = skipStringSimple(data, pos, end)
			} else if data[pos] == '{' {
				depth++
				pos++
			} else if data[pos] == '}' {
				depth--
				pos++
			} else {
				pos++
			}
		}
		return pos
	case '[':
		pos++
		depth := 1
		for pos < end && depth > 0 {
			if data[pos] == '"' {
				pos = skipStringSimple(data, pos, end)
			} else if data[pos] == '[' {
				depth++
				pos++
			} else if data[pos] == ']' {
				depth--
				pos++
			} else {
				pos++
			}
		}
		return pos
	case '"':
		return skipStringSimple(data, pos, end)
	case 't':
		if pos+4 <= end && string(data[pos:pos+4]) == "true" {
			return pos + 4
		}
		return pos
	case 'f':
		if pos+5 <= end && string(data[pos:pos+5]) == "false" {
			return pos + 5
		}
		return pos
	case 'n':
		if pos+4 <= end && string(data[pos:pos+4]) == "null" {
			return pos + 4
		}
		return pos
	default:
		if data[pos] == '-' || (data[pos] >= '0' && data[pos] <= '9') {
			// 跳过数字
			pos++
			for pos < end && ((data[pos] >= '0' && data[pos] <= '9') || data[pos] == '.' || data[pos] == 'e' || data[pos] == 'E' || data[pos] == '+' || data[pos] == '-') {
				pos++
			}
			return pos
		}
		return pos // 无效字符
	}
}

// skipStringSimple 简化的字符串跳过
func skipStringSimple(data []byte, pos int, end int) int {
	if pos >= end || data[pos] != '"' {
		return pos
	}
	pos++ // 跳过开始引号
	for pos < end {
		if data[pos] == '"' {
			return pos + 1 // 跳过结束引号
		}
		if data[pos] == '\\' && pos+1 < end {
			pos += 2 // 跳过转义字符
		} else {
			pos++
		}
	}
	return pos
}

// expandNestedJSON 迭代展开嵌套的JSON字符串，避免栈溢出
func expandNestedJSON(data []byte) []byte {
	node := parseRootNode(data)
	if !node.Exists() {
		return data
	}

	expanded, changed := expandNodeIterative(node)
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

// expandTaskType 表示展开任务的类型
type expandTaskType int

const (
	expandTaskExpand expandTaskType = iota // 展开任务
	expandTaskResult                       // 结果收集任务
)

// expandTask 展开任务结构
type expandTask struct {
	taskType expandTaskType
	node     Node
	result   *[]byte  // 用于存储结果
	changed  *bool    // 用于标记是否有变化
	parentID int      // 父任务ID，用于结果收集
}

// expandNodeIterative 使用迭代方式展开节点，避免栈溢出
func expandNodeIterative(rootNode Node) ([]byte, bool) {
	// 使用栈来管理展开任务
	stack := make([]expandTask, 0, 64) // 预分配容量避免频繁扩容
	
	// 结果存储
	var result []byte
	var changed bool
	
	// 推入根任务
	stack = append(stack, expandTask{
		taskType: expandTaskExpand,
		node:     rootNode,
		result:   &result,
		changed:  &changed,
	})
	
	for len(stack) > 0 {
		// 弹出任务
		task := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		
		switch task.taskType {
		case expandTaskExpand:
			data := task.node.getWorkingData()
			
			switch task.node.typ {
			case 'o':
				expandedObj, objChanged := expandObjectIterative(task.node, data)
				*task.result = expandedObj
				*task.changed = objChanged
				
			case 'a':
				expandedArr, arrChanged := expandArrayIterative(task.node, data)
				*task.result = expandedArr
				*task.changed = arrChanged
				
			case 's':
				expandedStr, strChanged := expandStringIterative(task.node, data)
				*task.result = expandedStr
				*task.changed = strChanged
				
			default:
				*task.result = data[task.node.start:task.node.end]
				*task.changed = false
			}
		}
	}
	
	return result, changed
}

// expandStringIterative 迭代展开字符串，避免栈溢出
func expandStringIterative(n Node, data []byte) ([]byte, bool) {
	if n.start+1 >= n.end {
		return data[n.start:n.end], false
	}

	// 提取字符串内容（不包括引号）
	strContent := string(data[n.start+1 : n.end-1])

	// 解转义
	unescaped := unescapeJSON(strContent)

	// 检查是否为有效的JSON
	if isValidJSON(unescaped) {
		// 使用迭代方式展开嵌套的JSON，避免递归调用expandNestedJSON
		nestedNode := parseRootNode([]byte(unescaped))
		if !nestedNode.Exists() {
			return data[n.start:n.end], false
		}
		
		// 直接调用迭代版本，避免递归
		nestedExpanded, _ := expandNodeIterative(nestedNode)
		return nestedExpanded, true
	}

	return data[n.start:n.end], false
}

// expandObjectIterative 迭代展开对象
func expandObjectIterative(n Node, data []byte) ([]byte, bool) {
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
		
		// 使用迭代方式展开值
		expandedValue, valueChanged := expandNodeIterative(valueNode)
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

// expandArrayIterative 迭代展开数组
func expandArrayIterative(n Node, data []byte) ([]byte, bool) {
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
		
		// 使用迭代方式展开值
		expandedValue, valueChanged := expandNodeIterative(valueNode)
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
		if data[start] == '-' || (data[start] >= '0' && data[start] <= '9') {
			typ = 'n'
		} else {
			return Node{} // 无效的开始字符
		}
	}

	valueEnd := skipValueFast(data, start, end)

	// 验证JSON是否完整
	if valueEnd == start {
		return Node{} // skipValueFast没有前进，说明格式错误
	}

	// 对于对象和数组，需要特别检查是否真正完整
	if typ == 'o' {
		if valueEnd > end || (valueEnd > 0 && data[valueEnd-1] != '}') {
			return Node{} // 对象不完整
		}
	}
	if typ == 'a' {
		if valueEnd > end || (valueEnd > 0 && data[valueEnd-1] != ']') {
			return Node{} // 数组不完整
		}
	}

	// 检查是否有多余的字符（除了空白）
	pos := valueEnd
	for pos < end && data[pos] <= ' ' {
		pos++
	}
	if pos < end {
		return Node{} // 有多余的非空白字符
	}

	return Node{raw: data, start: start, end: valueEnd, typ: typ}
}

// ===== From / 基本访问 =====

// FromBytes 创建节点并智能展开嵌套的转义JSON
func FromBytes(b []byte) Node {
	return FromBytesWithOptions(b, DefaultParseOptions)
}

// FromBytesWithOptions 使用指定选项解析 JSON
func FromBytesWithOptions(b []byte, opts ParseOptions) Node {
	if len(b) == 0 {
		return Node{}
	}

	// 安全检查
	if err := validateJSON(b, opts); err != nil {
		return Node{typ: byte(TypeInvalid)}
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

// validateJSON 验证 JSON 数据的安全性
func validateJSON(data []byte, opts ParseOptions) error {
	if len(data) == 0 {
		return nil
	}

	depth := 0
	maxDepth := 0
	stringLen := 0
	objectKeys := 0
	arrayItems := 0
	inString := false
	escaped := false

	for i := 0; i < len(data); i++ {
		c := data[i]

		if inString {
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				inString = false
				// 检查字符串长度
				if opts.MaxStringLen > 0 && stringLen > opts.MaxStringLen {
					return fmt.Errorf("string too long: %d > %d", stringLen, opts.MaxStringLen)
				}
				stringLen = 0
			} else {
				stringLen++
			}
			continue
		}

		switch c {
		case '"':
			inString = true
			stringLen = 0
		case '{':
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
			if opts.MaxDepth > 0 && depth > opts.MaxDepth {
				return fmt.Errorf("nesting too deep: %d > %d", depth, opts.MaxDepth)
			}
			objectKeys = 0
		case '}':
			if depth <= 0 && opts.StrictMode {
				return fmt.Errorf("unexpected '}'")
			}
			depth--
		case '[':
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
			if opts.MaxDepth > 0 && depth > opts.MaxDepth {
				return fmt.Errorf("nesting too deep: %d > %d", depth, opts.MaxDepth)
			}
			arrayItems = 0
		case ']':
			if depth <= 0 && opts.StrictMode {
				return fmt.Errorf("unexpected ']'")
			}
			depth--
		case ':':
			if depth > 0 {
				objectKeys++
				if opts.MaxObjectKeys > 0 && objectKeys > opts.MaxObjectKeys {
					return fmt.Errorf("too many object keys: %d > %d", objectKeys, opts.MaxObjectKeys)
				}
			}
		case ',':
			if depth > 0 {
				arrayItems++
				if opts.MaxArrayItems > 0 && arrayItems > opts.MaxArrayItems {
					return fmt.Errorf("too many array items: %d > %d", arrayItems, opts.MaxArrayItems)
				}
			}
		}
	}

	if opts.StrictMode && depth != 0 {
		return fmt.Errorf("unmatched brackets, depth: %d", depth)
	}

	return nil
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
	// 安全检查：确保路径非空且数据有效
	if len(path) == 0 || len(data) == 0 {
		return Node{}
	}
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
				// 安全检查：确保字符是数字
				if c < '0' || c > '9' {
					return Node{} // 无效的数组索引格式
				}
				// 防止整数溢出
				if idx > (int(^uint(0)>>1)-int(c-'0'))/10 {
					return Node{} // 索引过大，防止溢出
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
			// 优化：使用更高效的字节比较
			if keyLen > 0 {
				fieldBytes := data[fieldStart : fieldStart+keyLen]
				keyBytes := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(keyData), keyStart)), keyLen)
				
				// 对于较长的键，使用8字节块比较
				if keyLen >= 8 {
					// 比较前8字节
					fieldPtr := *(*uint64)(unsafe.Pointer(&fieldBytes[0]))
					keyPtr := *(*uint64)(unsafe.Pointer(&keyBytes[0]))
					if fieldPtr == keyPtr {
						// 比较剩余字节
						match = true
						for i := 8; i < keyLen; i++ {
							if fieldBytes[i] != keyBytes[i] {
								match = false
								break
							}
						}
					} else {
						match = false
					}
				} else {
					// 短键使用逐字节比较
					match = true
					for i := 0; i < keyLen; i++ {
						if fieldBytes[i] != keyBytes[i] {
							match = false
							break
						}
					}
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
							return pos // 不完整的转义
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
				return pos // 不完整的字符串
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
		if depth > 0 {
			return pos // 不完整的对象
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

	c := data[pos]
	valStart := pos

	switch c {
	case '"':
		// 字符串：快速跳过到结尾
		valEnd := pos + 1
		for valEnd < end {
			if data[valEnd] == '"' {
				return Node{raw: data, start: valStart, end: valEnd + 1, typ: 's'}
			}
			if data[valEnd] == '\\' {
				valEnd++
			}
			valEnd++
		}
		return Node{raw: data, start: valStart, end: end, typ: 's'}
	case '{':
		return Node{raw: data, start: valStart, end: skipValueFast(data, pos, end), typ: 'o'}
	case '[':
		return Node{raw: data, start: valStart, end: skipValueFast(data, pos, end), typ: 'a'}
	case 't':
		return Node{raw: data, start: valStart, end: pos + 4, typ: 'b'}
	case 'f':
		return Node{raw: data, start: valStart, end: pos + 5, typ: 'b'}
	case 'n':
		return Node{raw: data, start: valStart, end: pos + 4, typ: 'l'}
	default:
		if c == '-' || (c >= '0' && c <= '9') {
			return Node{raw: data, start: valStart, end: skipValueFast(data, pos, end), typ: 'n'}
		}
	}
	return Node{}
}

// ===== 字面量取值 =====

// String 返回节点的字符串值
// 如果节点类型不是 JSON 字符串，或内容为空，则返回错误
func (n Node) String() (string, error) {
	if n.typ != 's' {
		return "", fmt.Errorf("node is not a string type (got type=%q)", n.Kind())
	}
	data := n.getWorkingData()
	// 增强边界检查
	if len(data) == 0 || n.start < 0 || n.end > len(data) || n.start >= n.end {
		return "", fmt.Errorf("invalid node bounds: start=%d end=%d len(data)=%d", n.start, n.end, len(data))
	}
	if n.start+1 >= n.end {
		return "", fmt.Errorf("invalid string bounds: start=%d end=%d", n.start, n.end)
	}

	bytes := data[n.start+1 : n.end-1]
	if len(bytes) == 0 {
		return "", nil // 空字符串正常返回
	}

	str := unsafe.String(&bytes[0], len(bytes))
	// 如果包含转义字符，需要解转义
	if strings.Contains(str, "\\") {
		return unescapeJSON(str), nil
	}

	return str, nil
}

// Int 返回节点的 int64 整数值
// 如果节点类型不是 JSON 数字、为空、包含非整数字符，或超出 int64 范围，则返回错误
func (n Node) Int() (int64, error) {
	if n.typ != 'n' || n.start >= n.end {
		return 0, fmt.Errorf("node is not a number type (got type=%q)", n.Kind())
	}
	workingData := n.getWorkingData()
	// 增强边界检查
	if len(workingData) == 0 || n.start < 0 || n.end > len(workingData) || n.start >= n.end {
		return 0, fmt.Errorf("invalid node bounds: start=%d end=%d len(data)=%d", n.start, n.end, len(workingData))
	}
	data := workingData[n.start:n.end]
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

// 其他数据类型转换方法...

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

// FloatString 返回数字的字符串表示，保持原始JSON格式的精度
// 优先返回原始JSON中的数字字符串，避免浮点数格式化问题
func (n Node) FloatString() (string, error) {
	if n.typ != 'n' || n.start >= n.end {
		return "", fmt.Errorf("not a number: got type=%q at range [%d:%d]", n.Kind(), n.start, n.end)
	}
	// 直接返回原始数字字符串，保持JSON中的精度格式
	return n.NumStr()
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

// ToJSON 将节点序列化为JSON字符串（压缩模式）
func (n Node) ToJSON() (string, error) {
	return n.ToJSONWithOptions(DefaultSerializeOptions)
}

// ToJSONIndent 将节点序列化为格式化的JSON字符串
func (n Node) ToJSONIndent(prefix, indent string) (string, error) {
	opts := PrettySerializeOptions
	opts.Indent = indent
	return n.ToJSONWithOptions(opts)
}

// ToJSONWithOptions 使用指定选项将节点序列化为JSON字符串
func (n Node) ToJSONWithOptions(opts SerializeOptions) (string, error) {
	if !n.Exists() {
		return "null", nil
	}

	buf := getBuffer()
	defer putBuffer(buf)

	if err := n.marshalNode(buf, opts, 0); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ToJSONBytes 将节点序列化为JSON字节切片（压缩模式）
func (n Node) ToJSONBytes() ([]byte, error) {
	return n.ToJSONBytesWithOptions(DefaultSerializeOptions)
}

// ToJSONBytesWithOptions 使用指定选项将节点序列化为JSON字节切片
func (n Node) ToJSONBytesWithOptions(opts SerializeOptions) ([]byte, error) {
	if !n.Exists() {
		return []byte("null"), nil
	}

	buf := getBuffer()
	defer putBuffer(buf)

	if err := n.marshalNode(buf, opts, 0); err != nil {
		return nil, err
	}

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result, nil
}

// ToJSONFast 快速序列化节点为JSON字符串（最小开销）
func (n Node) ToJSONFast() string {
	if !n.Exists() {
		return "null"
	}

	buf := getBuffer()
	defer putBuffer(buf)

	n.fastMarshalNode(buf)
	return buf.String()
}

// marshalNode 序列化节点
func (n Node) marshalNode(buf *Buffer, opts SerializeOptions, depth int) error {
	if !n.Exists() {
		buf.WriteString("null")
		return nil
	}

	data := n.getWorkingData()

	switch n.typ {
	case 'o':
		return n.marshalObject(buf, opts, depth)
	case 'a':
		return n.marshalArray(buf, opts, depth)
	case 's':
		str, err := n.String()
		if err != nil {
			return err
		}
		writeString(buf, str, opts.EscapeHTML)
		return nil
	case 'n':
		// 直接使用原始数字字符串，保持精度
		buf.Write(data[n.start:n.end])
		return nil
	case 'b':
		buf.Write(data[n.start:n.end])
		return nil
	case 'l':
		buf.WriteString("null")
		return nil
	default:
		return fmt.Errorf("unknown node type: %d", n.typ)
	}
}

// fastMarshalNode 快速序列化节点
func (n Node) fastMarshalNode(buf *Buffer) {
	if !n.Exists() {
		buf.WriteString("null")
		return
	}

	data := n.getWorkingData()

	switch n.typ {
	case 'o':
		n.fastMarshalObject(buf)
	case 'a':
		n.fastMarshalArray(buf)
	case 's':
		if str, err := n.String(); err == nil {
			writeStringFast(buf, str)
		} else {
			buf.WriteString("null")
		}
	case 'n', 'b', 'l':
		// 直接复制原始数据
		buf.Write(data[n.start:n.end])
	default:
		buf.WriteString("null")
	}
}

// marshalObject 序列化对象节点
func (n Node) marshalObject(buf *Buffer, opts SerializeOptions, depth int) error {
	buf.WriteByte('{')

	written := false
	indent := opts.Indent
	hasIndent := indent != ""

	if hasIndent {
		depth++
	}

	// 收集键值对
	var pairs []struct {
		key   string
		value Node
	}

	n.ForEach(func(key string, value Node) bool {
		// 处理omitempty
		if opts.OmitEmpty && n.isEmptyNode(value) {
			return true
		}

		pairs = append(pairs, struct {
			key   string
			value Node
		}{key, value})
		return true
	})

	// 排序键（如果启用）
	if opts.SortKeys && len(pairs) > 1 {
		sortNodePairs(pairs)
	}

	for _, pair := range pairs {
		if written {
			buf.WriteByte(',')
		}

		if hasIndent {
			buf.WriteByte('\n')
			writeIndent(buf, indent, depth)
		}

		// 写入键
		writeString(buf, pair.key, opts.EscapeHTML)
		buf.WriteByte(':')

		if hasIndent {
			buf.WriteByte(' ')
		}

		// 写入值
		if err := pair.value.marshalNode(buf, opts, depth); err != nil {
			return err
		}

		written = true
	}

	if hasIndent && written {
		buf.WriteByte('\n')
		writeIndent(buf, indent, depth-1)
	}

	buf.WriteByte('}')
	return nil
}

// fastMarshalObject 快速序列化对象节点
func (n Node) fastMarshalObject(buf *Buffer) {
	buf.WriteByte('{')
	written := false

	n.ForEach(func(key string, value Node) bool {
		if written {
			buf.WriteByte(',')
		}

		// 写入键
		writeStringFast(buf, key)
		buf.WriteByte(':')

		// 写入值
		value.fastMarshalNode(buf)
		written = true
		return true
	})

	buf.WriteByte('}')
}

// marshalArray 序列化数组节点
func (n Node) marshalArray(buf *Buffer, opts SerializeOptions, depth int) error {
	length := n.Len()

	buf.WriteByte('[')

	indent := opts.Indent
	hasIndent := indent != ""

	if hasIndent && length > 0 {
		depth++
	}

	for i := 0; i < length; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}

		if hasIndent {
			buf.WriteByte('\n')
			writeIndent(buf, indent, depth)
		}

		item := n.Index(i)
		if err := item.marshalNode(buf, opts, depth); err != nil {
			return err
		}
	}

	if hasIndent && length > 0 {
		buf.WriteByte('\n')
		writeIndent(buf, indent, depth-1)
	}

	buf.WriteByte(']')
	return nil
}

// fastMarshalArray 快速序列化数组节点
func (n Node) fastMarshalArray(buf *Buffer) {
	length := n.Len()

	buf.WriteByte('[')

	for i := 0; i < length; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}

		item := n.Index(i)
		item.fastMarshalNode(buf)
	}

	buf.WriteByte(']')
}

// isEmptyNode 检查节点是否为空
func (n Node) isEmptyNode(node Node) bool {
	if !node.Exists() {
		return true
	}

	switch node.typ {
	case 'a':
		return node.Len() == 0
	case 'o':
		isEmpty := true
		node.ForEach(func(key string, value Node) bool {
			isEmpty = false
			return false // 找到一个键就停止
		})
		return isEmpty
	case 's':
		if str, err := node.String(); err == nil {
			return str == ""
		}
		return false
	case 'n':
		if num, err := node.Float(); err == nil {
			return num == 0
		}
		return false
	case 'b':
		if b, err := node.Bool(); err == nil {
			return !b
		}
		return false
	case 'l':
		return true
	default:
		return false
	}
}

// sortNodePairs 排序节点键值对
func sortNodePairs(pairs []struct {
	key   string
	value Node
}) {
	// 使用简单的插入排序，对小数组效率更高
	for i := 1; i < len(pairs); i++ {
		key := pairs[i]
		j := i - 1
		for j >= 0 && pairs[j].key > key.key {
			pairs[j+1] = pairs[j]
			j--
		}
		pairs[j+1] = key
	}
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
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer: got kind=%s, type=%T", rv.Kind(), v)
	}
	if rv.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer: type=%T", v)
	}

	return n.decodeValueFast(rv.Elem())
}

// decodeValueFast 高性能解码实现
func (n Node) decodeValueFast(rv reflect.Value) error {
	if !rv.CanSet() {
		return fmt.Errorf("cannot set value of type %s", rv.Type())
	}

	// 快速路径：直接处理常见类型，避免反射开销
	switch n.typ {
	case 'l': // null
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	case 's': // string
		return n.decodeStringFast(rv)
	case 'n': // number
		return n.decodeNumberFast(rv)
	case 'b': // bool
		return n.decodeBoolFast(rv)
	case 'a': // array
		return n.decodeArrayFast(rv)
	case 'o': // object
		return n.decodeObjectFast(rv)
	default:
		return fmt.Errorf("unknown JSON type: %d", n.Kind())
	}
}

// decodeStringFast 快速字符串解码
func (n Node) decodeStringFast(rv reflect.Value) error {
	data := n.getWorkingData()
	if n.start+1 >= n.end {
		return fmt.Errorf("invalid string bounds")
	}

	// 零拷贝字符串提取
	strBytes := data[n.start+1 : n.end-1]
	var str string
	if len(strBytes) == 0 {
		str = "" // 安全处理空字符串
	} else {
		str = unsafe.String(&strBytes[0], len(strBytes))
	}

	// 仅在需要时进行转义处理
	if strings.Contains(str, "\\") {
		str = unescapeJSON(str)
	}

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(str)
		return nil
	case reflect.Interface:
		rv.Set(reflect.ValueOf(str))
		return nil
	default:
		return fmt.Errorf("cannot decode string to %s", rv.Type())
	}
}

// decodeNumberFast 快速数字解码
func (n Node) decodeNumberFast(rv reflect.Value) error {
	data := n.getWorkingData()
	numBytes := data[n.start:n.end]

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(unsafe.String(&numBytes[0], len(numBytes)))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := parseIntFast(numBytes)
		if err != nil {
			return err
		}
		rv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := parseIntFast(numBytes)
		if err != nil {
			return err
		}
		if i < 0 {
			return fmt.Errorf("cannot assign negative number %d to unsigned type %s", i, rv.Type())
		}
		rv.SetUint(uint64(i))
		return nil
	case reflect.Float32, reflect.Float64:
		f := parseFloatFast(numBytes)
		rv.SetFloat(f)
		return nil
	case reflect.Interface:
		// 智能类型推断：整数 vs 浮点数
		if !strings.Contains(string(numBytes), ".") && !strings.ContainsAny(string(numBytes), "eE") {
			if i, err := parseIntFast(numBytes); err == nil {
				rv.Set(reflect.ValueOf(i))
				return nil
			}
		}
		f := parseFloatFast(numBytes)
		rv.Set(reflect.ValueOf(f))
		return nil
	default:
		return fmt.Errorf("cannot decode number to %s", rv.Type())
	}
}

// decodeBoolFast 快速布尔解码
func (n Node) decodeBoolFast(rv reflect.Value) error {
	data := n.getWorkingData()
	boolBytes := data[n.start:n.end]

	var b bool
	if len(boolBytes) == 4 && string(boolBytes) == "true" {
		b = true
	} else if len(boolBytes) == 5 && string(boolBytes) == "false" {
		b = false
	} else {
		return fmt.Errorf("invalid bool value: %s", string(boolBytes))
	}

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(string(boolBytes))
		return nil
	case reflect.Bool:
		rv.SetBool(b)
		return nil
	case reflect.Interface:
		rv.Set(reflect.ValueOf(b))
		return nil
	default:
		return fmt.Errorf("cannot decode bool to %s", rv.Type())
	}
}

// decodeArrayFast 快速数组解码
func (n Node) decodeArrayFast(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Slice:
		return n.decodeSliceFast(rv)
	case reflect.Array:
		return n.decodeArrayFixedFast(rv)
	case reflect.Interface:
		// 使用预分配容量避免扩容
		length := n.Len()
		slice := make([]interface{}, 0, length)

		var decodeErr error
		n.ArrayForEach(func(i int, child Node) bool {
			var elem interface{}
			elemRV := reflect.ValueOf(&elem).Elem()
			if err := child.decodeValueFast(elemRV); err != nil {
				decodeErr = err
				return false
			}
			slice = append(slice, elem)
			return true
		})

		if decodeErr != nil {
			return decodeErr
		}
		rv.Set(reflect.ValueOf(slice))
		return nil
	default:
		return fmt.Errorf("cannot decode array to %s", rv.Type())
	}
}

// decodeSliceFast 快速slice解码
func (n Node) decodeSliceFast(rv reflect.Value) error {
	length := n.Len()
	slice := reflect.MakeSlice(rv.Type(), length, length)

	var decodeErr error
	n.ArrayForEach(func(i int, child Node) bool {
		if decodeErr != nil {
			return false
		}
		if i < length {
			decodeErr = child.decodeValueFast(slice.Index(i))
		}
		return decodeErr == nil
	})

	if decodeErr != nil {
		return decodeErr
	}

	rv.Set(slice)
	return nil
}

// decodeArrayFixedFast 快速固定数组解码
func (n Node) decodeArrayFixedFast(rv reflect.Value) error {
	length := rv.Len()

	var decodeErr error
	n.ArrayForEach(func(i int, child Node) bool {
		if decodeErr != nil {
			return false
		}
		if i < length {
			decodeErr = child.decodeValueFast(rv.Index(i))
		}
		return decodeErr == nil
	})

	return decodeErr
}

// decodeObjectFast 快速对象解码
func (n Node) decodeObjectFast(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Struct:
		return n.decodeStructFast(rv)
	case reflect.Map:
		return n.decodeMapFast(rv)
	case reflect.Interface:
		// 使用预估容量减少map扩容
		m := make(map[string]interface{}, n.Len())

		var decodeErr error
		n.ForEach(func(key string, child Node) bool {
			if decodeErr != nil {
				return false
			}
			var val interface{}
			valRV := reflect.ValueOf(&val).Elem()
			if err := child.decodeValueFast(valRV); err != nil {
				decodeErr = err
				return false
			}
			m[key] = val
			return true
		})

		if decodeErr != nil {
			return decodeErr
		}
		rv.Set(reflect.ValueOf(m))
		return nil
	default:
		return fmt.Errorf("cannot decode object to %s", rv.Type())
	}
}

// decodeStructFast 快速结构体解码（缓存优化版本）
func (n Node) decodeStructFast(rv reflect.Value) error {
	structType := rv.Type()
	fieldMap := getStructFieldMapFast(structType)

	var decodeErr error
	n.ForEach(func(key string, child Node) bool {
		if decodeErr != nil {
			return false
		}

		if fieldInfo, exists := fieldMap[key]; exists {
			fieldValue := rv.Field(fieldInfo.Index)
			if fieldValue.CanSet() {
				decodeErr = child.decodeValueFast(fieldValue)
			}
		}
		return decodeErr == nil
	})

	return decodeErr
}

// decodeMapFast 快速map解码
func (n Node) decodeMapFast(rv reflect.Value) error {
	mapType := rv.Type()
	keyType := mapType.Key()
	valueType := mapType.Elem()

	if keyType.Kind() != reflect.String {
		return fmt.Errorf("map key must be string, got %s", keyType)
	}

	// 预分配容量
	m := reflect.MakeMapWithSize(mapType, n.Len())

	var decodeErr error
	n.ForEach(func(key string, child Node) bool {
		if decodeErr != nil {
			return false
		}

		keyVal := reflect.ValueOf(key)
		valueVal := reflect.New(valueType).Elem()

		if err := child.decodeValueFast(valueVal); err != nil {
			decodeErr = err
			return false
		}

		m.SetMapIndex(keyVal, valueVal)
		return true
	})

	if decodeErr != nil {
		return decodeErr
	}

	rv.Set(m)
	return nil
}

// getStructFieldMapFast 快速结构体字段映射（优化版本）
func getStructFieldMapFast(t reflect.Type) map[string]structFieldInfo {
	if cached, ok := structFieldCache.Load(t); ok {
		return cached.(map[string]structFieldInfo)
	}

	fieldMap := make(map[string]structFieldInfo, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		jsonName := getJSONFieldNameFast(field)
		if jsonName == "-" {
			continue
		}

		fieldMap[jsonName] = structFieldInfo{
			Index:    i,
			JSONName: jsonName,
		}
	}

	structFieldCache.Store(t, fieldMap)
	return fieldMap
}

// getJSONFieldNameFast 快速JSON字段名提取
func getJSONFieldNameFast(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	// 快速解析：只取第一个逗号前的部分
	if idx := strings.IndexByte(tag, ','); idx != -1 {
		tag = tag[:idx]
	}

	tag = strings.TrimSpace(tag)
	if tag == "" {
		return field.Name
	}

	return tag
}

// decodeValue 递归解码JSON值到reflect.Value
func (n Node) decodeValue(rv reflect.Value) error {
	if !rv.CanSet() {
		return fmt.Errorf("cannot set value of type %s", rv.Type())
	}

	switch n.Kind() {
	case TypeNull:
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	case TypeString:
		return n.decodeString(rv)
	case TypeNumber:
		return n.decodeNumber(rv)
	case TypeBool:
		return n.decodeBool(rv)
	case TypeArray:
		return n.decodeArray(rv)
	case TypeObject:
		return n.decodeObject(rv)
	default:
		return fmt.Errorf("unknown JSON type: %d", n.Kind())
	}
}

// decodeString 解码字符串值
func (n Node) decodeString(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.String:
		str, err := n.String()
		if err != nil {
			return err
		}
		rv.SetString(str)
		return nil
	case reflect.Interface:
		str, err := n.String()
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(str))
		return nil
	default:
		return fmt.Errorf("cannot decode string to %s", rv.Type())
	}
}

// decodeNumber 解码数字值
func (n Node) decodeNumber(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.String:
		// 支持将数字转换为字符串
		data := n.getWorkingData()
		jsonBytes := data[n.start:n.end]
		rv.SetString(string(jsonBytes))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := n.Int()
		if err != nil {
			return err
		}
		rv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := n.Int()
		if err != nil {
			return err
		}
		if i < 0 {
			return fmt.Errorf("cannot assign negative number %d to unsigned type %s", i, rv.Type())
		}
		rv.SetUint(uint64(i))
		return nil
	case reflect.Float32, reflect.Float64:
		f, err := n.Float()
		if err != nil {
			return err
		}
		rv.SetFloat(f)
		return nil
	case reflect.Interface:
		// 尝试解析为int，如果失败则解析为float
		if i, err := n.Int(); err == nil {
			rv.Set(reflect.ValueOf(i))
			return nil
		}
		f, err := n.Float()
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(f))
		return nil
	default:
		return fmt.Errorf("cannot decode number to %s", rv.Type())
	}
}

// decodeBool 解码布尔值
func (n Node) decodeBool(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.String:
		// 支持将布尔值转换为字符串
		data := n.getWorkingData()
		jsonBytes := data[n.start:n.end]
		rv.SetString(string(jsonBytes))
		return nil
	case reflect.Bool:
		b, err := n.Bool()
		if err != nil {
			return err
		}
		rv.SetBool(b)
		return nil
	case reflect.Interface:
		b, err := n.Bool()
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(b))
		return nil
	default:
		return fmt.Errorf("cannot decode bool to %s", rv.Type())
	}
}

// decodeArray 解码数组值
func (n Node) decodeArray(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Slice:
		return n.decodeSlice(rv)
	case reflect.Array:
		return n.decodeArrayFixed(rv)
	case reflect.Interface:
		// 创建一个slice来存储数组元素
		slice := make([]interface{}, 0, n.Len())
		n.ArrayForEach(func(i int, child Node) bool {
			var elem interface{}
			elemRV := reflect.ValueOf(&elem).Elem()
			if err := child.decodeValue(elemRV); err == nil {
				slice = append(slice, elem)
			}
			return true
		})
		rv.Set(reflect.ValueOf(slice))
		return nil
	default:
		return fmt.Errorf("cannot decode array to %s", rv.Type())
	}
}

// decodeSlice 解码到slice
func (n Node) decodeSlice(rv reflect.Value) error {
	length := n.Len()
	slice := reflect.MakeSlice(rv.Type(), length, length)

	var decodeErr error
	n.ArrayForEach(func(i int, child Node) bool {
		if decodeErr != nil {
			return false
		}
		if i < length {
			decodeErr = child.decodeValue(slice.Index(i))
		}
		return true
	})

	if decodeErr != nil {
		return decodeErr
	}

	rv.Set(slice)
	return nil
}

// decodeArrayFixed 解码到固定长度数组
func (n Node) decodeArrayFixed(rv reflect.Value) error {
	length := rv.Len()

	var decodeErr error
	n.ArrayForEach(func(i int, child Node) bool {
		if decodeErr != nil {
			return false
		}
		if i < length {
			decodeErr = child.decodeValue(rv.Index(i))
		}
		return true
	})

	return decodeErr
}

// decodeObject 解码对象值
func (n Node) decodeObject(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Struct:
		return n.decodeStruct(rv)
	case reflect.Map:
		return n.decodeMap(rv)
	case reflect.Interface:
		// 创建map[string]interface{}来存储对象
		m := make(map[string]interface{})
		n.ForEach(func(key string, child Node) bool {
			var val interface{}
			valRV := reflect.ValueOf(&val).Elem()
			if err := child.decodeValue(valRV); err == nil {
				m[key] = val
			}
			return true
		})
		rv.Set(reflect.ValueOf(m))
		return nil
	default:
		return fmt.Errorf("cannot decode object to %s", rv.Type())
	}
}

// decodeStruct 解码到结构体
func (n Node) decodeStruct(rv reflect.Value) error {
	structType := rv.Type()
	fieldMap := getStructFieldMap(structType)

	var decodeErr error
	n.ForEach(func(key string, child Node) bool {
		if decodeErr != nil {
			return false
		}

		if fieldInfo, exists := fieldMap[key]; exists {
			fieldValue := rv.Field(fieldInfo.Index)
			if fieldValue.CanSet() {
				decodeErr = child.decodeValue(fieldValue)
			}
		}
		return true
	})

	return decodeErr
}

// decodeMap 解码到map
func (n Node) decodeMap(rv reflect.Value) error {
	mapType := rv.Type()
	keyType := mapType.Key()
	valueType := mapType.Elem()

	if keyType.Kind() != reflect.String {
		return fmt.Errorf("map key must be string, got %s", keyType)
	}

	m := reflect.MakeMap(mapType)

	var decodeErr error
	n.ForEach(func(key string, child Node) bool {
		if decodeErr != nil {
			return false
		}

		keyVal := reflect.ValueOf(key)
		valueVal := reflect.New(valueType).Elem()

		if err := child.decodeValue(valueVal); err != nil {
			decodeErr = err
			return false
		}

		m.SetMapIndex(keyVal, valueVal)
		return true
	})

	if decodeErr != nil {
		return decodeErr
	}

	rv.Set(m)
	return nil
}

// structFieldInfo 存储结构体字段信息
type structFieldInfo struct {
	Index    int    // 字段在结构体中的索引
	JSONName string // JSON标签名或字段名
}

// structFieldMap 缓存结构体字段映射
var structFieldCache = sync.Map{}

// getStructFieldMap 获取结构体字段映射
func getStructFieldMap(t reflect.Type) map[string]structFieldInfo {
	if cached, ok := structFieldCache.Load(t); ok {
		return cached.(map[string]structFieldInfo)
	}

	fieldMap := make(map[string]structFieldInfo)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		jsonName := getJSONFieldName(field)
		if jsonName == "-" {
			// json:"-" 表示忽略此字段
			continue
		}

		fieldMap[jsonName] = structFieldInfo{
			Index:    i,
			JSONName: jsonName,
		}
	}

	structFieldCache.Store(t, fieldMap)
	return fieldMap
}

// getJSONFieldName 从结构体字段获取JSON字段名
func getJSONFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	// 解析json标签: "name,omitempty" -> "name"
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return field.Name
	}

	jsonName := strings.TrimSpace(parts[0])
	if jsonName == "" {
		return field.Name
	}

	return jsonName
}

// DecodeStruct 是一个优化版本的Decode方法，专门用于结构体解码
// 避免创建Node的开销，直接使用字节切片
func DecodeStruct(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer: got kind=%s, type=%T", rv.Kind(), v)
	}
	if rv.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer: type=%T", v)
	}

	// 直接解析，避免FromBytes的额外开销
	return decodeStructFromBytes(data, rv.Elem())
}

// decodeStructFromBytes 直接从字节切片解码到结构体，避免Node创建开销
func decodeStructFromBytes(data []byte, rv reflect.Value) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}

	// 快速查找JSON对象的开始和结束
	start := 0
	for start < len(data) && data[start] <= ' ' {
		start++
	}
	if start >= len(data) || data[start] != '{' {
		return fmt.Errorf("data is not a JSON object")
	}

	end := len(data)
	node := Node{raw: data, start: start, end: end, typ: 'o'}
	return node.decodeStructFast(rv)
}

// DecodeStructFast 极致优化的结构体解码函数
func DecodeStructFast(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer: got kind=%s, type=%T", rv.Kind(), v)
	}
	if rv.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer: type=%T", v)
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("v must point to a struct, got %s", elem.Kind())
	}

	return decodeStructDirectly(data, elem)
}

// decodeStructDirectly 直接解码结构体，跳过所有中间步骤
func decodeStructDirectly(data []byte, rv reflect.Value) error {
	if len(data) == 0 {
		return fmt.Errorf("empty JSON data")
	}

	// 获取结构体类型信息
	structType := rv.Type()
	fieldMap := getStructFieldMapFast(structType)

	// 快速扫描JSON对象
	pos := 0
	for pos < len(data) && data[pos] <= ' ' {
		pos++
	}
	if pos >= len(data) || data[pos] != '{' {
		return fmt.Errorf("invalid JSON object")
	}
	pos++ // skip '{'

	for pos < len(data) {
		// 跳过空白
		for pos < len(data) && data[pos] <= ' ' {
			pos++
		}
		if pos >= len(data) || data[pos] == '}' {
			break
		}

		// 解析键
		if data[pos] != '"' {
			return fmt.Errorf("expected key at position %d", pos)
		}
		pos++
		keyStart := pos

		// 快速键扫描
		for pos < len(data) && data[pos] != '"' {
			if data[pos] == '\\' {
				pos += 2
			} else {
				pos++
			}
		}
		keyEnd := pos
		pos++ // skip closing quote

		// 零拷贝键提取
		key := unsafe.String(&data[keyStart], keyEnd-keyStart)

		// 跳过冒号
		for pos < len(data) && data[pos] <= ' ' {
			pos++
		}
		if pos >= len(data) || data[pos] != ':' {
			return fmt.Errorf("expected ':' at position %d", pos)
		}
		pos++
		for pos < len(data) && data[pos] <= ' ' {
			pos++
		}

		// 查找对应的结构体字段
		if fieldInfo, exists := fieldMap[key]; exists {
			fieldValue := rv.Field(fieldInfo.Index)
			if fieldValue.CanSet() {
				// 解析值并直接设置到字段
				valueEnd := skipValueFast(data, pos, len(data))
				if valueEnd <= pos {
					return fmt.Errorf("invalid value at position %d", pos)
				}

				// 创建临时节点进行解码
				valueNode := Node{
					raw:   data,
					start: pos,
					end:   valueEnd,
					typ:   detectType(data[pos]),
				}

				if err := valueNode.decodeValueFast(fieldValue); err != nil {
					return fmt.Errorf("failed to decode field %s: %v", key, err)
				}

				pos = valueEnd
			} else {
				// 跳过无法设置的字段
				pos = skipValueFast(data, pos, len(data))
			}
		} else {
			// 跳过未匹配的字段
			pos = skipValueFast(data, pos, len(data))
		}

		// 跳过逗号
		for pos < len(data) && data[pos] <= ' ' {
			pos++
		}
		if pos < len(data) && data[pos] == ',' {
			pos++
		}
	}

	return nil
}

func fastDecode(buf []byte, start, end int) (any, int, error) {
	if start >= end {
		return nil, start, fmt.Errorf("empty node: start=%d, end=%d, len(buf)=%d", start, end, len(buf))
	}

	// 使用skipValueFast来确定值的边界，然后递归解析
	valueEnd := skipValueFast(buf, start, end)
	if valueEnd == start {
		return nil, start, fmt.Errorf("invalid JSON at position %d", start)
	}

	switch buf[start] {
	case '{':
		m := make(map[string]any)
		i := start + 1
		for i < valueEnd-1 {
			// 跳过空白
			for i < valueEnd-1 && buf[i] <= ' ' {
				i++
			}
			if i >= valueEnd-1 {
				break
			}

			// 解析键
			if buf[i] != '"' {
				return nil, i, fmt.Errorf("expected key at position %d", i)
			}
			keyEnd := skipValueFast(buf, i, valueEnd)
			key := unsafe.String(&buf[i+1], keyEnd-i-2)
			i = keyEnd

			// 跳过空白和冒号
			for i < valueEnd-1 && (buf[i] <= ' ' || buf[i] == ':') {
				i++
			}

			// 解析值
			val, ni, err := fastDecode(buf, i, valueEnd)
			if err != nil {
				return nil, ni, err
			}
			m[key] = val
			i = ni

			// 跳过逗号
			for i < valueEnd-1 && (buf[i] <= ' ' || buf[i] == ',') {
				i++
			}
		}
		return m, valueEnd, nil

	case '[':
		arr := make([]any, 0)
		i := start + 1
		for i < valueEnd-1 {
			// 跳过空白
			for i < valueEnd-1 && buf[i] <= ' ' {
				i++
			}
			if i >= valueEnd-1 {
				break
			}

			// 解析值
			val, ni, err := fastDecode(buf, i, valueEnd)
			if err != nil {
				return nil, ni, err
			}
			arr = append(arr, val)
			i = ni

			// 跳过逗号
			for i < valueEnd-1 && (buf[i] <= ' ' || buf[i] == ',') {
				i++
			}
		}
		return arr, valueEnd, nil

	case '"':
		str := unsafe.String(&buf[start+1], valueEnd-start-2)
		return str, valueEnd, nil

	case 't':
		return true, valueEnd, nil

	case 'f':
		return false, valueEnd, nil

	case 'n':
		return nil, valueEnd, nil

	default:
		// 数字
		fv := parseFloatFast(buf[start:valueEnd])
		return fv, valueEnd, nil
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

// ===== 遍历相关接口和类型 =====

// ForEachFunc 对象遍历回调函数类型
// key: 对象键名, value: 对象值节点
// 返回 false 可以提前终止遍历
type ForEachFunc func(key string, value Node) bool

// ArrayForEachFunc 数组遍历回调函数类型
// index: 数组索引, value: 数组元素节点
// 返回 false 可以提前终止遍历
type ArrayForEachFunc func(index int, value Node) bool

// ===== 对象遍历 =====

// ForEach 遍历对象的所有键值对（极限优化版本）
// 只有当节点是对象类型时才会执行遍历，否则直接返回
// 遍历过程中如果回调函数返回 false，则提前终止遍历
func (n Node) ForEach(fn ForEachFunc) {
	if n.typ != 'o' || fn == nil {
		return
	}

	data := n.getWorkingData()
	pos := n.start + 1 // 直接跳过 '{'
	end := n.end
	endMinus1 := end - 1 // 预计算边界

	// 批处理优化：预分配键值对缓冲区
	type keyValuePair struct {
		keyStart, keyEnd     int
		valueStart, valueEnd int
		valueType            byte
	}

	// 首先快速扫描收集所有键值对位置（避免重复遍历）
	var pairs [32]keyValuePair // 栈分配，避免堆分配
	pairCount := 0
	scanPos := pos

	for scanPos < endMinus1 && pairCount < 32 {
		// 快速空白跳过
		for scanPos < endMinus1 && data[scanPos] <= ' ' {
			scanPos++
		}
		if scanPos >= endMinus1 || data[scanPos] == '}' {
			break
		}

		// 快速键解析
		if data[scanPos] != '"' {
			break
		}
		scanPos++
		keyStart := scanPos

		// 优化键扫描：大部分键没有转义字符
		for scanPos < end && data[scanPos] != '"' {
			if data[scanPos] == '\\' {
				scanPos += 2 // 跳过转义
			} else {
				scanPos++
			}
		}
		keyEnd := scanPos
		scanPos++ // skip closing quote

		// 快速冒号跳过
		for scanPos < end && data[scanPos] <= ' ' {
			scanPos++
		}
		if scanPos >= end || data[scanPos] != ':' {
			break
		}
		scanPos++
		for scanPos < end && data[scanPos] <= ' ' {
			scanPos++
		}

		// 记录值位置
		valueStart := scanPos
		valueEnd := skipValueFastInline(data, scanPos, end)
		if valueEnd <= scanPos {
			break
		}

		// 存储键值对信息
		pairs[pairCount] = keyValuePair{
			keyStart:   keyStart,
			keyEnd:     keyEnd,
			valueStart: valueStart,
			valueEnd:   valueEnd,
			valueType:  detectType(data[valueStart]),
		}
		pairCount++
		scanPos = valueEnd

		// 快速逗号跳过
		for scanPos < end && data[scanPos] <= ' ' {
			scanPos++
		}
		if scanPos < end && data[scanPos] == ',' {
			scanPos++
		}
	}

	// 批量处理所有键值对，减少函数调用开销
	for i := 0; i < pairCount; i++ {
		pair := pairs[i]

		// 创建键字符串（零拷贝）
		key := unsafe.String(&data[pair.keyStart], pair.keyEnd-pair.keyStart)

		// 创建值节点
		valueNode := Node{
			raw:      n.raw,
			start:    pair.valueStart,
			end:      pair.valueEnd,
			typ:      pair.valueType,
			expanded: n.expanded,
		}

		if !fn(key, valueNode) {
			break
		}
	}

	// 如果有超过32个键值对，回退到流式处理
	if pairCount == 32 && scanPos < endMinus1 {
		pos = scanPos
		for pos < endMinus1 {
			// 继续处理剩余的键值对
			for pos < endMinus1 && data[pos] <= ' ' {
				pos++
			}
			if pos >= endMinus1 || data[pos] == '}' {
				break
			}

			if data[pos] != '"' {
				break
			}
			pos++
			keyStart := pos

			for pos < end && data[pos] != '"' {
				if data[pos] == '\\' {
					pos += 2
				} else {
					pos++
				}
			}
			keyEnd := pos
			pos++

			for pos < end && data[pos] <= ' ' {
				pos++
			}
			if pos >= end || data[pos] != ':' {
				break
			}
			pos++
			for pos < end && data[pos] <= ' ' {
				pos++
			}

			valueStart := pos
			valueEnd := skipValueFastInline(data, pos, end)
			if valueEnd <= pos {
				break
			}

			valueNode := Node{
				raw:      n.raw,
				start:    valueStart,
				end:      valueEnd,
				typ:      detectType(data[valueStart]),
				expanded: n.expanded,
			}

			key := unsafe.String(&data[keyStart], keyEnd-keyStart)
			if !fn(key, valueNode) {
				break
			}

			pos = valueEnd
			for pos < end && data[pos] <= ' ' {
				pos++
			}
			if pos < end && data[pos] == ',' {
				pos++
			}
		}
	}
}

// skipValueFastInline 内联优化的值跳过函数
func skipValueFastInline(data []byte, pos int, end int) int {
	if pos >= end {
		return pos
	}

	c := data[pos]
	switch c {
	case '{', '[':
		pos++
		depth := 1

		for pos < end && depth > 0 {
			c = data[pos]
			if c == '"' {
				// 快速字符串跳过
				pos++
				for pos < end {
					if data[pos] == '"' {
						pos++
						break
					}
					if data[pos] == '\\' {
						pos += 2
					} else {
						pos++
					}
				}
			} else if c == '{' || c == '[' {
				depth++
				pos++
			} else if c == '}' || c == ']' {
				depth--
				pos++
			} else {
				pos++
			}
		}
		return pos

	case '"':
		pos++
		for pos < end {
			if data[pos] == '"' {
				return pos + 1
			}
			if data[pos] == '\\' {
				pos += 2
				if pos > end { // 防止转义字符越界
					return end
				}
			} else {
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
		// 数字
		if c == '-' || (c >= '0' && c <= '9') {
			pos++
			for pos < end {
				c = data[pos]
				if (c >= '0' && c <= '9') || c == '.' || c == 'e' || c == 'E' || c == '+' || c == '-' {
					pos++
				} else {
					break
				}
			}
		}
		return pos
	}
}

// ===== 数组遍历 =====

// ArrayForEach 遍历数组的所有元素（极限优化版本）
// 只有当节点是数组类型时才会执行遍历，否则直接返回
// 遍历过程中如果回调函数返回 false，则提前终止遍历
func (n Node) ArrayForEach(fn ArrayForEachFunc) {
	if n.typ != 'a' || fn == nil {
		return
	}

	// 尝试使用缓存的数组偏移
	offsets := buildArrOffsetsCached(n)
	if len(offsets) > 0 {
		// 使用缓存的偏移进行高速遍历
		data := n.getWorkingData()
		for i, offset := range offsets {
			valueEnd := skipValueFastInline(data, offset, n.end)
			valueNode := Node{
				raw:      n.raw,
				start:    offset,
				end:      valueEnd,
				typ:      detectType(data[offset]),
				expanded: n.expanded,
			}

			if !fn(i, valueNode) {
				break
			}
		}
		return
	}

	// 回退到内联解析（用于没有缓存的情况）
	data := n.getWorkingData()
	pos := n.start + 1 // 直接跳过 '['
	end := n.end
	index := 0

	// 预计算常用边界，减少条件检查
	endMinus1 := end - 1

	for pos < endMinus1 {
		// 极致优化的空白跳过
		for pos < endMinus1 && data[pos] <= ' ' {
			pos++
		}
		if pos >= endMinus1 || data[pos] == ']' {
			break
		}

		// 批量处理：检查是否是简单数字，可以快速跳过
		valueStart := pos
		c := data[pos]
		var valueEnd int

		if c >= '0' && c <= '9' || c == '-' {
			// 数字快速路径
			pos++
			for pos < end && data[pos] >= '0' && data[pos] <= '9' {
				pos++
			}
			// 检查是否有小数点或指数
			if pos < end && (data[pos] == '.' || data[pos] == 'e' || data[pos] == 'E') {
				valueEnd = skipValueFastInline(data, valueStart, end)
			} else {
				valueEnd = pos
			}
		} else {
			// 其他类型使用内联跳过
			valueEnd = skipValueFastInline(data, pos, end)
		}

		if valueEnd <= pos {
			break
		}

		// 批量创建节点，减少分支
		valueNode := Node{
			raw:      n.raw,
			start:    valueStart,
			end:      valueEnd,
			typ:      detectType(data[valueStart]),
			expanded: n.expanded,
		}

		if !fn(index, valueNode) {
			break
		}

		pos = valueEnd
		index++

		// 极致优化的逗号跳过
		for pos < end && data[pos] <= ' ' {
			pos++
		}
		if pos < end && data[pos] == ',' {
			pos++
		}
	}
}

// ===== 批量获取方法 =====

// GetAllKeys 返回对象的所有键名（字符串形式）
func (n Node) GetAllKeys() []string {
	if n.typ != 'o' {
		return nil
	}

	var keys []string
	n.ForEach(func(key string, value Node) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}

// GetAllValues 返回数组的所有元素节点
func (n Node) GetAllValues() []Node {
	if n.typ != 'a' {
		return nil
	}

	var values []Node
	n.ArrayForEach(func(index int, value Node) bool {
		values = append(values, value)
		return true
	})
	return values
}

// ToMap 将对象节点转换为 map[string]Node
func (n Node) ToMap() map[string]Node {
	if n.typ != 'o' {
		return nil
	}

	result := make(map[string]Node)
	n.ForEach(func(key string, value Node) bool {
		result[key] = value
		return true
	})
	return result
}

// ToSlice 将数组节点转换为 []Node
func (n Node) ToSlice() []Node {
	if n.typ != 'a' {
		return nil
	}
	return n.GetAllValues()
}

// ===== 条件查找方法 =====

// FindInObject 在对象中查找满足条件的第一个键值对
func (n Node) FindInObject(predicate func(key string, value Node) bool) (string, Node, bool) {
	if n.typ != 'o' || predicate == nil {
		return "", Node{}, false
	}

	var foundKey string
	var foundValue Node
	found := false

	n.ForEach(func(key string, value Node) bool {
		if predicate(key, value) {
			foundKey = key
			foundValue = value
			found = true
			return false // 找到后停止遍历
		}
		return true
	})

	return foundKey, foundValue, found
}

// FindInArray 在数组中查找满足条件的第一个元素
func (n Node) FindInArray(predicate func(index int, value Node) bool) (int, Node, bool) {
	if n.typ != 'a' || predicate == nil {
		return -1, Node{}, false
	}

	var foundIndex int = -1
	var foundValue Node
	found := false

	n.ArrayForEach(func(index int, value Node) bool {
		if predicate(index, value) {
			foundIndex = index
			foundValue = value
			found = true
			return false // 找到后停止遍历
		}
		return true
	})

	return foundIndex, foundValue, found
}

// FilterArray 过滤数组元素，返回满足条件的所有元素
func (n Node) FilterArray(predicate func(index int, value Node) bool) []Node {
	if n.typ != 'a' || predicate == nil {
		return nil
	}

	var result []Node
	n.ArrayForEach(func(index int, value Node) bool {
		if predicate(index, value) {
			result = append(result, value)
		}
		return true
	})

	return result
}

// ===== 深度遍历方法 =====

// WalkFunc 深度遍历回调函数类型
// path: 当前节点的路径（如 "data.notes[0].comments_count"）
// node: 当前节点
// 返回 false 可以跳过当前节点的子节点遍历
type WalkFunc func(path string, node Node) bool

// walkItem 表示遍历栈中的一个项目
type walkItem struct {
	node Node
	path string
}

// Walk 深度优先遍历整个JSON树（零分配优化实现）
func (n Node) Walk(fn WalkFunc) {
	if fn == nil || !n.Exists() {
		return
	}

	// 使用显式栈避免递归开销，预分配足够大的容量
	stack := make([]walkItem, 0, 64)
	stack = append(stack, walkItem{node: n, path: ""})

	// 预分配大容量避免扩容
	var pathBuilder strings.Builder
	pathBuilder.Grow(512)

	// 复用字节缓冲区避免重复分配
	var pathBytes [512]byte

	for len(stack) > 0 {
		// 出栈
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// 调用回调函数
		if !fn(current.path, current.node) {
			continue // 跳过子节点
		}

		// 处理子节点（直接内联遍历避免额外分配）
		switch current.node.typ {
		case 'o':
			// 对象：直接内联遍历，避免GetAllKeys()的分配
			n := current.node
			data := n.getWorkingData()
			pos := n.start + 1
			end := n.end
			pathLen := len(current.path)

			// 收集键值对（逆序）
			type keyValue struct {
				key   string
				value Node
			}
			var pairs []keyValue

			for pos < end-1 {
				// 跳过空白
				for pos < end && data[pos] <= ' ' {
					pos++
				}
				if pos >= end || data[pos] == '}' {
					break
				}

				// 解析键
				if data[pos] != '"' {
					break
				}
				pos++
				keyStart := pos
				for pos < end && data[pos] != '"' {
					if data[pos] == '\\' {
						pos += 2
					} else {
						pos++
					}
				}
				keyEnd := pos
				pos++ // skip closing quote

				key := unsafe.String(&data[keyStart], keyEnd-keyStart)

				// 跳过冒号
				for pos < end && data[pos] <= ' ' {
					pos++
				}
				if pos < end && data[pos] == ':' {
					pos++
				}
				for pos < end && data[pos] <= ' ' {
					pos++
				}

				// 解析值
				valueStart := pos
				valueEnd := skipValueFastInline(data, pos, end)
				if valueEnd <= pos {
					break
				}

				value := Node{
					raw:      n.raw,
					start:    valueStart,
					end:      valueEnd,
					typ:      detectType(data[valueStart]),
					expanded: n.expanded,
				}

				pairs = append(pairs, keyValue{key: key, value: value})
				pos = valueEnd

				// 跳过逗号
				for pos < end && data[pos] <= ' ' {
					pos++
				}
				if pos < end && data[pos] == ',' {
					pos++
				}
			}

			// 逆序添加到栈
			for i := len(pairs) - 1; i >= 0; i-- {
				pair := pairs[i]

				// 使用字节缓冲区构建路径，避免字符串分配
				pathPos := 0
				if pathLen > 0 {
					copy(pathBytes[pathPos:], current.path)
					pathPos += pathLen
					pathBytes[pathPos] = '.'
					pathPos++
				}
				copy(pathBytes[pathPos:], pair.key)
				pathPos += len(pair.key)

				newPath := string(pathBytes[:pathPos])
				stack = append(stack, walkItem{
					node: pair.value,
					path: newPath,
				})
			}

		case 'a':
			// 数组：直接内联遍历
			n := current.node
			length := n.Len()

			for i := length - 1; i >= 0; i-- {
				value := n.Index(i)
				if !value.Exists() {
					continue
				}

				// 使用字节缓冲区构建路径
				pathPos := 0
				pathLen := len(current.path)
				if pathLen > 0 {
					copy(pathBytes[pathPos:], current.path)
					pathPos += pathLen
				}
				pathBytes[pathPos] = '['
				pathPos++

				// 内联整数转字符串
				if i == 0 {
					pathBytes[pathPos] = '0'
					pathPos++
				} else {
					start := pathPos
					for num := i; num > 0; num /= 10 {
						pathBytes[pathPos] = byte('0' + num%10)
						pathPos++
					}
					// 反转数字
					for left, right := start, pathPos-1; left < right; left, right = left+1, right-1 {
						pathBytes[left], pathBytes[right] = pathBytes[right], pathBytes[left]
					}
				}

				pathBytes[pathPos] = ']'
				pathPos++

				newPath := string(pathBytes[:pathPos])
				stack = append(stack, walkItem{
					node: value,
					path: newPath,
				})
			}
		}
	}
}

// formatInt 优化的整数转字符串函数，避免fmt.Sprintf的开销
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	var buf [10]byte // 足够存储32位整数
	i := len(buf)

	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}

	return string(buf[i:])
}

// ===== 便捷查找方法 =====

// FindByPath 根据路径查找节点（支持深层嵌套）
// path 支持格式如: "data.notes[0].comments_count"
func (n Node) FindByPath(path string) Node {
	return n.GetPath(path)
}

// HasKey 检查对象是否包含指定键
func (n Node) HasKey(key string) bool {
	if n.typ != 'o' {
		return false
	}
	return n.Get(key).Exists()
}

// GetKeyValue 获取对象中指定键的值，如果不存在返回指定的默认值
func (n Node) GetKeyValue(key string, defaultValue Node) Node {
	if n.typ != 'o' {
		return defaultValue
	}
	value := n.Get(key)
	if !value.Exists() {
		return defaultValue
	}
	return value
}

// ===== 统计和分析方法 =====

// CountIf 统计数组中满足条件的元素个数
func (n Node) CountIf(predicate func(index int, value Node) bool) int {
	if n.typ != 'a' || predicate == nil {
		return 0
	}

	count := 0
	n.ArrayForEach(func(index int, value Node) bool {
		if predicate(index, value) {
			count++
		}
		return true
	})
	return count
}

// AllMatch 检查数组中是否所有元素都满足条件
func (n Node) AllMatch(predicate func(index int, value Node) bool) bool {
	if n.typ != 'a' || predicate == nil {
		return false
	}

	allMatch := true
	n.ArrayForEach(func(index int, value Node) bool {
		if !predicate(index, value) {
			allMatch = false
			return false // 提前终止
		}
		return true
	})
	return allMatch
}

// AnyMatch 检查数组中是否有任何元素满足条件
func (n Node) AnyMatch(predicate func(index int, value Node) bool) bool {
	if n.typ != 'a' || predicate == nil {
		return false
	}

	_, _, found := n.FindInArray(predicate)
	return found
}
