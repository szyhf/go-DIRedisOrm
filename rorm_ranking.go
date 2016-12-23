package rorm // import "go.szyhf.org/di-rorm"
import (
	"time"

	redis "gopkg.in/redis.v5"
)

var rankingRegistry = make(map[string]rankingCache)

type rankingCache struct {
	key         string
	rebuildFunc func() (members []redis.Z, expire time.Duration)
}

func RegisterRanking(key string, rebuildFunc func() (members []redis.Z, expire time.Duration)) {
	rankingRegistry[key] = rankingCache{
		key:         key,
		rebuildFunc: rebuildFunc,
	}
}

func (r *rankingCache) Key() string {
	return r.key
}

func (r *rankingCache) RebuildFunc() (members []redis.Z, expire time.Duration) {
	if r.rebuildFunc == nil {
		return []redis.Z{}, -1
	}
	return r.rebuildFunc()
}
