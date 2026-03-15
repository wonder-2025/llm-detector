package core

import (
	"regexp"
	"sync"
	"time"
)

// RegexCache 正则表达式缓存
type RegexCache struct {
	cache    map[string]*regexp.Regexp
	mu       sync.RWMutex
	ttl      time.Duration
	lastUsed map[string]time.Time
}

// NewRegexCache 创建新的正则缓存
func NewRegexCache() *RegexCache {
	return &RegexCache{
		cache:    make(map[string]*regexp.Regexp),
		ttl:      30 * time.Minute,
		lastUsed: make(map[string]time.Time),
	}
}

// NewRegexCacheWithTTL 创建带自定义TTL的缓存
func NewRegexCacheWithTTL(ttl time.Duration) *RegexCache {
	return &RegexCache{
		cache:    make(map[string]*regexp.Regexp),
		ttl:      ttl,
		lastUsed: make(map[string]time.Time),
	}
}

// Get 获取编译好的正则表达式
func (c *RegexCache) Get(pattern string) *regexp.Regexp {
	c.mu.RLock()
	if re, exists := c.cache[pattern]; exists {
		c.mu.RUnlock()
		c.mu.Lock()
		c.lastUsed[pattern] = time.Now()
		c.mu.Unlock()
		return re
	}
	c.mu.RUnlock()

	// 编译新正则
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	c.mu.Lock()
	c.cache[pattern] = re
	c.lastUsed[pattern] = time.Now()
	c.mu.Unlock()

	return re
}

// Match 使用缓存的正则进行匹配
func (c *RegexCache) Match(pattern, text string) bool {
	re := c.Get(pattern)
	if re == nil {
		return false
	}
	return re.MatchString(text)
}

// FindStringSubmatch 使用缓存的正则查找子匹配
func (c *RegexCache) FindStringSubmatch(pattern, text string) []string {
	re := c.Get(pattern)
	if re == nil {
		return nil
	}
	return re.FindStringSubmatch(text)
}

// FindAllString 使用缓存的正则查找所有匹配
func (c *RegexCache) FindAllString(pattern, text string, n int) []string {
	re := c.Get(pattern)
	if re == nil {
		return nil
	}
	return re.FindAllString(text, n)
}

// ReplaceAllString 使用缓存的正则替换
func (c *RegexCache) ReplaceAllString(pattern, text, replacement string) string {
	re := c.Get(pattern)
	if re == nil {
		return text
	}
	return re.ReplaceAllString(text, replacement)
}

// Has 检查模式是否已缓存
func (c *RegexCache) Has(pattern string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.cache[pattern]
	return exists
}

// Size 返回缓存大小
func (c *RegexCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Clear 清空缓存
func (c *RegexCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*regexp.Regexp)
	c.lastUsed = make(map[string]time.Time)
}

// Remove 移除特定模式
func (c *RegexCache) Remove(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, pattern)
	delete(c.lastUsed, pattern)
}

// Cleanup 清理过期缓存
func (c *RegexCache) Cleanup() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	for pattern, lastUsed := range c.lastUsed {
		if now.Sub(lastUsed) > c.ttl {
			delete(c.cache, pattern)
			delete(c.lastUsed, pattern)
			removed++
		}
	}

	return removed
}

// GetStats 获取缓存统计
func (c *RegexCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Size:      len(c.cache),
		TTL:       c.ttl,
		LastUsed:  len(c.lastUsed),
	}
}

// CacheStats 缓存统计
type CacheStats struct {
	Size      int
	TTL       time.Duration
	LastUsed  int
}

// GlobalCache 全局正则缓存实例
var GlobalCache = NewRegexCache()

// Match 使用全局缓存进行匹配
func Match(pattern, text string) bool {
	return GlobalCache.Match(pattern, text)
}

// FindStringSubmatch 使用全局缓存查找子匹配
func FindStringSubmatch(pattern, text string) []string {
	return GlobalCache.FindStringSubmatch(pattern, text)
}

// FindAllString 使用全局缓存查找所有匹配
func FindAllString(pattern, text string, n int) []string {
	return GlobalCache.FindAllString(pattern, text, n)
}
