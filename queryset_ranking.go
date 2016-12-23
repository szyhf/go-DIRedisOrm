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
	defaultIsMembersFunc func() bool
}

func (r *rankingQuerySet) Count() uint {
	// 尝试直接从缓存拿
	ro := r.rorm
	qr := ro.Querier()
	count := qr.ZCardIfExist(r.Key())
	if count >= 0 {
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

func (r *rankingQuerySet) IsMembers(member interface{}) bool {
	panic("Not imp")
}

func (r rankingQuerySet) SetDefaultCountFunc(defaultCountFunc func() uint) RankingQuerySeter {
	r.defaultCountFunc = defaultCountFunc
	return &r
}

func (r rankingQuerySet) SetDefaultIsMembersFunc(defaultIsMembersFunc func() bool) RankingQuerySeter {
	r.defaultIsMembersFunc = defaultIsMembersFunc
	return &r
}

func (r rankingQuerySet) SetRebuildFunc(rebuildFunc func() ([]redis.Z, time.Duration)) RankingQuerySeter {
	r.rebuildFunc = rebuildFunc
	return &r
}

func (r *rankingQuerySet) callDefaultCountFunc() uint {
	if r.defaultCountFunc == nil {
		return 0
	}
	return r.defaultCountFunc()
}

func (r *rankingQuerySet) callDefaultIsMembersFunc() bool {
	if r.defaultIsMembersFunc == nil {
		return false
	}
	return true
}

func (r *rankingQuerySet) callRebuildFunc() ([]redis.Z, time.Duration) {
	if r.rebuildFunc == nil {
		return []redis.Z{}, -1
	}
	return r.rebuildFunc()
}

func (r *rankingQuerySet) rebuild() bool {
	// 获取缓存重建锁
	if r.tryGetRebuildLock(r.Key()) {
		defer r.tryReleaseRebuildLock(r.Key())
		// 重建缓存
		beego.Notice("norm.TryRebuildRanking(", r.Key(), ")")
		if members, expire := r.callRebuildFunc(); len(members) > 0 {
			// TODO: 临时处理过期时间
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
