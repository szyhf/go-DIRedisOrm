package rorm

import (
	"strings"
	"time"

	"github.com/astaxie/beego"
	redis "gopkg.in/redis.v5"
)

func (r *RedisQuerier) HSetExpire(key string, field string, value interface{}, expire time.Duration) (bool, error) {
	panic("Not imp")
}
func (r *RedisQuerier) HSetExpireIfExist(key string, field string, value interface{}, expire time.Duration) (bool, error) {
	beego.Notice("[Redis HSetExpireIfExist]")
	cmds, _ := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.HSet(key, field, value)
		pipe.Expire(key, expire)
		return nil
	})
	// Pipeline默认返回的是最后一个err，所以这里的判定方式要做调整
	if cmds[0].Err() != nil {
		return false, cmds[0].Err()
	}

	if cmds[0].(*redis.BoolCmd).Val() {
		if cmds[1].Err() == nil {
			return cmds[1].(*redis.BoolCmd).Val(), nil
		} else if strings.HasPrefix(cmds[1].Err().Error(), "WRONGTYPE") {
			// 数据库保护产生的空键
			return false, nil
		} else {
			return false, cmds[1].Err()
		}
	} else {
		return false, ErrorKeyNotExist
	}
}
func (r *RedisQuerier) HMSetExpire(key string, kvMap map[string]string, expire time.Duration) (bool, error) {
	beego.Notice("[Redis HMSetExpire]", key, kvMap, expire)
	cmds, err := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.HMSet(key, kvMap)
		pipe.Expire(key, expire)
		return nil
	})
	if err == nil {
		return cmds[0].(*redis.StatusCmd).Val() == "OK", nil
	}
	return false, err
}

func (r *RedisQuerier) HGetIfExist(key string, field string) (string, error) {
	beego.Notice("[Redis HGetIfExist]", key)
	cmds, _ := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.HGet(key, field)
		return nil
	})
	// Pipeline默认返回的是最后一个err，所以这里的判定方式要做调整
	if cmds[0].Err() != nil {
		return "", cmds[0].Err()
	}
	if cmds[0].(*redis.BoolCmd).Val() {
		if cmds[1].Err() == nil {
			return cmds[1].(*redis.StringCmd).Val(), nil
		} else if strings.HasPrefix(cmds[1].Err().Error(), "WRONGTYPE") {
			// 数据库保护产生的空键
			return "", nil
		} else {
			return "", cmds[1].Err()
		}
	} else {
		return "", ErrorKeyNotExist
	}
}
