package cache

import (
	"github.com/golang/groupcache"
)

var Cache *SlowDB
var Group *groupcache.Group

const TimeCache int = 10

const BaseCacheExp = "repeat-req_"

const ListKey = BaseCacheExp + "time:%s"

const TimeFormat = "2006-01-02 15:04"

func InitCache() {
	Cache = newSlowDB()
	Group = cacheGet()
}

func cacheGet() *groupcache.Group {
	return groupcache.NewGroup("Repeat-Req-Cache", 64<<20, groupcache.GetterFunc(
		func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
			result := Cache.Get(key)
			dest.SetBytes([]byte(result))
			return nil
		}))
}

type SlowDB struct {
	data map[string]string
}

func (db *SlowDB) Get(key string) string {
	return db.data[key]
}

func (db *SlowDB) Set(key string, value string) {
	db.data[key] = value
}

func newSlowDB() *SlowDB {
	ndb := new(SlowDB)
	ndb.data = make(map[string]string)
	return ndb
}
