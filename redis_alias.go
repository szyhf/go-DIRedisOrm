package rorm

import redis "gopkg.in/redis.v5"

var redisRegistry = make(map[string]*redis.Client)

func RegistryRedisClient(alias string, redisClient *redis.Client) {
	if cmd := redisClient.Ping(); cmd.Err() != nil {
		panic("registry redis-client failed: " + cmd.Err().Error())
	}
	redisRegistry[alias] = redisClient
}
