package rorm

import (
	"time"

	"strings"

	"github.com/astaxie/beego"
	"gopkg.in/redis.v5"
)

// 使用pipline实现的带过期时间的ZAdd
func (r *RedisQuerier) ZAddExpire(key string, members []redis.Z, expire time.Duration) error {
	beego.Notice("[Redis ZAddExpire]", key, members, expire)
	_, err := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.ZAdd(key, members...)
		pipe.Expire(key, expire)
		return nil
	})

	return err
}

// 使用Pipline实现的优先检查存在性的ZCard
func (r *RedisQuerier) ZCardIfExist(key string) (int64, error) {
	beego.Notice("[Redis ZCardIfExist]", key)
	cmds, err := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.ZCard(key)
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

// 判定Key是否存在，如果存在则返回指定排序区间的成员（正序）
func (r *RedisQuerier) ZRangeIfExist(key string, start, stop int64) ([]string, error) {
	beego.Notice("[Redis ZRangeIfExist]", key)
	cmds, err := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.ZRange(key, start, stop)
		return nil
	})
	if err != nil {
		return nil, err
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

// 判定Key是否存在，如果存在则返回指定排序区间的成员（逆序）
func (r *RedisQuerier) ZRevRangeIfExist(key string, start, stop int64) ([]string, error) {
	beego.Notice("[Redis ZRevRangeIfExist]", key)
	cmds, err := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.ZRevRange(key, start, stop)
		return nil
	})
	if err != nil {
		return nil, err
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

// 判定Key是否存在，如果存在则检查member是否在集合中
func (r *RedisQuerier) ZIsMember(key string, member string) (bool, error) {
	beego.Notice("[Redis ZIsMember]", key)
	// 通过ZRank间接实现存在性判断
	// ZScore返回member在ZSet中的Index
	cmds, _ := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.ZScore(key, member)
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

// ======= 原生命令 =======

// 将所有指定成员添加到键为key有序集合（sorted set）里面。
// 添加时可以指定多个分数/成员（score/member）对。
// 如果指定添加的成员已经是有序集合里面的成员，则会更新改成员的分数（scrore）并更新到正确的排序位置。
func (r *RedisQuerier) ZAdd(key string, members ...redis.Z) *redis.IntCmd {
	beego.Notice("[Redis ZAdd]", key, members)
	return r.Client.ZAdd(key, members...)
}

// 返回key的有序集元素个数。
func (r *RedisQuerier) ZCard(key string) *redis.IntCmd {
	beego.Notice("[Redis ZCard]", key)
	return r.Client.ZCard(key)
}

// 返回有序集key中，score值在min和max之间(默认包括score值等于min或max)的成员。
// func (r *RedisQuerier) ZCount(key string, min int, max int) *redis.IntCmd {
// 	beego.Notice("[Redis Count]", key, min, max)
// 	return r.Client.ZCount(key, fmt.Sprintf("%d", min), fmt.Sprintf("%d", max))
// }

// 从ZSet中删除一个或多个成员
func (r *RedisQuerier) ZRem(key string, members ...interface{}) *redis.IntCmd {
	beego.Notice("[Redis ZRem]", key, members)
	return r.Client.ZRem(key, members...)
}

// 返回有序集key中，指定区间内的成员。其中成员的位置按score值递减(从大到小)来排列。具有相同score值的成员按字典序排列。
func (r *RedisQuerier) ZRange(key string, start int64, stop int64) *redis.StringSliceCmd {
	beego.Notice("[Redis ZRange]", key, start, stop)
	return r.Client.ZRange(key, start, stop)
}

// 返回有序集key中，指定区间内的成员。其中成员的位置按score值递减(从大到小)来排列。具有相同score值的成员按字典序的反序排列。
func (r *RedisQuerier) ZRevRange(key string, start int64, stop int64) *redis.StringSliceCmd {
	beego.Notice("[Redis ZRevRange]", key, start, stop)
	return r.Client.ZRevRange(key, start, stop)
}

// 返回有序集key中，成员member的score值。
// 如果member元素不是有序集key的成员，或key不存在，返回nil。
func (r *RedisQuerier) ZScore(key, member string) *redis.FloatCmd {
	beego.Notice("[Redis ZScore]", key, member)
	return r.Client.ZScore(key, member)
}
