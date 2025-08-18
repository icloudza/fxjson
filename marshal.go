package fxjson

import (
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

// SerializeOptions 序列化选项
type SerializeOptions struct {
	Indent          string // 缩进字符串，空字符串表示压缩模式
	EscapeHTML      bool   // 是否转义HTML字符 (<, >, &)
	SortKeys        bool   // 是否对对象键进行排序
	OmitEmpty       bool   // 是否忽略空值
	FloatPrecision  int    // 浮点数精度，-1表示默认
	UseNumberString bool   // 大数字是否用字符串表示
}

// DefaultSerializeOptions 默认序列化选项（压缩模式）
var DefaultSerializeOptions = SerializeOptions{
	Indent:          "",
	EscapeHTML:      false,
	SortKeys:        false,
	OmitEmpty:       false,
	FloatPrecision:  -1,
	UseNumberString: false,
}

// PrettySerializeOptions 美化打印选项
var PrettySerializeOptions = SerializeOptions{
	Indent:          "  ",
	EscapeHTML:      false,
	SortKeys:        true,
	OmitEmpty:       false,
	FloatPrecision:  -1,
	UseNumberString: false,
}

// Buffer 高性能字节缓冲区
type Buffer struct {
	buf []byte
}

// bufferPool 缓冲区池，减少内存分配
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &Buffer{buf: make([]byte, 0, 1024)}
	},
}

// getBuffer 从池中获取缓冲区
func getBuffer() *Buffer {
	return bufferPool.Get().(*Buffer)
}

// putBuffer 将缓冲区归还到池中
func putBuffer(buf *Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}

// Reset 重置缓冲区
func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
}

// Bytes 返回缓冲区字节切片
func (b *Buffer) Bytes() []byte {
	return b.buf
}

// String 返回缓冲区字符串
func (b *Buffer) String() string {
	return unsafe.String(unsafe.SliceData(b.buf), len(b.buf))
}

// WriteByte 写入单个字节
func (b *Buffer) WriteByte(c byte) {
	b.buf = append(b.buf, c)
}

// WriteString 写入字符串
func (b *Buffer) WriteString(s string) {
	b.buf = append(b.buf, s...)
}

// Write 写入字节切片
func (b *Buffer) Write(p []byte) {
	b.buf = append(b.buf, p...)
}

// Grow 扩展缓冲区容量
func (b *Buffer) Grow(n int) {
	if cap(b.buf)-len(b.buf) < n {
		newBuf := make([]byte, len(b.buf), len(b.buf)+n+1024)
		copy(newBuf, b.buf)
		b.buf = newBuf
	}
}

// fieldInfo 字段信息缓存
type fieldInfo struct {
	index       int
	name        string
	jsonName    string
	omitEmpty   bool
	isPointer   bool
	isInterface bool
	fieldType   reflect.Type
}

// typeInfo 类型信息缓存
type typeInfo struct {
	fields []fieldInfo
}

// typeCache 类型信息缓存
var typeCache sync.Map

// getTypeInfo 获取类型信息（带缓存）
func getTypeInfo(t reflect.Type) *typeInfo {
	if cached, ok := typeCache.Load(t); ok {
		return cached.(*typeInfo)
	}

	info := &typeInfo{
		fields: make([]fieldInfo, 0, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过未导出字段
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		jsonName := field.Name
		omitEmpty := false

		if jsonTag != "" {
			parts := parseJSONTag(jsonTag)
			if parts[0] != "" {
				jsonName = parts[0]
			}
			for _, part := range parts[1:] {
				if part == "omitempty" {
					omitEmpty = true
				}
			}
		}

		fieldType := field.Type
		isPointer := fieldType.Kind() == reflect.Ptr
		isInterface := fieldType.Kind() == reflect.Interface

		info.fields = append(info.fields, fieldInfo{
			index:       i,
			name:        field.Name,
			jsonName:    jsonName,
			omitEmpty:   omitEmpty,
			isPointer:   isPointer,
			isInterface: isInterface,
			fieldType:   fieldType,
		})
	}

	typeCache.Store(t, info)
	return info
}

// parseJSONTag 解析JSON标签
func parseJSONTag(tag string) []string {
	var parts []string
	var current []byte

	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' {
			parts = append(parts, string(current))
			current = current[:0]
		} else {
			current = append(current, tag[i])
		}
	}

	if len(current) > 0 {
		parts = append(parts, string(current))
	}

	return parts
}

