package fxjson

import (
	"fmt"
	"hash/crc32"
	"sync"
	"time"
)

// Cache 缓存接口
type Cache interface {
	Get(key string) (Node, bool)
	Set(key string, node Node, ttl time.Duration)
	Delete(key string)
	Clear()
	Stats() CacheStats
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Sets      int64   `json:"sets"`
	Deletes   int64   `json:"deletes"`
	Evictions int64   `json:"evictions"`
	Size      int     `json:"size"`
	MaxSize   int     `json:"max_size"`
	HitRate   float64 `json:"hit_rate"`
}

// CacheItem 缓存项
type CacheItem struct {
	Value     Node      `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	AccessAt  time.Time `json:"access_at"`
	HitCount  int64     `json:"hit_count"`
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	items   map[string]*CacheItem
	mutex   sync.RWMutex
	maxSize int
	stats   CacheStats
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(maxSize int) *MemoryCache {
	cache := &MemoryCache{
		items:   make(map[string]*CacheItem),
		maxSize: maxSize,
		stats:   CacheStats{MaxSize: maxSize},
	}

	// 启动清理goroutine
	go cache.cleanupExpired()

	return cache
}

// Get 获取缓存值
func (mc *MemoryCache) Get(key string) (Node, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	item, exists := mc.items[key]
	if !exists {
		mc.stats.Misses++
		return Node{}, false
	}

	// 检查是否过期
	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		mc.mutex.RUnlock()
		mc.mutex.Lock()
		delete(mc.items, key)
		mc.stats.Evictions++
		mc.stats.Size--
		mc.mutex.Unlock()
		mc.mutex.RLock()

		mc.stats.Misses++
		return Node{}, false
	}

	// 更新访问信息
	item.AccessAt = time.Now()
	item.HitCount++
	mc.stats.Hits++

	// 计算命中率
	total := mc.stats.Hits + mc.stats.Misses
	if total > 0 {
		mc.stats.HitRate = float64(mc.stats.Hits) / float64(total)
	}

	return item.Value, true
}

// Set 设置缓存值
func (mc *MemoryCache) Set(key string, node Node, ttl time.Duration) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// 检查是否需要清理空间
	if len(mc.items) >= mc.maxSize {
		mc.evictLRU()
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	mc.items[key] = &CacheItem{
		Value:     node,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		AccessAt:  time.Now(),
		HitCount:  0,
	}

	mc.stats.Sets++
	mc.stats.Size = len(mc.items)
}

// Delete 删除缓存项
func (mc *MemoryCache) Delete(key string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if _, exists := mc.items[key]; exists {
		delete(mc.items, key)
		mc.stats.Deletes++
		mc.stats.Size--
	}
}

// Clear 清空所有缓存
func (mc *MemoryCache) Clear() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.items = make(map[string]*CacheItem)
	mc.stats.Size = 0
}

// Stats 获取缓存统计
func (mc *MemoryCache) Stats() CacheStats {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	stats := mc.stats
	stats.Size = len(mc.items)
	return stats
}

// evictLRU 使用LRU策略清理缓存
func (mc *MemoryCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range mc.items {
		if oldestKey == "" || item.AccessAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.AccessAt
		}
	}

	if oldestKey != "" {
		delete(mc.items, oldestKey)
		mc.stats.Evictions++
	}
}

// cleanupExpired 清理过期项
func (mc *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mc.mutex.Lock()
		now := time.Now()

		for key, item := range mc.items {
			if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
				delete(mc.items, key)
				mc.stats.Evictions++
			}
		}

		mc.stats.Size = len(mc.items)
		mc.mutex.Unlock()
	}
}

// 全局缓存实例
var globalCache Cache = NewMemoryCache(1000)

// EnableCaching 启用全局缓存
func EnableCaching(cache Cache) {
	globalCache = cache
}

// DisableCaching 禁用全局缓存
func DisableCaching() {
	globalCache = nil
}

// generateCacheKey 生成缓存键
// 使用CRC32代替MD5以提高性能，因为我们只需要快速哈希而不需要加密强度
func generateCacheKey(data []byte) string {
	hash := crc32.ChecksumIEEE(data)
	return fmt.Sprintf("fxjson:%08x", hash)
}

// FromBytesWithCache 带缓存的JSON解析
func FromBytesWithCache(b []byte, ttl time.Duration) Node {
	if globalCache == nil {
		return FromBytes(b)
	}

	key := generateCacheKey(b)

	// 尝试从缓存获取
	if cached, exists := globalCache.Get(key); exists {
		return cached
	}

	// 解析并缓存
	node := FromBytes(b)
	if node.Exists() {
		globalCache.Set(key, node, ttl)
	}

	return node
}

// 性能监控
type PerformanceMonitor struct {
	parseCount     int64
	totalParseTime time.Duration
	mutex          sync.RWMutex
}

var perfMonitor = &PerformanceMonitor{}

// GetPerformanceStats 获取性能统计
func GetPerformanceStats() map[string]interface{} {
	perfMonitor.mutex.RLock()
	defer perfMonitor.mutex.RUnlock()

	var avgParseTime time.Duration
	if perfMonitor.parseCount > 0 {
		avgParseTime = perfMonitor.totalParseTime / time.Duration(perfMonitor.parseCount)
	}

	return map[string]interface{}{
		"parse_count":      perfMonitor.parseCount,
		"total_parse_time": perfMonitor.totalParseTime.String(),
		"avg_parse_time":   avgParseTime.String(),
		"cache_stats":      globalCache.Stats(),
	}
}

// recordParseTime 记录解析时间
func recordParseTime(duration time.Duration) {
	perfMonitor.mutex.Lock()
	defer perfMonitor.mutex.Unlock()

	perfMonitor.parseCount++
	perfMonitor.totalParseTime += duration
}

// FromBytesWithMetrics 带性能监控的JSON解析
func FromBytesWithMetrics(b []byte) Node {
	start := time.Now()
	defer func() {
		recordParseTime(time.Since(start))
	}()

	return FromBytes(b)
}

// BatchProcessor 批处理器
type BatchProcessor struct {
	batchSize int
	processor func([]Node) error
	buffer    []Node
	mutex     sync.Mutex
}

// NewBatchProcessor 创建批处理器
func NewBatchProcessor(batchSize int, processor func([]Node) error) *BatchProcessor {
	return &BatchProcessor{
		batchSize: batchSize,
		processor: processor,
		buffer:    make([]Node, 0, batchSize),
	}
}

// Add 添加项到批处理器
func (bp *BatchProcessor) Add(node Node) error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	bp.buffer = append(bp.buffer, node)

	if len(bp.buffer) >= bp.batchSize {
		return bp.flush()
	}

	return nil
}

// Flush 手动刷新批处理器
func (bp *BatchProcessor) Flush() error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	return bp.flush()
}

// flush 内部刷新方法
func (bp *BatchProcessor) flush() error {
	if len(bp.buffer) == 0 {
		return nil
	}

	err := bp.processor(bp.buffer)
	bp.buffer = bp.buffer[:0] // 清空buffer但保持容量

	return err
}

// LazyLoader 延迟加载器
type LazyLoader struct {
	loadFunc func() (Node, error)
	cache    *Node
	loaded   bool
	mutex    sync.RWMutex
}

// NewLazyLoader 创建延迟加载器
func NewLazyLoader(loadFunc func() (Node, error)) *LazyLoader {
	return &LazyLoader{
		loadFunc: loadFunc,
	}
}

// Get 获取值（延迟加载）
func (ll *LazyLoader) Get() (Node, error) {
	ll.mutex.RLock()
	if ll.loaded {
		result := *ll.cache
		ll.mutex.RUnlock()
		return result, nil
	}
	ll.mutex.RUnlock()

	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	// 双重检查
	if ll.loaded {
		return *ll.cache, nil
	}

	node, err := ll.loadFunc()
	if err != nil {
		return Node{}, err
	}

	ll.cache = &node
	ll.loaded = true

	return node, nil
}

// Reset 重置延迟加载器
func (ll *LazyLoader) Reset() {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	ll.loaded = false
	ll.cache = nil
}
