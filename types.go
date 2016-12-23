package rorm // import "go.szyhf.org/di-rorm"
import redis "gopkg.in/redis.v5"
import "time"

type ROrmer interface {
	QueryRanking(key string) RankingQuerySeter
	Using(alias string) ROrmer
	Querier() *RedisQuerier
}

type QuerySeter interface {
}

type RankingQuerySeter interface {
	Count() uint
}

type KeyCacher interface {
	Key() string
}

type RankingKeyCacher interface {
	KeyCacher
	RebuildFunc() (members []redis.Z, expire time.Duration)
}
