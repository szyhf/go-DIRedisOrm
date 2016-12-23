package rorm

import redis "gopkg.in/redis.v5"

type RedisQuerier struct {
	*redis.Client
	// redis.Cmdable
}
