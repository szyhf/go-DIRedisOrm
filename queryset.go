package rorm // import "go.szyhf.org/di-rorm"
import (
	"time"

	"github.com/astaxie/beego"
)

type querySet struct {
	rorm          ROrmer
	valueCache    ValueCacher
	isProtectDB   bool
	protectExpire time.Duration
}

// 防止频繁重建
// expire 保护有效时间
func (this querySet) Protect(expire time.Duration) querySet {
	this.isProtectDB = true
	this.protectExpire = expire
	return this
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
