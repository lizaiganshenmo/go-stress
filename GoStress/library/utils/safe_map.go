package utils

import "sync"

// 对外暴露的map
type SafeMap struct {
	items map[string]interface{}
	mu    *sync.RWMutex
}

// 新建一个map
func NewSafeMap() *SafeMap {
	return &SafeMap{
		items: map[string]interface{}{},
		mu:    new(sync.RWMutex),
	}
}

// Set 设置key,value
func (m *SafeMap) Set(key string, value interface{}) {
	m.mu.Lock()
	m.items[key] = value
	m.mu.Unlock()
}

// Get 获取key对应的value
func (m *SafeMap) Get(key string) (value interface{}, ok bool) {
	m.mu.RLock()
	value, ok = m.items[key]
	m.mu.RUnlock()
	return value, ok
}

// Count 统计key个数
func (m *SafeMap) Count() int {
	m.mu.RLock()
	count := len(m.items)
	m.mu.RUnlock()
	return count
}

// Keys 所有的key
func (m *SafeMap) Keys() []string {
	m.mu.RLock()
	keys := make([]string, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	m.mu.RUnlock()

	return keys

}

// rlock
func (m *SafeMap) RLock() {
	m.mu.RLock()
}

// runlock
func (m *SafeMap) RUnLock() {
	m.mu.RUnlock()
}

// lock
func (m *SafeMap) Lock() {
	m.mu.Lock()
}

// unlock
func (m *SafeMap) UnLock() {
	m.mu.Unlock()
}
