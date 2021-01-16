package cache

import (
	"log"
	"sync"
)

// Cache 快取介面
type Cache interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Del(Key string)
	DelOldest()
	Len() int
}

// DefaultMaxBytes，即預設允許佔用的最大記憶體
// 1 * 2^29
const DefaultMaxBytes = 1 << 29

// safeCache，即平行處理安全快取
type safeCache struct {
	m     sync.RWMutex
	cache Cache

	nhit, nget int
}

func newSafeCache(cache Cache) *safeCache {
	return &safeCache{
		cache: cache,
	}
}

func (sc *safeCache) set(key string, value interface{}) {
	sc.m.Lock()
	defer sc.m.Unlock()
	sc.cache.Set(key, value)
}

func (sc *safeCache) get(key string) interface{} {
	sc.m.RLock()
	defer sc.m.RUnlock()
	sc.nget++
	if sc.cache == nil {
		return nil
	}

	v := sc.cache.Get(key)
	if v != nil {
		log.Println("[TourCache] hit")
		sc.nhit++
	}

	return v
}

func (sc *safeCache) stat() *Stat {
	sc.m.RLock()
	defer sc.m.RUnlock()
	return &Stat{
		NHit: sc.nhit,
		NGet: sc.nget,
	}
}

type Stat struct {
	NHit, NGet int
}
