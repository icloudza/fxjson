package fxjson

import (
	"unsafe"
)

// Node 零分配 JSON 节点
type Node struct {
	raw   []byte
	start int
	end   int
	typ   byte
}

// NodeType 表示 JSON 节点类型
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

// FromBytes 最小化初始化
func FromBytes(b []byte) Node {
	if len(b) == 0 {
		return Node{}
	}
	return Node{raw: b, start: 0, end: len(b), typ: 'o'}
}

// GetByPath 极速路径访问 - 完全重写
func (n Node) GetByPath(path string) Node {
	if len(n.raw) == 0 || len(path) == 0 {
		return Node{}
	}

	data := n.raw
	dataLen := len(data)

	// 将路径转换为字节访问（零开销）
	pathData := unsafe.StringData(path)
	pathLen := len(path)

	// 当前位置
	pos := 0

	// 跳过前导空白
	for pos < dataLen && (data[pos] == ' ' || data[pos] == '\t' || data[pos] == '\n' || data[pos] == '\r') {
		pos++
	}

	// 必须以 { 开始
	if pos >= dataLen || data[pos] != '{' {
		return Node{}
	}
	pos++

	// 路径解析位置
	pathPos := 0

	// 处理每个路径段
	for pathPos < pathLen {
		// 提取路径段
		segStart := pathPos
		segLen := 0

		// 找到段的结束
		for pathPos < pathLen {
			c := *(*byte)(unsafe.Add(unsafe.Pointer(pathData), pathPos))
			if c == '.' || c == '[' {
				break
			}
			segLen++
			pathPos++
		}

		if segLen == 0 {
			// 处理数组索引
			if pathPos < pathLen && *(*byte)(unsafe.Add(unsafe.Pointer(pathData), pathPos)) == '[' {
				pathPos++ // 跳过 [
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

				// 查找数组中的第 idx 个元素
				pos = findArrayElement(data, pos, dataLen, idx)
				if pos < 0 {
					return Node{}
				}
			}
		} else {
			// 处理对象字段
			pos = findObjectField(data, pos, dataLen, pathData, segStart, segLen)
			if pos < 0 {
				return Node{}
			}
		}

		// 跳过路径中的点
		if pathPos < pathLen && *(*byte)(unsafe.Add(unsafe.Pointer(pathData), pathPos)) == '.' {
			pathPos++
		}
	}

	// 解析最终值
	return parseValueAt(data, pos, dataLen)
}

// findObjectField 在对象中查找指定字段
func findObjectField(data []byte, start int, end int, keyData *byte, keyStart int, keyLen int) int {
	pos := start

	for pos < end {
		// 跳过空白
		for pos < end && (data[pos] <= ' ') {
			pos++
		}

		if pos >= end || data[pos] == '}' {
			return -1
		}

		// 必须是引号
		if data[pos] != '"' {
			return -1
		}
		pos++

		// 比较字段名
		fieldStart := pos
		match := true

		// 快速比较
		if pos+keyLen <= end && data[pos+keyLen] == '"' {
			for i := 0; i < keyLen; i++ {
				if data[fieldStart+i] != *(*byte)(unsafe.Add(unsafe.Pointer(keyData), keyStart+i)) {
					match = false
					break
				}
			}

			if match {
				// 找到匹配字段
				pos += keyLen + 1 // 跳过字段名和引号

				// 跳过冒号前的空白
				for pos < end && data[pos] <= ' ' {
					pos++
				}

				if pos >= end || data[pos] != ':' {
					return -1
				}
				pos++

				// 跳过值前的空白
				for pos < end && data[pos] <= ' ' {
					pos++
				}

				return pos
			}
		}

		// 跳过不匹配的字段名
		for pos < end && data[pos] != '"' {
			if data[pos] == '\\' {
				pos++
			}
			pos++
		}
		pos++ // 跳过引号

		// 跳过冒号
		for pos < end && data[pos] != ':' {
			pos++
		}
		pos++

		// 跳过空白
		for pos < end && data[pos] <= ' ' {
			pos++
		}

		// 跳过值
		pos = skipValueFast(data, pos, end)

		// 跳过逗号
		if pos < end && data[pos] == ',' {
			pos++
		}
	}

	return -1
}

// findArrayElement 查找数组中的第 n 个元素
func findArrayElement(data []byte, start int, end int, index int) int {
	pos := start

	// 跳过空白
	for pos < end && data[pos] <= ' ' {
		pos++
	}

	if pos >= end || data[pos] != '[' {
		return -1
	}
	pos++

	currentIndex := 0

	for pos < end {
		// 跳过空白
		for pos < end && data[pos] <= ' ' {
			pos++
		}

		if pos >= end || data[pos] == ']' {
			return -1
		}

		if currentIndex == index {
			return pos
		}

		// 跳过当前元素
		pos = skipValueFast(data, pos, end)
		currentIndex++

		// 跳过逗号
		if pos < end && data[pos] == ',' {
			pos++
		}
	}

	return -1
}

// skipValueFast 超快速跳过任意值
func skipValueFast(data []byte, pos int, end int) int {
	if pos >= end {
		return pos
	}

	switch data[pos] {
	case '"':
		// 字符串
		pos++
		for pos < end {
			if data[pos] == '"' {
				return pos + 1
			}
			if data[pos] == '\\' {
				pos += 2
			} else {
				pos++
			}
		}

	case '{':
		// 对象
		pos++
		depth := 1
		for pos < end && depth > 0 {
			if data[pos] == '"' {
				// 跳过字符串
				pos++
				for pos < end && data[pos] != '"' {
					if data[pos] == '\\' {
						pos++
					}
					pos++
				}
			} else if data[pos] == '{' {
				depth++
			} else if data[pos] == '}' {
				depth--
			}
			pos++
		}
		return pos

	case '[':
		// 数组
		pos++
		depth := 1
		for pos < end && depth > 0 {
			if data[pos] == '"' {
				// 跳过字符串
				pos++
				for pos < end && data[pos] != '"' {
					if data[pos] == '\\' {
						pos++
					}
					pos++
				}
			} else if data[pos] == '[' {
				depth++
			} else if data[pos] == ']' {
				depth--
			}
			pos++
		}
		return pos

	case 't':
		// true
		return pos + 4

	case 'f':
		// false
		return pos + 5

	case 'n':
		// null
		return pos + 4

	default:
		// 数字
		if data[pos] == '-' {
			pos++
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
	}

	return pos
}

// parseValueAt 解析指定位置的值
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

// String 返回字符串值
func (n Node) String() string {
	if n.typ != 's' || n.start+1 >= n.end {
		return ""
	}
	bytes := n.raw[n.start+1 : n.end-1]
	if len(bytes) == 0 {
		return ""
	}
	return unsafe.String(&bytes[0], len(bytes))
}

// Get 获取对象字段
func (n Node) Get(key string) Node {
	if n.typ != 'o' {
		return Node{}
	}

	keyData := unsafe.StringData(key)
	keyLen := len(key)

	pos := findObjectField(n.raw, n.start+1, n.end, keyData, 0, keyLen)
	if pos < 0 {
		return Node{}
	}

	return parseValueAt(n.raw, pos, n.end)
}

// Type 返回节点类型
func (n Node) Type() byte {
	return n.typ
}

// Kind 返回节点类型（语义化）
func (n Node) Kind() NodeType {
	return NodeType(n.typ)
}
