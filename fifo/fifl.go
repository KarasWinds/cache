package fifo

import (
	"container/list"

	"github.com/KarasWinds/cache"
)

// fifo 是一個FIFO cache，不是平行處理安全的
type fifo struct {
	// 快取最大的容量(位元組)
	// groupcache 使用的是最大儲存entry個數
	maxBytes int
	// 當一個entry從快取中移除時呼叫該callback函數，預設nil
	// groupcache 中的key是任意的可比較類型；value是interface{}
	onEvicted func(key string, value interface{})

	// 已使用的位元組數，只包含值，key不算
	usedBytes int

	ll    *list.List
	cache map[string]*list.Element
}

type entry struct {
	key   string
	value interface{}
}

func (e *entry) Len() int {
	return cache.CalcLen(e.value)
}

// New 創建一個新的 Cache，如果 maxBytes 是 0，表示沒有容量限制
func New(maxBytes int, onEvicted func(key string, value interface{})) cache.Cache {
	return &fifo{
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

// Set 往 Cache 尾部增加一個元素（如果已經存在，則放入尾部，並修改值）
func (f *fifo) Set(key string, value interface{}) {
	if e, ok := f.cache[key]; ok {
		f.ll.MoveToBack(e)
		en := e.Value.(*entry)
		f.usedBytes = f.usedBytes - cache.CalcLen(en.value) + cache.CalcLen(value)
		en.value = value
		return
	}

	en := &entry{key, value}
	e := f.ll.PushBack(en)
	f.cache[key] = e

	f.usedBytes += en.Len()
	if f.maxBytes > 0 && f.usedBytes > f.maxBytes {
		f.DelOldest()
	}
}

// Get 從 cache 中獲取 key 對應的值，nil 表示 key 不存在
func (f *fifo) Get(key string) interface{} {
	if e, ok := f.cache[key]; ok {
		return e.Value.(*entry).value
	}

	return nil
}

// Del 從 cache 中刪除 key 對應的記錄
func (f *fifo) Del(key string) {
	if e, ok := f.cache[key]; ok {
		f.removeElement(e)
	}
}

// DelOldest 從 cache 中刪除最舊的記錄
func (f *fifo) DelOldest() {
	f.removeElement(f.ll.Front())
}

// Len 返回當前 cache 中的記錄數
func (f *fifo) Len() int {
	return f.ll.Len()
}

func (f *fifo) removeElement(e *list.Element) {
	if e == nil {
		return
	}

	f.ll.Remove(e)
	en := e.Value.(*entry)
	f.usedBytes -= en.Len()
	delete(f.cache, en.key)

	if f.onEvicted != nil {
		f.onEvicted(en.key, en.value)
	}
}
