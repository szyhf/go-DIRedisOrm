package rorm
import (
	"time"

	"github.com/astaxie/beego"
)

type querySet struct {
	rorm ROrmer

	key           string
	isProtectDB   bool
	protectExpire time.Duration

	// 状态标识，防止重构缓存失败后陷入死循环
	isRebuilding bool
}

// 防止频繁重建
// expire 保护有效时间
func (q querySet) Protect(expire time.Duration) QuerySeter {
	q.isProtectDB = true
	q.protectExpire = expire
	return &q
}

func (q *querySet) Key() string {
	return q.key
}

func (q *querySet) Rebuilding() error {
	panic("Should implement method \"Rebuilding\" in sub queryset")
}

func (q *querySet) Querier() Querier {
	return q.ROrmer().Querier()
}

func (q *querySet) ROrmer() ROrmer {
	return q.rorm
}

func (q *querySet) tryGetRebuildLock(key string) bool {
	beego.Notice("tryGetRebuildLock:", key)
	// 通过setNX设置锁，同设置超时，防止del失败
	if cmd := q.Querier().SetNX(key+":mutex", "", 30*time.Second); cmd.Err() == nil {
		return cmd.Val()
	} else {
		beego.Warn("querySet.TryGetRebuildLock(", key, ") failed: ", cmd.Err())
	}
	return false
}

func (q *querySet) tryReleaseRebuildLock(key string) bool {
	beego.Notice("tryReleaseRebuildLock:", key)
	if cmd := q.Querier().Del(key + ":mutex"); cmd.Err() == nil {
		return true
	} else {
		beego.Warn("querySet.TryReleaseRebuildLock(", key, ") failed: ", cmd.Err())
	}

	return false
}

func (q *querySet) tryProtectDB(key string) bool {
	cmd := q.Querier().Set(key, nil, q.protectExpire)
	beego.Notice("tryProtectDB:", key, "for", q.protectExpire, "seconds.")
	return cmd.Err() == nil
}

func (q *querySet) rebuildingProcess(qs QuerySeter) bool {
	if q.isRebuilding {
		beego.Warn("Rebuilding break for dead loop.")
		// 防止重构缓存失败陷入死循环
		return false
	}

	// 获取缓存重建锁
	if q.tryGetRebuildLock(q.Key()) {
		q.isRebuilding = true
		defer q.tryReleaseRebuildLock(q.Key())
		if err := qs.Rebuilding(); err != nil {
			// 失败了，建立缓存保护盾保护DB
			if q.isProtectDB {
				q.tryProtectDB(q.Key())
			}
		} else {
			return true
		}
	}
	return false
}
