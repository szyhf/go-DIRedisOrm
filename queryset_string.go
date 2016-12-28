package rorm

import (
	"time"

	"github.com/astaxie/beego"
)

type stringQuerySet struct {
	*querySet
	rebuildFunc func() (interface{}, time.Duration)
}

// ======== 读取接口 ========

func (q *stringQuerySet) Get() (string, error) {
	// 尝试直接从缓存获取
	cmd := q.Querier().Get(q.Key())
	if cmd.Err() == nil {
		return cmd.Val(), nil
	}

	// 尝试重建缓存
	if q.rebuildingProcess(q) {
		return q.Get()
	}

	// 尝试从默认方法获取
	return "", ErrorCanNotRebuild
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

	return ErrorCanNotRebuild
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

// 如果key存在，则给当前key增长指定的值
func (q *stringQuerySet) IncrBy(incr int64) (int64, error) {
	val, err := q.Querier().
		IncrByIfExist(q.Key(), incr)
	if err == nil {
		return val, nil
	}

	if q.rebuildingProcess(q) {
		return q.IncrBy(incr)
	}

	return 0, ErrorCanNotRebuild
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

// ========= 辅助方法 =========

func (q *stringQuerySet) Rebuilding() error {
	// 重建缓存
	beego.Notice("stringQuerySet.rebuild(", q.Key(), ")")
	if value, expire := q.callRebuildFunc(); value != nil {
		cmd := q.Querier().Set(q.Key(), value, expire)
		return cmd.Err()
	}
	return ErrorCanNotRebuild
}

func (q *stringQuerySet) callRebuildFunc() (interface{}, time.Duration) {
	if q.rebuildFunc == nil {
		return nil, -1
	}
	return q.rebuildFunc()
}
