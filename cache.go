package cache

// Cache 快取介面
type Cache interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Del(Key string)
	DelOldest()
	Len() int
}
