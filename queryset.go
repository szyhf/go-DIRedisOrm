package rorm // import "go.szyhf.org/di-rorm"
import (
	"time"

	"github.com/astaxie/beego"
)

type querySet struct {
	rorm ROrmer

	key           string
	isProtectDB   bool
	protectExpire time.Duration
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

func (this *querySet) tryGetRebuildLock(key string) bool {
	// 通过setNX设置锁，同设置超时，防止del失败
	if cmd := this.rorm.Querier().SetNX(key+":mutex", "", 30*time.Second); cmd.Err() == nil {
		return cmd.Val()
	} else {
		beego.Warn("querySet.TryGetRebuildLock(", key, ") failed: ", cmd.Err())
	}
	return false
}

func (this *querySet) tryReleaseRebuildLock(key string) bool {
	if cmd := this.rorm.Querier().Del(key + ":mutex"); cmd.Err() == nil {
		return true
	} else {
		beego.Warn("querySet.TryReleaseRebuildLock(", key, ") failed: ", cmd.Err())
	}

	return false
}

func (this *querySet) tryProtectDB(key string) bool {
	cmd := this.rorm.Querier().Set(key, nil, this.protectExpire)
	return cmd.Err() == nil
}
