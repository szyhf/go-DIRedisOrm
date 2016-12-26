package rorm

import redis "gopkg.in/redis.v5"

type RedisQuerier struct {
	Querier
	*redis.Client
	// redis.Cmdable
}
