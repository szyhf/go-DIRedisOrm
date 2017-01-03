package rorm

type rorm struct {
	redisQuerier *RedisQuerier
}

func NewROrm() ROrmer {
	return new(rorm).Using("default")
}

func (r *rorm) QueryHash(key string) HashQuerySeter {
	return &hashQuerySet{
		querySet: &querySet{
			rorm: r,
			key:  key,
		},
	}
}

func (r *rorm) QueryKeys(key string) KeysQuerySeter {
	return &keysQuerySet{
		querySet: &querySet{
			rorm: r,
			key:  key,
		},
	}
}

func (r *rorm) QueryString(key string) StringQuerySeter {
	return &stringQuerySet{
		querySet: &querySet{
			rorm: r,
			key:  key,
		},
	}
}

func (r *rorm) QueryZSet(key string) ZSetQuerySeter {
	return &zsetQuerySet{
		querySet: &querySet{
			rorm: r,
			key:  key,
		},
	}
}

func (r *rorm) QuerySet(key string) SetQuerySeter {
	return &setQuerySet{
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

func (r *rorm) Querier() Querier {
	return r.redisQuerier
}
