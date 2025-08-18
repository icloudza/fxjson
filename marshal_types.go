package fxjson

import (
	"reflect"
	"sort"
)

// marshalStruct 序列化结构体
func marshalStruct(buf *Buffer, rv reflect.Value, opts SerializeOptions, depth int) error {
	structType := rv.Type()
	typeInfo := getTypeInfo(structType)

	buf.WriteByte('{')

	written := false
	indent := opts.Indent
	hasIndent := indent != ""

	if hasIndent {
		depth++
	}

	for _, field := range typeInfo.fields {
		fieldValue := rv.Field(field.index)

		// 处理omitempty
		if field.omitEmpty && isEmptyValue(fieldValue) {
			continue
		}

		// 全局omitEmpty选项
		if opts.OmitEmpty && isEmptyValue(fieldValue) {
			continue
		}

		if written {
			buf.WriteByte(',')
		}

		if hasIndent {
			buf.WriteByte('\n')
			writeIndent(buf, indent, depth)
		}

		// 写入键
		writeString(buf, field.jsonName, opts.EscapeHTML)
		buf.WriteByte(':')

		if hasIndent {
			buf.WriteByte(' ')
		}

		// 写入值
		if err := marshalValue(buf, fieldValue, opts, depth); err != nil {
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

// fastMarshalStruct 快速序列化结构体
func fastMarshalStruct(buf *Buffer, rv reflect.Value) {
	structType := rv.Type()
	typeInfo := getTypeInfo(structType)

	buf.WriteByte('{')
	written := false

	for _, field := range typeInfo.fields {
		fieldValue := rv.Field(field.index)

		// 跳过空值（如果设置了omitempty）
		if field.omitEmpty && isEmptyValue(fieldValue) {
			continue
		}

		if written {
			buf.WriteByte(',')
		}

		// 写入键
		writeStringFast(buf, field.jsonName)
		buf.WriteByte(':')

		// 写入值
		fastMarshalValue(buf, fieldValue)
		written = true
	}

	buf.WriteByte('}')
}

// marshalSlice 序列化切片/数组
func marshalSlice(buf *Buffer, rv reflect.Value, opts SerializeOptions, depth int) error {
	length := rv.Len()

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

		if err := marshalValue(buf, rv.Index(i), opts, depth); err != nil {
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

// fastMarshalSlice 快速序列化切片/数组
func fastMarshalSlice(buf *Buffer, rv reflect.Value) {
	length := rv.Len()

	buf.WriteByte('[')

	for i := 0; i < length; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		fastMarshalValue(buf, rv.Index(i))
	}

	buf.WriteByte(']')
}

// marshalMap 序列化Map
func marshalMap(buf *Buffer, rv reflect.Value, opts SerializeOptions, depth int) error {
	if rv.IsNil() {
		buf.WriteString("null")
		return nil
	}

	keys := rv.MapKeys()

	// 排序键（如果启用）
	if opts.SortKeys {
		sortMapKeys(keys)
	}

	buf.WriteByte('{')

	written := false
	indent := opts.Indent
	hasIndent := indent != ""

	if hasIndent && len(keys) > 0 {
		depth++
	}

	for _, key := range keys {
		value := rv.MapIndex(key)

		// 处理omitempty
		if opts.OmitEmpty && isEmptyValue(value) {
			continue
		}

		if written {
			buf.WriteByte(',')
		}

		if hasIndent {
			buf.WriteByte('\n')
			writeIndent(buf, indent, depth)
		}

		// 写入键（必须是字符串）
		keyStr := getStringFromValue(key)
		writeString(buf, keyStr, opts.EscapeHTML)
		buf.WriteByte(':')

		if hasIndent {
			buf.WriteByte(' ')
		}

		// 写入值
		if err := marshalValue(buf, value, opts, depth); err != nil {
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

// fastMarshalMap 快速序列化Map
func fastMarshalMap(buf *Buffer, rv reflect.Value) {
	if rv.IsNil() {
		buf.WriteString("null")
		return
	}

	keys := rv.MapKeys()
	buf.WriteByte('{')

	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}

		// 写入键
		keyStr := getStringFromValue(key)
		writeStringFast(buf, keyStr)
		buf.WriteByte(':')

		// 写入值
		fastMarshalValue(buf, rv.MapIndex(key))
	}

	buf.WriteByte('}')
}

// writeIndent 写入缩进
func writeIndent(buf *Buffer, indent string, depth int) {
	for i := 0; i < depth; i++ {
		buf.WriteString(indent)
	}
}

// isEmptyValue 检查值是否为空
func isEmptyValue(rv reflect.Value) bool {
	if !rv.IsValid() {
		return true
	}

	switch rv.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return rv.IsNil()
	}
	return false
}

// getStringFromValue 从反射值获取字符串
func getStringFromValue(rv reflect.Value) string {
	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := rv.Int()
		return int64ToString(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := rv.Uint()
		return uint64ToString(n)
	case reflect.Float32, reflect.Float64:
		return floatToString(rv.Float())
	default:
		return rv.String()
	}
}

// int64ToString 整数转字符串（优化版本）
func int64ToString(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	// 计算位数
	digits := 0
	temp := n
	for temp > 0 {
		temp /= 10
		digits++
	}

	if negative {
		digits++
	}

	buf := make([]byte, digits)
	i := digits - 1

	// 填充数字
	for n > 0 {
		buf[i] = byte('0' + n%10)
		n /= 10
		i--
	}

	if negative {
		buf[0] = '-'
	}

	return string(buf)
}

// uint64ToString 无符号整数转字符串
func uint64ToString(n uint64) string {
	if n == 0 {
		return "0"
	}

	// 计算位数
	digits := 0
	temp := n
	for temp > 0 {
		temp /= 10
		digits++
	}

	buf := make([]byte, digits)
	i := digits - 1

	// 填充数字
	for n > 0 {
		buf[i] = byte('0' + n%10)
		n /= 10
		i--
	}

	return string(buf)
}

// floatToString 浮点数转字符串（简化版本）
func floatToString(f float64) string {
	// 对于map键，使用简单的转换
	return int64ToString(int64(f))
}

// sortMapKeys 排序map键
func sortMapKeys(keys []reflect.Value) {
	if len(keys) < 2 {
		return
	}

	// 检查键类型
	firstKey := keys[0]
	switch firstKey.Kind() {
	case reflect.String:
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Int() < keys[j].Int()
		})
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Uint() < keys[j].Uint()
		})
	case reflect.Float32, reflect.Float64:
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Float() < keys[j].Float()
		})
	}
}

