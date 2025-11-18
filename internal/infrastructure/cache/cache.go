package cache

import (
	"time"

	"github.com/coocood/freecache"
	_ "github.com/coocood/freecache"
)

type Cache struct {
	store *freecache.Cache
}

func NewCache(size int) *Cache {
	return &Cache{store: freecache.NewCache(size * 1024 * 1024)}
}
func (c *Cache) Set(ID string, val []byte, TTL time.Duration) error {
	return c.store.Set([]byte(ID), val, int(TTL.Seconds()))
}
func (c *Cache) Get(ID string) ([]byte, error) {
	return c.store.Get([]byte(ID))
}
func (c *Cache) Del(ID string) bool {
	return c.store.Del([]byte(ID))
}