// Marshal 将Go值序列化为JSON字节切片（压缩模式）
func Marshal(v interface{}) ([]byte, error) {
	return MarshalWithOptions(v, DefaultSerializeOptions)
}

// MarshalIndent 将Go值序列化为格式化的JSON字节切片
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	opts := PrettySerializeOptions
	opts.Indent = indent
	return MarshalWithOptions(v, opts)
}

// MarshalWithOptions 使用指定选项序列化
func MarshalWithOptions(v interface{}, opts SerializeOptions) ([]byte, error) {
	buf := getBuffer()
	defer putBuffer(buf)

	if err := marshalValue(buf, reflect.ValueOf(v), opts, 0); err != nil {
		return nil, err
	}

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result, nil
}

// MarshalToString 序列化为字符串（压缩模式）
func MarshalToString(v interface{}) (string, error) {
	return MarshalToStringWithOptions(v, DefaultSerializeOptions)
}

// MarshalToStringWithOptions 使用指定选项序列化为字符串
func MarshalToStringWithOptions(v interface{}, opts SerializeOptions) (string, error) {
	buf := getBuffer()
	defer putBuffer(buf)

	if err := marshalValue(buf, reflect.ValueOf(v), opts, 0); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// FastMarshal 极速序列化（最小开销）
func FastMarshal(v interface{}) []byte {
	buf := getBuffer()
	defer putBuffer(buf)

	// 跳过错误检查，追求极致性能
	fastMarshalValue(buf, reflect.ValueOf(v))

	result := make([]byte, len(buf.buf))
	copy(result, buf.buf)
	return result
}

// marshalValue 序列化反射值
func marshalValue(buf *Buffer, rv reflect.Value, opts SerializeOptions, depth int) error {
	if !rv.IsValid() {
		buf.WriteString("null")
		return nil
	}

	// 处理指针
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			buf.WriteString("null")
			return nil
		}
		rv = rv.Elem()
	}

	// 处理接口
	if rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			buf.WriteString("null")
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		writeInt(buf, rv.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		writeUint(buf, rv.Uint())

	case reflect.Float32, reflect.Float64:
		writeFloat(buf, rv.Float(), opts.FloatPrecision)

	case reflect.String:
		writeString(buf, rv.String(), opts.EscapeHTML)

	case reflect.Slice, reflect.Array:
		return marshalSlice(buf, rv, opts, depth)

	case reflect.Map:
		return marshalMap(buf, rv, opts, depth)

	case reflect.Struct:
		return marshalStruct(buf, rv, opts, depth)

	default:
		// 处理其他类型，转换为字符串
		writeString(buf, rv.String(), opts.EscapeHTML)
	}

	return nil
}

// fastMarshalValue 快速序列化（无错误检查）
func fastMarshalValue(buf *Buffer, rv reflect.Value) {
	if !rv.IsValid() {
		buf.WriteString("null")
		return
	}

	// 处理指针
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			buf.WriteString("null")
			return
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		writeInt(buf, rv.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		writeUint(buf, rv.Uint())

	case reflect.Float32, reflect.Float64:
		writeFloat(buf, rv.Float(), -1)

	case reflect.String:
		writeStringFast(buf, rv.String())

	case reflect.Slice, reflect.Array:
		fastMarshalSlice(buf, rv)

	case reflect.Map:
		fastMarshalMap(buf, rv)

	case reflect.Struct:
		fastMarshalStruct(buf, rv)

	default:
		buf.WriteString("null")
	}
}

