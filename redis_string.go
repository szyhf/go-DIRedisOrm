package rorm

import (
	"github.com/astaxie/beego"
	"gopkg.in/redis.v5"
)

func (r *RedisQuerier) IncrByIfExist(key string, incr int64) (int64, error) {
	beego.Notice("[Redis IncrByIfExist]", key)
	cmds, err := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.IncrBy(key, incr)
		return nil
	})

	if err != nil {
		return 0, err
	}
	if cmds[0].(*redis.BoolCmd).Val() {
		return cmds[1].(*redis.IntCmd).Val(), nil
	} else {
		return 0, ErrorKeyNotExist
	}
}
