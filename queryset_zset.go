package rorm

import (
	"time"

	"github.com/astaxie/beego"
	redis "gopkg.in/redis.v5"
)

type zsetQuerySet struct {
	*querySet
	rebuildFunc func() ([]redis.Z, time.Duration)
}

// ========= 查询接口 =========

func (q *zsetQuerySet) Count() (int64, error) {
	// 尝试直接从缓存拿
	count, err := q.Querier().ZCardIfExist(q.Key())
	if err == nil {
		return count, nil
	}

	// 重建缓存
	if q.rebuildingProcess(q) {
		// 重建成功则重新获取
		return q.Count()
	}

	// 从用户提供的默认方法获取
	return 0, ErrorCanNotRebuild
}

func (q *zsetQuerySet) IsMember(member string) (bool, error) {
	// 尝试直接从缓存拿
	exist, err := q.Querier().ZIsMember(q.Key(), member)
	if err == nil {
		return exist, nil
	}

	// 重建缓存
	if q.rebuildingProcess(q) {
		return q.IsMember(member)
	}

	// 从用户提供的默认方法获取
	return false, ErrorCanNotRebuild
}

func (q *zsetQuerySet) RangeASC(start, stop int64) ([]string, error) {
	// 尝试直接从缓存拿
	members, err := q.Querier().ZRangeIfExist(q.Key(), start, stop)
	if err == nil {
		return members, nil
	}

	// 缓存获取失败尝试重构缓存
	if q.rebuildingProcess(q) {
		return q.RangeASC(start, stop)
	}

	// 使用用户的默认设置
	return nil, ErrorCanNotRebuild
}

func (q *zsetQuerySet) RangeDESC(start, stop int64) ([]string, error) {
	// 尝试直接从缓存拿
	members, err := q.Querier().ZRevRangeIfExist(q.Key(), start, stop)
	if err == nil {
		return members, nil
	}

	// 缓存获取失败尝试重构缓存
	if q.rebuildingProcess(q) {
		return q.RangeDESC(start, stop)
	}

	// 使用用户的默认设置
	return nil, ErrorCanNotRebuild
}

func (q *zsetQuerySet) Members() ([]string, error) {
	// 利用Range的负数参数指向倒数的元素的特性实现
	return q.RangeASC(0, -1)
}

// ========= 写入接口 =========

func (q *zsetQuerySet) AddExpire(member interface{}, score float64, expire time.Duration) error {
	ro := q.rorm
	qr := ro.Querier()
	// 如果不增加过期方法，可能会创建一个不会过期的集合
	qr.ZAddExpire(q.Key(), []redis.Z{redis.Z{
		Member: member,
		Score:  score,
	}},
		expire)
	return nil
}

func (q *zsetQuerySet) Rem(member ...interface{}) error {
	cmd := q.Querier().ZRem(q.Key(), member...)
	return cmd.Err()
}

// ============= 连贯操作 =============

// 防止频繁重建
// expire 保护有效时间
func (q zsetQuerySet) Protect(expire time.Duration) ZSetQuerySeter {
	q.isProtectDB = true
	q.protectExpire = expire
	return &q
}

func (q zsetQuerySet) SetRebuildFunc(rebuildFunc func() ([]redis.Z, time.Duration)) ZSetQuerySeter {
	q.rebuildFunc = rebuildFunc
	return &q
}

// ============= 辅助方法 =============

func (q *zsetQuerySet) Rebuilding() error {
	// 重建缓存
	beego.Notice("zsetQuerySet.rebuild(", q.Key(), ")")
	if members, expire := q.callRebuildFunc(); len(members) > 0 {
		return q.rorm.Querier().ZAddExpire(q.Key(), members, expire)
	}
	return ErrorCanNotRebuild
}

func (q *zsetQuerySet) callRebuildFunc() ([]redis.Z, time.Duration) {
	if q.rebuildFunc == nil {
		return []redis.Z{}, -1
	}
	return q.rebuildFunc()
}
