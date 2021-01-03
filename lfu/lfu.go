package lfu

import (
	"container/heap"

	"github.com/KarasWinds/cache"
)

// lfu 是一個LFU cache，它不是平行處理安全的
type lfu struct {
	// 快取最大的容量(位元組)
	// groupcache 使用的是最大儲存entry個數
	maxBytes int
	// 當一個entry從快取中移除時呼叫該callback函數，預設為nil
	// groupcache 中的key是任意的可比較類型，value是interface{}
	onEvicted func(key string, value interface{})

	// 已使用的字節數，只包括值，key 不算
	usedBytes int

	queue *queue
	cache map[string]*entry
}

// New 創建一個新的 Cache，如果 maxBytes 是 0，表示沒有容量限制
func New(maxBytes int, onEvicted func(key string, value interface{})) cache.Cache {
	q := make(queue, 0, 1024)
	return &lfu{
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
		queue:     &q,
		cache:     make(map[string]*entry),
	}
}

// Set 往 Cache 增加一個元素（如果已經存在，更新值，並增加權重，重新構建堆）
func (l *lfu) Set(key string, value interface{}) {
	if e, ok := l.cache[key]; ok {
		l.usedBytes = l.usedBytes - cache.CalcLen(e.value) + cache.CalcLen(value)
		l.queue.update(e, value, e.weight+1)
		return
	}

	en := &entry{key: key, value: value}
	heap.Push(l.queue, en)
	l.cache[key] = en

	l.usedBytes += en.Len()
	if l.maxBytes > 0 && l.usedBytes > l.maxBytes {
		l.removeElement(heap.Pop(l.queue))
	}
}

// Get 從 cache 中獲取 key 對應的值，nil 表示 key 不存在
func (l *lfu) Get(key string) interface{} {
	if e, ok := l.cache[key]; ok {
		l.queue.update(e, e.value, e.weight+1)
		return e.value
	}

	return nil
}

// Del 從 cache 中刪除 key 對應的元素
func (l *lfu) Del(key string) {
	if e, ok := l.cache[key]; ok {
		heap.Remove(l.queue, e.index)
		l.removeElement(e)
	}
}

// DelOldest 從 cache 中刪除最舊的記錄
func (l *lfu) DelOldest() {
	if l.queue.Len() == 0 {
		return
	}
	l.removeElement(heap.Pop(l.queue))
}

// Len 返回當前 cache 中的記錄數
func (l *lfu) Len() int {
	return l.queue.Len()
}

func (l *lfu) removeElement(x interface{}) {
	if x == nil {
		return
	}

	en := x.(*entry)

	delete(l.cache, en.key)

	l.usedBytes -= en.Len()

	if l.onEvicted != nil {
		l.onEvicted(en.key, en.value)
	}
}
