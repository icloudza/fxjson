package fxjson

import (
	"errors"
	"reflect"
	"sync"
	"unsafe"
)

const (
	maxInt64U = uint64(9223372036854775807)  // 2^63-1
	minInt64U = uint64(9223372036854775808)  // -(min int64) 的绝对值
	maxUint64 = uint64(18446744073709551615) // 2^64-1
)

type Node struct {
	raw   []byte
	start int
	end   int
	typ   byte
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
	key := arrKey{data: dataPtr(n.raw), s: n.start, e: n.end}
	if v, ok := arrIdxCache.Load(key); ok {
		return v.([]int)
	}
	data := n.raw
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

// ===== From / 基本访问 =====

func FromBytes(b []byte) Node {
	if len(b) == 0 {
		return Node{}
	}
	start, end := 0, len(b)
	for start < end && b[start] <= ' ' {
		start++
	}
	if start >= end {
		return Node{}
	}

	var typ byte
	switch b[start] {
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
	end = skipValueFast(b, start, end)
	return Node{raw: b, start: start, end: end, typ: typ}
}

func (n Node) Get(path string) Node {
	if len(path) == 0 || len(n.raw) == 0 {
		return Node{}
	}
	for i := 0; i < len(path); i++ {
		if path[i] == '.' || path[i] == '[' {
			return n.GetByPath(path)
		}
	}
	if n.typ != 'o' {
		return Node{}
	}
	keyData := unsafe.StringData(path)
	keyLen := len(path)
	pos := findObjectField(n.raw, n.start+1, n.end, keyData, 0, keyLen)
	if pos < 0 {
		return Node{}
	}
	return parseValueAt(n.raw, pos, n.end)
}

func (n Node) GetByPath(path string) Node {
	if len(n.raw) == 0 || len(path) == 0 {
		return Node{}
	}
	data := n.raw
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

	return parseValueAt(data, pos, end)
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

// 借助全局缓存，O(1) 取第 i 个元素起点；保持值接收器以支持链式

func (n Node) Index(i int) Node {
	offs := buildArrOffsetsCached(n)
	if i < 0 || i >= len(offs) {
		return Node{}
	}
	pos := offs[i]
	end := skipValueFast(n.raw, pos, n.end)
	return Node{raw: n.raw, start: pos, end: end, typ: detectType(n.raw[pos])}
}

// ===== 跳值 / 解析 =====
func skipValueFast(data []byte, pos int, end int) int {
	if pos >= end {
		return pos
	}
	switch data[pos] {
	case '"':
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
		pos++
		depth := 1
		for pos < end && depth > 0 {
			if data[pos] == '"' {
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
		pos++
		depth := 1
		for pos < end && depth > 0 {
			if data[pos] == '"' {
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
		return pos + 4
	case 'f':
		return pos + 5
	case 'n':
		return pos + 4
	default:
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

func (n Node) Int() (int64, error) {
	if n.typ != 'n' || n.start >= n.end {
		return 0, errors.New("not a number")
	}
	data := n.raw[n.start:n.end]
	if len(data) == 0 {
		return 0, errors.New("empty number")
	}
	i := 0
	neg := false
	if data[0] == '-' {
		neg = true
		i++
		if i >= len(data) {
			return 0, errors.New("invalid number")
		}
	}
	var val uint64
	for ; i < len(data); i++ {
		c := data[i]
		if c < '0' || c > '9' {
			return 0, errors.New("not an integer")
		}
		d := uint64(c - '0')
		if !neg {
			if val > (maxInt64U-d)/10 {
				return 0, errors.New("int64 overflow")
			}
		} else {
			if val > (minInt64U-d)/10 {
				return 0, errors.New("int64 overflow")
			}
		}
		val = val*10 + d
	}
	if neg {
		return -int64(val), nil
	}
	return int64(val), nil
}

func (n Node) Uint() (uint64, error) {
	if n.typ != 'n' || n.start >= n.end {
		return 0, errors.New("not a number")
	}
	data := n.raw[n.start:n.end]
	if len(data) == 0 {
		return 0, errors.New("empty number")
	}
	if data[0] == '-' {
		return 0, errors.New("negative to uint")
	}
	var val uint64
	for i := 0; i < len(data); i++ {
		c := data[i]
		if c < '0' || c > '9' {
			return 0, errors.New("not an unsigned integer")
		}
		d := uint64(c - '0')
		if val > (maxUint64-d)/10 {
			return 0, errors.New("uint64 overflow")
		}
		val = val*10 + d
	}
	return val, nil
}

func (n Node) Float() float64 {
	if n.typ != 'n' || n.start >= n.end {
		return 0
	}
	data := n.raw[n.start:n.end]
	if len(data) == 0 {
		return 0
	}
	i := 0
	neg := false
	if data[i] == '-' {
		neg = true
		i++
		if i >= len(data) {
			return 0
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
			return 0
		}
		expNeg := false
		if data[i] == '+' || data[i] == '-' {
			expNeg = data[i] == '-'
			i++
		}
		if i >= len(data) || data[i] < '0' || data[i] > '9' {
			return 0
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

func (n Node) Bool() (bool, error) {
	if n.typ != 'b' || n.start >= n.end {
		return false, errors.New("not a bool")
	}
	data := n.raw[n.start:n.end]
	if len(data) == 4 && data[0] == 't' && data[1] == 'r' && data[2] == 'u' && data[3] == 'e' {
		return true, nil
	}
	if len(data) == 5 && data[0] == 'f' && data[1] == 'a' && data[2] == 'l' && data[3] == 's' && data[4] == 'e' {
		return false, nil
	}
	return false, errors.New("invalid bool")
}

func (n Node) Exists() bool { return len(n.raw) > 0 && n.start >= 0 && n.end > n.start }
func (n Node) IsNull() bool { return n.typ == 'l' }

func (n Node) NumStr() string {
	if n.typ != 'n' || n.start >= n.end {
		return ""
	}
	return unsafe.String(&n.raw[n.start], n.end-n.start)
}

func (n Node) Raw() []byte {
	if n.start >= 0 && n.end <= len(n.raw) && n.start < n.end {
		return n.raw[n.start:n.end]
	}
	return nil
}

// ===== 统计 / Keys =====

func (n Node) Len() int {
	if n.typ == 'a' {
		pos := n.start
		end := n.end
		for pos < end && n.raw[pos] != '[' {
			pos++
		}
		if pos >= end {
			return 0
		}
		pos++
		count := 0
		for pos < end {
			for pos < end && n.raw[pos] <= ' ' {
				pos++
			}
			if pos >= end || n.raw[pos] == ']' {
				break
			}
			count++
			pos = skipValueFast(n.raw, pos, end)
			for pos < end && n.raw[pos] <= ' ' {
				pos++
			}
			if pos < end && n.raw[pos] == ',' {
				pos++
			}
		}
		return count
	}
	if n.typ == 'o' {
		pos := n.start
		end := n.end
		for pos < end && n.raw[pos] != '{' {
			pos++
		}
		if pos >= end {
			return 0
		}
		pos++
		count := 0
		for pos < end {
			for pos < end && n.raw[pos] <= ' ' {
				pos++
			}
			if pos >= end || n.raw[pos] == '}' {
				break
			}
			if n.raw[pos] != '"' {
				return count
			}
			pos++
			for pos < end && n.raw[pos] != '"' {
				if n.raw[pos] == '\\' {
					pos++
				}
				pos++
			}
			pos++
			for pos < end && n.raw[pos] <= ' ' {
				pos++
			}
			if pos >= end || n.raw[pos] != ':' {
				return count
			}
			pos++
			for pos < end && n.raw[pos] <= ' ' {
				pos++
			}
			pos = skipValueFast(n.raw, pos, end)
			count++
			for pos < end && n.raw[pos] <= ' ' {
				pos++
			}
			if pos < end && n.raw[pos] == ',' {
				pos++
			}
		}
		return count
	}
	return 0
}

func (n Node) Keys() [][]byte {
	if n.typ != 'o' {
		return nil
	}
	var keys [][]byte
	pos := n.start
	end := n.end
	for pos < end && n.raw[pos] != '{' {
		pos++
	}
	if pos >= end {
		return nil
	}
	pos++
	for pos < end {
		for pos < end && n.raw[pos] <= ' ' {
			pos++
		}
		if pos >= end || n.raw[pos] == '}' {
			break
		}
		if n.raw[pos] != '"' {
			return keys
		}
		pos++
		keyStart := pos
		for pos < end && n.raw[pos] != '"' {
			if n.raw[pos] == '\\' {
				pos++
			}
			pos++
		}
		keyEnd := pos
		keys = append(keys, n.raw[keyStart:keyEnd])
		pos++
		for pos < end && n.raw[pos] <= ' ' {
			pos++
		}
		if pos >= end || n.raw[pos] != ':' {
			return keys
		}
		pos++
		for pos < end && n.raw[pos] <= ' ' {
			pos++
		}
		pos = skipValueFast(n.raw, pos, end)
		for pos < end && n.raw[pos] <= ' ' {
			pos++
		}
		if pos < end && n.raw[pos] == ',' {
			pos++
		}
	}
	return keys
}

// ===== 解码 =====

func (n Node) RawString() string {
	if n.start >= 0 && n.end <= len(n.raw) && n.start < n.end {
		return unsafe.String(&n.raw[n.start], n.end-n.start)
	}
	return ""
}

func (n Node) Decode(v any) error {
	if !n.Exists() {
		return errors.New("node does not exist")
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a non-nil pointer")
	}
	val, _, err := fastDecode(n.raw, n.start, n.end)
	if err != nil {
		return err
	}
	rv.Elem().Set(reflect.ValueOf(val))
	return nil
}

func fastDecode(buf []byte, start, end int) (any, int, error) {
	if start >= end {
		return nil, start, errors.New("empty node")
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
				return nil, i, errors.New("invalid object key")
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
		return 0, errors.New("empty number")
	}
	i := 0
	neg := false
	if data[0] == '-' {
		neg = true
		i++
		if i >= len(data) {
			return 0, errors.New("invalid number")
		}
	}
	var val uint64
	for ; i < len(data); i++ {
		c := data[i]
		if c < '0' || c > '9' {
			return 0, errors.New("not an integer")
		}
		d := uint64(c - '0')
		if !neg {
			if val > (maxInt64U-d)/10 {
				return 0, errors.New("int64 overflow")
			}
		} else {
			if val > (minInt64U-d)/10 {
				return 0, errors.New("int64 overflow")
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

// ===== tool =====
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
