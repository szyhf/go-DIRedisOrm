package rorm

import (
	"time"

	"github.com/astaxie/beego"
	redis "gopkg.in/redis.v5"
)

type zsetQuerySet struct {
	*querySet
	rebuildFunc          func() ([]redis.Z, time.Duration)
	defaultCountFunc     func() int64
	defaultIsMemberFunc  func(string) bool
	defaultRangeASCFunc  func(start, stop int64) []string
	defaultRangeDESCFunc func(start, stop int64) []string
}

// ========= 查询接口 =========

func (q *zsetQuerySet) Count() int64 {
	// 尝试直接从缓存拿
	count, err := q.Querier().ZCardIfExist(q.Key())
	if err == nil {
		return count
	}

	// 重建缓存
	if q.rebuildingProcess(q) {
		// 重建成功则重新获取
		return q.Count()
	}

	// 从用户提供的默认方法获取
	return q.callDefaultCountFunc()
}

func (q *zsetQuerySet) IsMember(member string) bool {
	// 尝试直接从缓存拿
	exist, err := q.Querier().ZIsMember(q.Key(), member)
	if err == nil {
		return exist
	}

	// 重建缓存
	if q.rebuildingProcess(q) {
		return q.IsMember(member)
	}

	// 从用户提供的默认方法获取
	return q.callDefaultIsMemberFunc(member)
}

func (q *zsetQuerySet) RangeASC(start, stop int64) []string {
	// 尝试直接从缓存拿
	members, err := q.Querier().ZRangeIfExist(q.Key(), start, stop)
	if err == nil {
		return members
	}

	// 缓存获取失败尝试重构缓存
	if q.rebuildingProcess(q) {
		return q.RangeASC(start, stop)
	}

	// 使用用户的默认设置
	return q.callDefaultRangeASCFunc(start, stop)
}

func (q *zsetQuerySet) RangeDESC(start, stop int64) []string {
	// 尝试直接从缓存拿
	members, err := q.Querier().ZRevRangeIfExist(q.Key(), start, stop)
	if err == nil {
		return members
	}

	// 缓存获取失败尝试重构缓存
	if q.rebuildingProcess(q) {
		return q.RangeDESC(start, stop)
	}

	// 使用用户的默认设置
	return q.callDefaultRangeDESCFunc(start, stop)
}

func (q *zsetQuerySet) Members() []string {
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

func (q zsetQuerySet) SetDefaultCountFunc(defaultCountFunc func() int64) ZSetQuerySeter {
	q.defaultCountFunc = defaultCountFunc
	return &q
}

func (q zsetQuerySet) SetDefaultIsMemberFunc(defaultIsMemberFunc func(member string) bool) ZSetQuerySeter {
	q.defaultIsMemberFunc = defaultIsMemberFunc
	return &q
}

// 默认获取ZSet成员的方法
func (q zsetQuerySet) SetDefaultRangeASCFunc(defaultRangeASCFunc func(start, stop int64) []string) ZSetQuerySeter {
	q.defaultRangeASCFunc = defaultRangeASCFunc
	return &q
}

// 默认获取ZSet成员的方法
func (q zsetQuerySet) SetDefaultRangeDESCFunc(defaultRangeDESCFunc func(start, stop int64) []string) ZSetQuerySeter {
	q.defaultRangeDESCFunc = defaultRangeDESCFunc
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

func (q *zsetQuerySet) callDefaultCountFunc() int64 {
	if q.defaultCountFunc == nil {
		return 0
	}
	return q.defaultCountFunc()
}

func (q *zsetQuerySet) callDefaultIsMemberFunc(member string) bool {
	if q.defaultIsMemberFunc == nil {
		return false
	}
	return q.defaultIsMemberFunc(member)
}

func (q *zsetQuerySet) callDefaultRangeASCFunc(start, stop int64) []string {
	if q.defaultRangeASCFunc == nil {
		return []string{}
	}
	return q.defaultRangeASCFunc(start, stop)
}

func (q *zsetQuerySet) callDefaultRangeDESCFunc(start, stop int64) []string {
	if q.defaultRangeDESCFunc == nil {
		return []string{}
	}
	return q.defaultRangeDESCFunc(start, stop)
}

func (q *zsetQuerySet) callRebuildFunc() ([]redis.Z, time.Duration) {
	if q.rebuildFunc == nil {
		return []redis.Z{}, -1
	}
	return q.rebuildFunc()
}