// writeInt 写入整数
func writeInt(buf *Buffer, n int64) {
	if n == 0 {
		buf.WriteByte('0')
		return
	}

	if n < 0 {
		buf.WriteByte('-')
		n = -n
	}

	// 计算数字位数
	digits := 0
	temp := n
	for temp > 0 {
		temp /= 10
		digits++
	}

	// 预分配空间
	start := len(buf.buf)
	buf.Grow(digits)
	buf.buf = buf.buf[:start+digits]

	// 从右到左填充数字
	for i := digits - 1; i >= 0; i-- {
		buf.buf[start+i] = byte('0' + n%10)
		n /= 10
	}
}

// writeUint 写入无符号整数
func writeUint(buf *Buffer, n uint64) {
	if n == 0 {
		buf.WriteByte('0')
		return
	}

	// 计算数字位数
	digits := 0
	temp := n
	for temp > 0 {
		temp /= 10
		digits++
	}

	// 预分配空间
	start := len(buf.buf)
	buf.Grow(digits)
	buf.buf = buf.buf[:start+digits]

	// 从右到左填充数字
	for i := digits - 1; i >= 0; i-- {
		buf.buf[start+i] = byte('0' + n%10)
		n /= 10
	}
}

// writeFloat 写入浮点数
func writeFloat(buf *Buffer, f float64, precision int) {
	if f != f { // NaN
		buf.WriteString("null")
		return
	}

	if f > 1e20 || f < -1e20 {
		// 使用科学计数法
		buf.WriteString(strconv.FormatFloat(f, 'e', precision, 64))
	} else {
		if precision >= 0 {
			buf.WriteString(strconv.FormatFloat(f, 'f', precision, 64))
		} else {
			buf.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
		}
	}
}

// writeString 写入字符串（带转义）
func writeString(buf *Buffer, s string, escapeHTML bool) {
	buf.WriteByte('"')

	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		var escape string

		switch c {
		case '"':
			escape = `\"`
		case '\\':
			escape = `\\`
		case '\b':
			escape = `\b`
		case '\f':
			escape = `\f`
		case '\n':
			escape = `\n`
		case '\r':
			escape = `\r`
		case '\t':
			escape = `\t`
		case '<':
			if escapeHTML {
				escape = `\u003c`
			}
		case '>':
			if escapeHTML {
				escape = `\u003e`
			}
		case '&':
			if escapeHTML {
				escape = `\u0026`
			}
		default:
			if c < 0x20 {
				// 控制字符需要转义
				buf.WriteString(s[start:i])
				buf.WriteString(`\u00`)
				buf.WriteByte(hexDigits[c>>4])
				buf.WriteByte(hexDigits[c&0xF])
				start = i + 1
			}
			continue
		}

		if escape != "" {
			buf.WriteString(s[start:i])
			buf.WriteString(escape)
			start = i + 1
		}
	}

	buf.WriteString(s[start:])
	buf.WriteByte('"')
}

// writeStringFast 快速写入字符串（最小转义）
func writeStringFast(buf *Buffer, s string) {
	buf.WriteByte('"')

	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]

		switch c {
		case '"':
			buf.WriteString(s[start:i])
			buf.WriteString(`\"`)
			start = i + 1
		case '\\':
			buf.WriteString(s[start:i])
			buf.WriteString(`\\`)
			start = i + 1
		case '\n':
			buf.WriteString(s[start:i])
			buf.WriteString(`\n`)
			start = i + 1
		case '\r':
			buf.WriteString(s[start:i])
			buf.WriteString(`\r`)
			start = i + 1
		case '\t':
			buf.WriteString(s[start:i])
			buf.WriteString(`\t`)
			start = i + 1
		}
	}

	buf.WriteString(s[start:])
	buf.WriteByte('"')
}

// hexDigits 十六进制数字
var hexDigits = "0123456789abcdef"
