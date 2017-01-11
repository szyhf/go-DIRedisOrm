package rorm

import redis "gopkg.in/redis.v5"

func (r *RedisQuerier) ExecPipeline(fun func(pipe *redis.Pipeline) error) ([]redis.Cmder, error) {
	pipe := r.Client.TxPipeline()
	defer pipe.Close()
	if err := fun(pipe); err == nil {
		cmderAry, err := pipe.Exec()
		return cmderAry, err
	} else {
		pipe.Discard()
		return nil, err
	}
}
