package fxjson

import (
	"fmt"
	"reflect"
	"time"
)

// CompactJSON 压缩JSON字符串（移除空白字符）
func CompactJSON(src []byte) []byte {
	buf := getBuffer()
	defer putBuffer(buf)

	inString := false
	escaped := false

	for i := 0; i < len(src); i++ {
		c := src[i]

		if inString {
			buf.WriteByte(c)
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
		} else {
			switch c {
			case '"':
				inString = true
				buf.WriteByte(c)
			case ' ', '\t', '\n', '\r':
				// 跳过空白字符
				continue
			default:
				buf.WriteByte(c)
			}
		}
	}

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result
}

// PrettyJSON 美化JSON字符串（添加缩进和换行）
func PrettyJSON(src []byte) []byte {
	return PrettyJSONWithIndent(src, "  ")
}

// PrettyJSONWithIndent 使用指定缩进美化JSON字符串
func PrettyJSONWithIndent(src []byte, indent string) []byte {
	buf := getBuffer()
	defer putBuffer(buf)

	inString := false
	escaped := false
	depth := 0

	for i := 0; i < len(src); i++ {
		c := src[i]

		if inString {
			buf.WriteByte(c)
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
		} else {
			switch c {
			case '"':
				inString = true
				buf.WriteByte(c)
			case '{', '[':
				buf.WriteByte(c)
				depth++
				// 检查下一个字符是否是结束符
				if i+1 < len(src) {
					next := src[i+1]
					for next == ' ' || next == '\t' || next == '\n' || next == '\r' {
						i++
						if i+1 >= len(src) {
							break
						}
						next = src[i+1]
					}
					if next != '}' && next != ']' {
						buf.WriteByte('\n')
						writeIndent(buf, indent, depth)
					}
				}
			case '}', ']':
				// 检查前一个字符是否是开始符
				prevChar := byte(0)
				if len(buf.buf) > 0 {
					prevChar = buf.buf[len(buf.buf)-1]
				}
				depth--
				if prevChar != '{' && prevChar != '[' {
					buf.WriteByte('\n')
					writeIndent(buf, indent, depth)
				}
				buf.WriteByte(c)
			case ',':
				buf.WriteByte(c)
				buf.WriteByte('\n')
				writeIndent(buf, indent, depth)
			case ':':
				buf.WriteByte(c)
				buf.WriteByte(' ')
			case ' ', '\t', '\n', '\r':
				// 跳过现有的空白字符
				continue
			default:
				buf.WriteByte(c)
			}
		}
	}

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result
}

// ValidateJSON 验证JSON格式是否正确
func ValidateJSON(data []byte) bool {
	node := FromBytes(data)
	return node.Exists()
}

// JSONSize 计算JSON数据大小（字节）
func JSONSize(v interface{}) int {
	if data, err := Marshal(v); err == nil {
		return len(data)
	}
	return 0
}

// EstimateJSONSize 估算JSON数据大小（不进行实际序列化）
func EstimateJSONSize(v interface{}) int {
	return estimateSize(reflect.ValueOf(v))
}

// estimateSize 估算反射值的JSON大小
func estimateSize(rv reflect.Value) int {
	if !rv.IsValid() {
		return 4 // "null"
	}

	// 处理指针
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return 4 // "null"
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			return 4 // "true"
		}
		return 5 // "false"

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := rv.Int()
		if n == 0 {
			return 1
		}
		size := 0
		if n < 0 {
			size = 1
			n = -n
		}
		for n > 0 {
			size++
			n /= 10
		}
		return size

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := rv.Uint()
		if n == 0 {
			return 1
		}
		size := 0
		for n > 0 {
			size++
			n /= 10
		}
		return size

	case reflect.Float32, reflect.Float64:
		return 20 // 估算浮点数长度

	case reflect.String:
		return rv.Len() + 2 // 字符串长度 + 两个引号

	case reflect.Slice, reflect.Array:
		size := 2 // []
		length := rv.Len()
		for i := 0; i < length; i++ {
			if i > 0 {
				size++ // 逗号
			}
			size += estimateSize(rv.Index(i))
		}
		return size

	case reflect.Map:
		if rv.IsNil() {
			return 4 // "null"
		}
		size := 2 // {}
		keys := rv.MapKeys()
		for i, key := range keys {
			if i > 0 {
				size++ // 逗号
			}
			size += estimateSize(key) + 1 // 键 + 冒号
			size += estimateSize(rv.MapIndex(key))
		}
		return size

	case reflect.Struct:
		size := 2 // {}
		structType := rv.Type()
		typeInfo := getTypeInfo(structType)

		fieldCount := 0
		for _, field := range typeInfo.fields {
			fieldValue := rv.Field(field.index)
			if field.omitEmpty && isEmptyValue(fieldValue) {
				continue
			}

			if fieldCount > 0 {
				size++ // 逗号
			}

			size += len(field.jsonName) + 3 // 字段名 + 引号 + 冒号
			size += estimateSize(fieldValue)
			fieldCount++
		}
		return size

	default:
		return 10 // 默认估算
	}
}

