package cache

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var Cache *cache.Cache

const TimeCache int = 10

const BaseCacheExp = "repeat-req_"

const ListKey = BaseCacheExp + "time:%s"

const TimeFormat = "2006-01-02 15:04"

func InitCache() {
	Cache = cache.New(0*time.Minute, 0*time.Minute)
}