// MarshalStruct 专门用于结构体序列化的优化函数
func MarshalStruct(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)

	// 处理指针
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return []byte("null"), nil
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return Marshal(v)
	}

	buf := getBuffer()
	defer putBuffer(buf)

	if err := marshalStruct(buf, rv, DefaultSerializeOptions, 0); err != nil {
		return nil, err
	}

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result, nil
}

// MarshalStructFast 极速结构体序列化
func MarshalStructFast(v interface{}) []byte {
	rv := reflect.ValueOf(v)

	// 处理指针
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return []byte("null")
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return FastMarshal(v)
	}

	buf := getBuffer()
	defer putBuffer(buf)

	fastMarshalStruct(buf, rv)

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result
}

// MarshalSlice 专门用于切片序列化的优化函数
func MarshalSlice(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return Marshal(v)
	}

	buf := getBuffer()
	defer putBuffer(buf)

	if err := marshalSlice(buf, rv, DefaultSerializeOptions, 0); err != nil {
		return nil, err
	}

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result, nil
}

// MarshalMap 专门用于Map序列化的优化函数
func MarshalMap(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Map {
		return Marshal(v)
	}

	buf := getBuffer()
	defer putBuffer(buf)

	if err := marshalMap(buf, rv, DefaultSerializeOptions, 0); err != nil {
		return nil, err
	}

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result, nil
}