// JSONDepth 计算JSON数据的最大嵌套深度
func JSONDepth(data []byte) int {
	node := FromBytes(data)
	return calculateDepth(node, 0)
}

// calculateDepth 计算节点深度
func calculateDepth(node Node, currentDepth int) int {
	if !node.Exists() {
		return currentDepth
	}

	maxDepth := currentDepth

	switch node.Type() {
	case 'o':
		node.ForEach(func(key string, value Node) bool {
			depth := calculateDepth(value, currentDepth+1)
			if depth > maxDepth {
				maxDepth = depth
			}
			return true
		})
	case 'a':
		for i := 0; i < node.Len(); i++ {
			depth := calculateDepth(node.Index(i), currentDepth+1)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	return maxDepth
}

// MarshalTime 序列化时间（RFC3339格式）
func MarshalTime(t time.Time) []byte {
	return []byte(`"` + t.Format(time.RFC3339) + `"`)
}

// MarshalTimeUnix 序列化时间（Unix时间戳）
func MarshalTimeUnix(t time.Time) []byte {
	buf := getBuffer()
	defer putBuffer(buf)

	writeInt(buf, t.Unix())

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result
}

// MarshalDuration 序列化时间间隔（纳秒）
func MarshalDuration(d time.Duration) []byte {
	buf := getBuffer()
	defer putBuffer(buf)

	writeInt(buf, int64(d))

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result
}

// MarshalBinary 序列化二进制数据（Base64编码）
func MarshalBinary(data []byte) []byte {
	// 简化的Base64编码
	encoded := base64Encode(data)
	result := make([]byte, len(encoded)+2)
	result[0] = '"'
	copy(result[1:], encoded)
	result[len(result)-1] = '"'
	return result
}

// base64Encode 简化的Base64编码
func base64Encode(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}

	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	n := len(src)
	encoded := make([]byte, (n+2)/3*4)

	si, ei := 0, 0
	for si < n-2 {
		val := uint32(src[si])<<16 | uint32(src[si+1])<<8 | uint32(src[si+2])

		encoded[ei] = base64Table[val>>18&0x3F]
		encoded[ei+1] = base64Table[val>>12&0x3F]
		encoded[ei+2] = base64Table[val>>6&0x3F]
		encoded[ei+3] = base64Table[val&0x3F]

		si += 3
		ei += 4
	}

	remain := n - si
	if remain > 0 {
		val := uint32(src[si]) << 16
		if remain == 2 {
			val |= uint32(src[si+1]) << 8
		}

		encoded[ei] = base64Table[val>>18&0x3F]
		encoded[ei+1] = base64Table[val>>12&0x3F]

		if remain == 2 {
			encoded[ei+2] = base64Table[val>>6&0x3F]
		} else {
			encoded[ei+2] = '='
		}
		encoded[ei+3] = '='
	}

	return encoded
}

// StructToMap 将结构体转换为map[string]interface{}
func StructToMap(v interface{}) (map[string]interface{}, error) {
	rv := reflect.ValueOf(v)

	// 处理指针
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", rv.Kind())
	}

	result := make(map[string]interface{})
	structType := rv.Type()
	typeInfo := getTypeInfo(structType)

	for _, field := range typeInfo.fields {
		fieldValue := rv.Field(field.index)

		if field.omitEmpty && isEmptyValue(fieldValue) {
			continue
		}

		value := fieldValue.Interface()
		result[field.jsonName] = value
	}

	return result, nil
}

// MapToStruct 将map转换为结构体
func MapToStruct(m map[string]interface{}, v interface{}) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer")
	}

	if rv.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("v must point to a struct")
	}

	structType := elem.Type()
	typeInfo := getTypeInfo(structType)

	for _, field := range typeInfo.fields {
		if value, exists := m[field.jsonName]; exists {
			fieldValue := elem.Field(field.index)
			if fieldValue.CanSet() {
				valueRV := reflect.ValueOf(value)
				if valueRV.Type().AssignableTo(fieldValue.Type()) {
					fieldValue.Set(valueRV)
				}
			}
		}
	}

	return nil
}
