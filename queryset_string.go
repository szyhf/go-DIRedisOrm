package rorm

import (
	"time"

	"github.com/astaxie/beego"
)

type stringQuerySet struct {
	*querySet
	rebuildFunc    func() (interface{}, time.Duration)
	defaultGetFunc func() string
}

// ========= 连贯操作接口 =========
// 保护数据库
func (q stringQuerySet) Protect(expire time.Duration) StringQuerySeter {
	q.isProtectDB = true
	q.protectExpire = expire
	return &q
}

// 重构String的方法
func (q stringQuerySet) SetRebuildFunc(rebuildFunc func() (interface{}, time.Duration)) StringQuerySeter {
	q.rebuildFunc = rebuildFunc
	return &q
}

func (q stringQuerySet) SetDefaultGetFunc(getFunc func() string) StringQuerySeter {
	q.defaultGetFunc = getFunc
	return &q
}

// ======== 读取接口 ========

func (q *stringQuerySet) Get() string {
	// 尝试直接从缓存获取
	ro := q.rorm
	qr := ro.Querier()
	cmd := qr.Get(q.Key())
	if cmd.Err() == nil {
		return cmd.Val()
	}

	// 尝试重建缓存
	if q.rebuildingProcess(q) {
		return q.Get()
	}

	// 尝试从默认方法获取
	return q.callDefaultGetFunc()
}
func (q *stringQuerySet) Scan(value interface{}) error {
	panic("Not imp")
}

// ========= 写入接口 =========
func (q *stringQuerySet) Set(value interface{}, expire time.Duration) error {
	panic("Not imp")
}

// ========= 辅助方法 =========
func (q *stringQuerySet) callDefaultGetFunc() string {
	if q.defaultGetFunc == nil {
		return ""
	}
	return q.defaultGetFunc()
}

func (q *stringQuerySet) Rebuilding() error {
	// 重建缓存
	beego.Notice("stringQuerySet.rebuild(", q.Key(), ")")
	if value, expire := q.rebuildFunc(); value != nil {
		cmd := q.rorm.Querier().Set(q.Key(), value, expire)
		if cmd.Err() == nil {
			return nil
		} else {
			return cmd.Err()
		}
	}
	return ErrorCanNotRebuild
}
