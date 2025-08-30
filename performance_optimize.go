package fxjson

import (
	"sync"
	"unsafe"
)

// ObjectKeyCache 对象键位置缓存
type ObjectKeyCache struct {
	mu      sync.RWMutex
	cache   map[uintptr]map[string]int // dataPtr -> key -> position
	maxSize int
}

var defaultKeyCache = &ObjectKeyCache{
	cache:   make(map[uintptr]map[string]int),
	maxSize: 1000, // 最多缓存1000个对象
}

// findObjectFieldFast 快速对象字段查找，带缓存
func findObjectFieldFast(data []byte, start int, end int, key string) int {
	// 对于大对象，使用缓存加速查找
	if len(data) > 10000 { // 只对大对象启用缓存
		dataPtr := dataPtr(data)
		if dataPtr != 0 {
			defaultKeyCache.mu.RLock()
			if objCache, exists := defaultKeyCache.cache[dataPtr]; exists {
				if pos, found := objCache[key]; found {
					defaultKeyCache.mu.RUnlock()
					return pos
				}
			}
			defaultKeyCache.mu.RUnlock()
		}
	}
	
	// 原始查找逻辑
	keyData := unsafe.StringData(key)
	pos := findObjectField(data, start, end, keyData, 0, len(key))
	
	// 缓存结果（仅对大对象）
	if len(data) > 10000 && pos >= 0 {
		dataPtr := dataPtr(data)
		if dataPtr != 0 {
			defaultKeyCache.mu.Lock()
			if len(defaultKeyCache.cache) < defaultKeyCache.maxSize {
				if _, exists := defaultKeyCache.cache[dataPtr]; !exists {
					defaultKeyCache.cache[dataPtr] = make(map[string]int)
				}
				defaultKeyCache.cache[dataPtr][key] = pos
			}
			defaultKeyCache.mu.Unlock()
		}
	}
	
	return pos
}

// GetFast 优化版本的Get方法
func (n Node) GetFast(path string) Node {
	if len(path) == 0 || len(n.raw) == 0 {
		return Node{}
	}
	for i := 0; i < len(path); i++ {
		if path[i] == '.' || path[i] == '[' {
			return n.GetPath(path) // 复杂路径仍使用原方法
		}
	}
	if n.typ != 'o' {
		return Node{}
	}

	data := n.getWorkingData()
	if len(path) == 0 || len(data) == 0 {
		return Node{}
	}
	
	// 使用快速查找
	pos := findObjectFieldFast(data, n.start+1, n.end, path)
	if pos < 0 {
		return Node{}
	}
	return parseValueAtWithData(data, pos, n.end, n.expanded)
}

// ClearKeyCache 清除键缓存
func ClearKeyCache() {
	defaultKeyCache.mu.Lock()
	defaultKeyCache.cache = make(map[uintptr]map[string]int)
	defaultKeyCache.mu.Unlock()
}

// BatchObjectAccess 批量对象访问优化
type BatchObjectAccess struct {
	node Node
	keys []string
}

// NewBatchAccess 创建批量访问器
func (n Node) NewBatchAccess(keys []string) *BatchObjectAccess {
	return &BatchObjectAccess{
		node: n,
		keys: keys,
	}
}

// GetAll 批量获取所有键的值
func (b *BatchObjectAccess) GetAll() map[string]Node {
	result := make(map[string]Node, len(b.keys))
	
	if b.node.typ != 'o' || len(b.node.raw) == 0 {
		// 返回空节点
		for _, key := range b.keys {
			result[key] = Node{}
		}
		return result
	}
	
	data := b.node.getWorkingData()
	
	// 一次遍历获取所有需要的键
	pos := b.node.start + 1 // skip '{'
	end := b.node.end
	
	// 创建键查找表
	keySet := make(map[string]bool, len(b.keys))
	for _, key := range b.keys {
		keySet[key] = true
		result[key] = Node{} // 初始化为空节点
	}
	
	foundCount := 0
	for pos < end && foundCount < len(b.keys) {
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
		
		// 找到键的结束位置
		for pos < end && data[pos] != '"' {
			if data[pos] == '\\' {
				pos++
			}
			pos++
		}
		
		if pos >= end {
			break
		}
		
		key := string(data[keyStart:pos])
		pos++ // skip closing quote
		
		// 检查是否是我们需要的键
		if keySet[key] {
			// 跳过冒号
			for pos < end && data[pos] <= ' ' {
				pos++
			}
			if pos < end && data[pos] == ':' {
				pos++
				for pos < end && data[pos] <= ' ' {
					pos++
				}
				
				// 解析值
				valueNode := parseValueAt(data, pos, end)
				if len(b.node.expanded) > 0 {
					valueNode.expanded = b.node.expanded
				}
				result[key] = valueNode
				foundCount++
				pos = valueNode.end
			}
		} else {
			// 跳过不需要的值
			for pos < end && data[pos] != ':' {
				pos++
			}
			if pos < end {
				pos++ // skip ':'
				for pos < end && data[pos] <= ' ' {
					pos++
				}
				pos = skipValueFast(data, pos, end)
			}
		}
		
		// 跳过逗号
		for pos < end && data[pos] <= ' ' {
			pos++
		}
		if pos < end && data[pos] == ',' {
			pos++
		}
	}
	
	return result
}