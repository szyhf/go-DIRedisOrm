package rorm

import (
	"time"

	"github.com/astaxie/beego"
	redis "gopkg.in/redis.v5"
)

type rankingQuerySet struct {
	*querySet
	rebuildFunc          func() ([]redis.Z, time.Duration)
	defaultCountFunc     func() uint
	defaultIsMembersFunc func(string) bool
	defaultRangeASCFunc  func(start, stop int64) []string
	defaultRangeDESCFunc func(start, stop int64) []string

	// 状态标识，防止重构缓存失败后陷入死循环
	isRebuilding bool
}

// ========= 查询接口 =========

func (r *rankingQuerySet) Count() uint {
	// 尝试直接从缓存拿
	ro := r.rorm
	qr := ro.Querier()
	count, err := qr.ZCardIfExist(r.Key())
	if err == nil {
		return uint(count)
	}

	// 重建缓存
	if r.rebuild() {
		// 重建成功则重新获取
		r.Count()
	}

	// 从用户提供的默认方法获取
	return r.callDefaultCountFunc()
}

func (r *rankingQuerySet) IsMembers(member string) bool {
	// 尝试直接从缓存拿
	ro := r.rorm
	qr := ro.Querier()
	exist, err := qr.ZIsMembers(r.Key(), member)
	if err == nil {
		return exist
	}

	// 重建缓存
	if r.rebuild() {
		return r.IsMembers(member)
	}

	// 从用户提供的默认方法获取
	return r.callDefaultIsMembersFunc(member)
}

func (r *rankingQuerySet) RangeASC(start, stop int64) []string {
	// 尝试直接从缓存拿
	ro := r.rorm
	qr := ro.Querier()
	members, err := qr.ZRangeIfExist(r.Key(), start, stop)
	if err == nil {
		return members
	}

	// 缓存获取失败尝试重构缓存
	if r.rebuild() {
		return r.RangeASC(start, stop)
	}

	// 使用用户的默认设置
	return r.defaultRangeASCFunc(start, stop)
}

func (r *rankingQuerySet) RangeDESC(start, stop int64) []string {
	// 尝试直接从缓存拿
	ro := r.rorm
	qr := ro.Querier()
	members, err := qr.ZRevRangeIfExist(r.Key(), start, stop)
	if err == nil {
		return members
	}

	// 缓存获取失败尝试重构缓存
	if r.rebuild() {
		return r.RangeDESC(start, stop)
	}

	// 使用用户的默认设置
	return r.defaultRangeDESCFunc(start, stop)
}

// ========= 写入接口 =========

func (r *rankingQuerySet) AddExpire(member interface{}, score float64, expire time.Duration) error {
	ro := r.rorm
	qr := ro.Querier()
	// 如果不增加过期方法，可能会创建一个不会过期的集合
	qr.ZAddExpire(r.Key(), []redis.Z{redis.Z{
		Member: member,
		Score:  score,
	}},
		expire)
	return nil
}

func (r *rankingQuerySet) Rem(member ...interface{}) error {
	ro := r.rorm
	qr := ro.Querier()
	qr.ZRem(r.Key(), member...)
	return nil
}

// ============= 连贯操作 =============

// 防止频繁重建
// expire 保护有效时间
func (r rankingQuerySet) Protect(expire time.Duration) RankingQuerySeter {
	r.isProtectDB = true
	r.protectExpire = expire
	return &r
}

func (r rankingQuerySet) SetRebuildFunc(rebuildFunc func() ([]redis.Z, time.Duration)) RankingQuerySeter {
	r.rebuildFunc = rebuildFunc
	return &r
}

func (r rankingQuerySet) SetDefaultCountFunc(defaultCountFunc func() uint) RankingQuerySeter {
	r.defaultCountFunc = defaultCountFunc
	return &r
}

func (r rankingQuerySet) SetDefaultIsMembersFunc(defaultIsMembersFunc func(member string) bool) RankingQuerySeter {
	r.defaultIsMembersFunc = defaultIsMembersFunc
	return &r
}

// 默认获取ZSet成员的方法
func (r rankingQuerySet) SetDefaultRangeASCFunc(defaultRangeASCFunc func(start, stop int64) []string) RankingQuerySeter {
	r.defaultRangeASCFunc = defaultRangeASCFunc
	return &r
}

// 默认获取ZSet成员的方法
func (r rankingQuerySet) SetDefaultRangeDESCFunc(defaultRangeDESCFunc func(start, stop int64) []string) RankingQuerySeter {
	r.defaultRangeDESCFunc = defaultRangeDESCFunc
	return &r
}

func (r *rankingQuerySet) callDefaultCountFunc() uint {
	if r.defaultCountFunc == nil {
		return 0
	}
	return r.defaultCountFunc()
}

func (r *rankingQuerySet) callDefaultIsMembersFunc(member string) bool {
	if r.defaultIsMembersFunc == nil {
		return false
	}
	return r.defaultIsMembersFunc(member)
}

func (r *rankingQuerySet) callDefaultRangeASCFunc(start, stop int64) []string {
	if r.defaultRangeASCFunc == nil {
		return []string{}
	}
	return r.defaultRangeASCFunc(start, stop)
}

func (r *rankingQuerySet) callDefaultRangeDESCFunc(start, stop int64) []string {
	if r.defaultRangeDESCFunc == nil {
		return []string{}
	}
	return r.defaultRangeDESCFunc(start, stop)
}

func (r *rankingQuerySet) callRebuildFunc() ([]redis.Z, time.Duration) {
	if r.rebuildFunc == nil {
		return []redis.Z{}, -1
	}
	return r.rebuildFunc()
}

func (r *rankingQuerySet) rebuild() bool {
	if r.isRebuilding {
		// 防止重构缓存失败陷入死循环
		return false
	}

	r.isRebuilding = true
	// 获取缓存重建锁
	if r.tryGetRebuildLock(r.Key()) {
		defer r.tryReleaseRebuildLock(r.Key())
		// 重建缓存
		beego.Notice("norm.TryRebuildRanking(", r.Key(), ")")
		if members, expire := r.callRebuildFunc(); len(members) > 0 {
			r.rorm.Querier().ZAddExpire(r.Key(), members, expire)
			return true
		} else {
			// 失败了，建立缓存保护盾保护DB
			if r.isProtectDB {
				r.tryProtectDB(r.Key())
			}
		}
	}
	return false
}
