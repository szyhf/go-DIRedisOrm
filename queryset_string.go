package rorm

import (
	"time"

	"github.com/astaxie/beego"
)

type stringQuerySet struct {
	*querySet
	rebuildFunc func() (interface{}, time.Duration)

	defaultGetFunc  func() string
	defaultScanFunc func(interface{}) error
}

// ======== 读取接口 ========

func (q *stringQuerySet) Get() string {
	// 尝试直接从缓存获取
	cmd := q.Querier().Get(q.Key())
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
	// 尝试直接从缓存获取
	cmd := q.Querier().Get(q.Key())
	if cmd.Err() == nil {
		return cmd.Scan(value)
	}

	if q.rebuildingProcess(q) {
		return q.Scan(value)
	}

	return q.defaultScanFunc(value)
}

// ========= 写入接口 =========
// 写入当前key
func (q *stringQuerySet) Set(value interface{}, expire time.Duration) error {
	cmd := q.Querier().Set(q.Key(), value, expire)
	return cmd.Err()
}

// 移除当前key
func (q *stringQuerySet) Del() error {
	cmd := q.Querier().Del(q.Key())
	return cmd.Err()
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

func (q stringQuerySet) SetDefaultScanFunc(scanFunc func(val interface{}) error) StringQuerySeter {
	q.defaultScanFunc = scanFunc
	return &q
}

// ========= 辅助方法 =========

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

func (q *stringQuerySet) callDefaultScanFunc(val interface{}) error {
	if q.defaultScanFunc == nil {
		return nil
	}
	return q.defaultScanFunc(val)
}

func (q *stringQuerySet) callDefaultGetFunc() string {
	if q.defaultGetFunc == nil {
		return ""
	}
	return q.defaultGetFunc()
}
