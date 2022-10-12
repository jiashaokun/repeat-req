package cache

import (
	"github.com/allegro/bigcache/v3"
	"time"
)

var Cache *bigcache.BigCache

const TimeCache int = 10

const BaseCacheExp = "repeat-req_"

const ListKey = BaseCacheExp + "time:%s"

const TimeFormat = "2006-01-02 15:04"

func InitCache() {
	Cache, _ = bigcache.NewBigCache(bigcache.DefaultConfig(3600 * 12 * time.Minute))
}

func Set(key, val string) {
	Cache.Set(key, []byte(val))
}

func Get(key string) string {
	body, err := Cache.Get(key)
	if err != nil {
		return ""
	}
	return string(body)
}
