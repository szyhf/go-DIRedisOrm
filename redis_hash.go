package rorm

import (
	"strings"
	"time"

	"github.com/astaxie/beego"
	redis "gopkg.in/redis.v5"
)

func (r *RedisQuerier) HSetExpire(key string, field string, value interface{}, expire time.Duration) (int64, error) {
	panic("Not imp")
}
func (r *RedisQuerier) HSetExpireIfExsit(key string, field string, value interface{}, expire time.Duration) (int64, error) {
	panic("Not imp")
}
func (r *RedisQuerier) HMSetExpire(key string, kvMap map[string]string, expire time.Duration) (bool, error) {
	panic("Not imp")
}

func (r *RedisQuerier) HGetIfExsit(key string, field string) (string, error) {
	beego.Notice("[Redis HGetIfExsit]", key)
	cmds, _ := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.HGet(key, field)
		return nil
	})
	beego.Warn(cmds[0])
	beego.Warn(cmds[1])
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
