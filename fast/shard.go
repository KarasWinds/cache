package fast

import (
	"container/list"
	"sync"
)

type cacheShard struct {
	locker sync.RWMutex

	// 最大儲存 entry 個數
	maxEntries int
	// 當一個 entry 從快取中移除調用該 callback 函數，默認為 nil
	// groupcache 中的 key 是任意的可比較類型；value 是 interface{}
	onEvicted func(key string, value interface{})

	ll    *list.List
	cache map[string]*list.Element
}

type entry struct {
	key   string
	value interface{}
}

// new 建立一個新的 cacheShard，如果 maxByte 是0，表示沒有容量限制
func newCacheShard(maxEntries int, onEvicted func(key string, value interface{})) *cacheShard {
	return &cacheShard{
		maxEntries: maxEntries,
		onEvicted:  onEvicted,
		ll:         list.New(),
		cache:      make(map[string]*list.Element),
	}
}

// set 往 Cache 尾部增加一個元素(如果已經存在，則放入尾部，並更新值)
func (c *cacheShard) set(key string, value interface{}) {
	c.locker.Lock()
	defer c.locker.Unlock()

	if e, ok := c.cache[key]; ok {
		c.ll.MoveToBack(e)
		en := e.Value.(*entry)
		en.value = value
		return
	}

	en := &entry{key, value}
	e := c.ll.PushBack(en)
	c.cache[key] = e

	if c.maxEntries > 0 && c.ll.Len() > c.maxEntries {
		c.removeElement(c.ll.Front())
	}
}

// get 從 cache 中獲取 key 對應的值，nil表示 key 不存在
func (c *cacheShard) get(key string) interface{} {
	c.locker.Lock()
	defer c.locker.Unlock()

	if e, ok := c.cache[key]; ok {
		c.ll.MoveToBack(e)
		return e.Value.(*entry).value
	}

	return nil
}

// del 從 cache 中刪除 key 對應的 element
func (c *cacheShard) del(key string) {
	c.locker.Lock()
	defer c.locker.Unlock()
	if e, ok := c.cache[key]; ok {
		c.removeElement(e)
	}
}

// delOldest 從 cache 中刪除最舊的 element
func (c *cacheShard) delOldest() {
	c.locker.Lock()
	defer c.locker.Unlock()

	c.removeElement(c.ll.Front())

}

// len 返回當前 cache 中的 entry 數量
func (c *cacheShard) len() int {
	c.locker.Lock()
	defer c.locker.Unlock()

	return c.ll.Len()
}

func (c *cacheShard) removeElement(e *list.Element) {
	if e != nil {
		c.ll.Remove(e)
		en := e.Value.(*entry)
		delete(c.cache, en.key)

		if c.onEvicted != nil {
			c.onEvicted(en.key, en.value)
		}
	}

	return
}
