package rorm

type rorm struct {
	redisQuerier *RedisQuerier
}

func NewROrm() ROrmer {
	return new(rorm).Using("default")
}

func (r *rorm) QueryRanking(key string) RankingQuerySeter {
	return &rankingQuerySet{
		querySet: &querySet{
			rorm: r,
			key:  key,
		},
	}
}

func (r rorm) Using(alias string) ROrmer {
	client, ok := redisRegistry[alias]
	if !ok {
		panic("using reids '" + alias + "' not exist.")
	}
	r.redisQuerier = &RedisQuerier{
		Client: client,
	}
	return &r
}

func (r *rorm) Querier() *RedisQuerier {
	return r.redisQuerier
}
