package rorm

import (
	"fmt"

	"time"

	"github.com/astaxie/beego"
	"gopkg.in/redis.v5"
)

// 使用pipline实现的带过期时间的ZAdd
func (r *RedisQuerier) ZAddExpire(key string, members []redis.Z, expire time.Duration) bool {
	beego.Warn("[Redis ZAddExpire]", key, members, expire)
	cmds, err := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.ZAdd(key, members...)
		pipe.Expire(key, expire)
		return nil
	})
	if err != nil {
		for _, cmd := range cmds {
			if cmd.Err() != nil {
				beego.Warn("norm.ZAddExpire(", key, ") failed:", cmd.Err())
			}
		}
		return false
	}
	return true
}

// 使用Pipline实现的优先检查存在性的ZCard
// 如果return -1表示查询失败。
func (r *RedisQuerier) ZCardIfExist(key string) int64 {
	beego.Warn("[Redis ZCardIfExist]", key)
	cmds, err := r.ExecPipeline(func(pipe *redis.Pipeline) error {
		pipe.Exists(key)
		pipe.ZCard(key)
		return nil
	})
	if err != nil {
		for _, cmd := range cmds {
			if cmd.Err() != nil {
				beego.Warn("norm.ZCardIfExist(", key, ") failed:", cmd.Err())
			}
		}
		return -1
	}
	if cmds[0].(*redis.BoolCmd).Val() {
		return cmds[1].(*redis.IntCmd).Val()
	} else {
		return -1
	}

}

func (r *RedisQuerier) ZIsMembers(key string, member string) bool {
	floatCmd := r.ZScore(key, member)
	return floatCmd.Err() == nil
}

// ======= 原生命令 =======

// 将所有指定成员添加到键为key有序集合（sorted set）里面。
// 添加时可以指定多个分数/成员（score/member）对。
// 如果指定添加的成员已经是有序集合里面的成员，则会更新改成员的分数（scrore）并更新到正确的排序位置。
func (r *RedisQuerier) ZAdd(key string, members ...redis.Z) *redis.IntCmd {
	beego.Warn("[Redis ZAdd]", key, members)
	return r.Client.ZAdd(key, members...)
}

// 返回key的有序集元素个数。
func (r *RedisQuerier) ZCard(key string) *redis.IntCmd {
	beego.Warn("[Redis ZCard]", key)
	return r.Client.ZCard(key)
}

// 返回有序集key中，score值在min和max之间(默认包括score值等于min或max)的成员。
func (r *RedisQuerier) ZCount(key string, min int, max int) *redis.IntCmd {
	return r.Client.ZCount(key, fmt.Sprintf("%d", min), fmt.Sprintf("%d", max))
}

// 从ZSet中删除一个或多个成员
func (r *RedisQuerier) ZRem(key string, members ...interface{}) *redis.IntCmd {
	return r.Client.ZRem(key, members...)
}

// 返回有序集key中，指定区间内的成员。其中成员的位置按score值递减(从大到小)来排列。具有相同score值的成员按字典序排列。
func (r *RedisQuerier) ZRange(key string, start int64, stop int64) *redis.StringSliceCmd {
	beego.Warn("[Redis ZRange]", key, start, stop)
	return r.Client.ZRange(key, start, stop)
}

// 返回有序集key中，指定区间内的成员。其中成员的位置按score值递减(从大到小)来排列。具有相同score值的成员按字典序的反序排列。
func (r *RedisQuerier) ZRevRange(key string, start int64, stop int64) *redis.StringSliceCmd {
	beego.Warn("[Redis ZRevRange]", key, start, stop)
	return r.Client.ZRevRange(key, start, stop)
}

// 返回有序集key中，成员member的score值。
// 如果member元素不是有序集key的成员，或key不存在，返回nil。
func (r *RedisQuerier) ZScore(key, member string) *redis.FloatCmd {
	return r.Client.ZScore(key, member)
}
