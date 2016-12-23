package rorm

import "github.com/astaxie/beego"

type rankingQuerySet struct {
	*querySet
	defaultCountFunc     func() uint
	defaultIsMembersFunc func() bool
}

func (r *rankingQuerySet) Count() uint {
	// 尝试直接从缓存拿
	count := r.rorm.
		Querier().
		ZCardIfExist(r.valueCache.Key())
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

func (r rankingQuerySet) SetDefaultCountFunc(defaultCountFunc func() uint) rankingQuerySet {
	r.defaultCountFunc = defaultCountFunc
	return r
}

func (r rankingQuerySet) SetDefaultIsMembersFunc(defaultIsMembersFunc func() bool) rankingQuerySet {
	r.defaultIsMembersFunc = defaultIsMembersFunc
	return r
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

func (r *rankingQuerySet) rebuild() bool {
	// 获取缓存重建锁
	if r.tryGetRebuildLock(r.valueCache.Key()) {
		defer r.tryReleaseRebuildLock(r.valueCache.Key())
		// 重建缓存
		beego.Notice("norm.TryRebuildRanking(", r.valueCache.Key(), ")")
		if r.valueCache.RebuildFunc() {
			return true
		} else {
			// 失败了，建立缓存保护盾保护DB
			if r.isProtectDB {
				r.tryProtectDB(r.valueCache.Key())
			}
		}
	}
	return false
}
