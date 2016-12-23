package rorm // import "go.szyhf.org/di-rorm"
import redis "gopkg.in/redis.v5"
import "time"

type ROrmer interface {
	QueryRanking(key string) RankingQuerySeter
	Using(alias string) ROrmer
	Querier() *RedisQuerier
}

type QuerySeter interface {
	Key() string
}

type RankingQuerySeter interface {
	SetDefaultCountFunc(defaultCountFunc func() uint) RankingQuerySeter
	SetDefaultIsMembersFunc(defaultIsMembersFunc func() bool) RankingQuerySeter
	SetRebuildFunc(rebuildFunc func() ([]redis.Z, time.Duration)) RankingQuerySeter

	IsMembers(member interface{}) bool
	Count() uint
}
