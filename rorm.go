package rorm

type rorm struct {
	redisQuerier *RedisQuerier
}

func NewROrmer() ROrmer {
	return new(rorm).Using("default")
}

func (r *rorm) QueryRanking(key string) RankingQuerySeter {
	valueCache, ok := rankingRegistry[key]
	if !ok {
		panic("query ranking '" + key + "' not registried.")
	}
	return &rankingQuerySet{
		querySet: &querySet{
			rorm:       r,
			valueCache: &valueCache,
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
