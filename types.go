package rorm // import "go.szyhf.org/di-rorm"

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

type ValueCacher interface {
	Key() string
	RebuildFunc() bool
}
