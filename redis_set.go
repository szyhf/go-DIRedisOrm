package rorm

import (
	"strings"
	"time"

	"github.com/astaxie/beego"
	redis "gopkg.in/redis.v5"
)

// ========== 扩展方法 ==========

// 添加一个或多个member到集合中，并设置集合的过期时间
func (r *RedisQuerier) SAddExpire(key string, members []interface{}, expire time.Duration) (int64, error) {
	beego.Notice("[Redis SAddExpire]", key, members, expire)
	cmds, err := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.SAdd(key, members...)
		pipe.Expire(key, expire)
		return nil
	})
	if err == nil {
		return cmds[0].(*redis.IntCmd).Val(), nil
	}
	return 0, err
}

func (r *RedisQuerier) SAddExpireIfExist(key string, members []interface{}, expire time.Duration) (int64, error) {
	beego.Notice("[Redis SAddExpireIfExist]", key, members, expire)
	cmds, _ := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.SAdd(key, members...)
		pipe.Expire(key, expire)
		return nil
	})

	// Pipeline默认返回的是最后一个err，所以这里的判定方式要做调整
	if cmds[0].Err() != nil {
		return 0, cmds[0].Err()
	}

	if cmds[0].(*redis.BoolCmd).Val() {
		if cmds[1].Err() == nil {
			return cmds[1].(*redis.IntCmd).Val(), nil
			// } else if strings.HasPrefix(cmds[1].Err().Error(), "WRONGTYPE") {
			// 	// 数据库保护产生的空键
			// 	return 0, nil
		} else {
			return 0, cmds[1].Err()
		}
	} else {
		return 0, ErrorKeyNotExist
	}
}

// 统计当前集合中有多少个元素
func (r *RedisQuerier) SCardIfExist(key string) (int64, error) {
	beego.Notice("[Redis SCardIfExist]", key)
	cmds, _ := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.SCard(key)
		return nil
	})
	// Pipeline默认返回的是最后一个err，所以这里的判定方式要做调整
	if cmds[0].Err() != nil {
		return 0, cmds[0].Err()
	}
	if cmds[0].(*redis.BoolCmd).Val() {
		return cmds[1].(*redis.IntCmd).Val(), nil
	} else {
		return 0, ErrorKeyNotExist
	}
}

// 获取集合中的所有成员
func (r *RedisQuerier) SMembersIfExist(key string) ([]string, error) {
	beego.Notice("[Redis SMembersIfExist]", key)
	cmds, _ := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.SMembers(key)
		return nil
	})
	// Pipeline默认返回的是最后一个err，所以这里的判定方式要做调整
	if cmds[0].Err() != nil {
		return nil, cmds[0].Err()
	}
	if cmds[0].(*redis.BoolCmd).Val() {
		if cmds[1].Err() == nil {
			return cmds[1].(*redis.StringSliceCmd).Val(), nil
		} else if strings.HasPrefix(cmds[1].Err().Error(), "WRONGTYPE") {
			// 数据库保护产生的空键
			return nil, nil
		} else {
			return nil, cmds[1].Err()
		}
	} else {
		return nil, ErrorKeyNotExist
	}
}

// 确认集合中是否有成员
func (r *RedisQuerier) SIsMemberIfExist(key string, member interface{}) (bool, error) {
	beego.Notice("[Redis SIsMemberIfExist]", key, member)
	cmds, _ := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.SIsMember(key, member)
		return nil
	})

	// Pipeline默认返回的是最后一个err，所以这里的判定方式要做调整
	if cmds[0].Err() != nil {
		return false, cmds[0].Err()
	}
	if cmds[0].(*redis.BoolCmd).Val() {
		// 如果member不存在，则会返回error=redis.Nil
		if cmds[1].Err() == nil {
			// member存在
			return true, nil
		} else if cmds[1].Err() == redis.Nil {
			// member不存在，虽然有err但属于正常情况
			return false, nil
		} else if strings.HasPrefix(cmds[1].Err().Error(), "WRONGTYPE") {
			// 数据库保护产生的空键
			return false, nil
		} else {
			// err!=redis.Nil，说明是其他异常，要返回异常
			return false, cmds[1].Err()
		}
	} else {
		return false, ErrorKeyNotExist
	}
}
