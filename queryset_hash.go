package rorm

import (
	"time"

	"github.com/astaxie/beego"
)

type hashQuerySet struct {
	*querySet
	rebuildFunc func() (map[string]string, time.Duration)
}

// ========= 读取接口 =========
func (q *hashQuerySet) Get(field string) (string, error) {
	val, err := q.Querier().HGetIfExsit(q.Key(), field)
	if err == nil {
		return val, nil
	}

	// 重构缓存
	if q.rebuildingProcess(q) {
		return q.Get(field)
	}

	return "", ErrorCanNotRebuild
}
func (q *hashQuerySet) MutiGet(field ...string) ([]string, error) {
	panic("Not imp")
}
func (q *hashQuerySet) Exist(field string) (bool, error) {
	panic("Not imp")
}

// ========== 写入接口 ==========
func (q *hashQuerySet) Set(field string, value interface{}) (int64, error) {
	panic("Not imp")
}
func (q *hashQuerySet) MutiSet(kvMap map[string]string) (int64, error) {
	panic("Not imp")
}

// ========== 连贯操作 ==========

// 重构ZSet的方法
func (q hashQuerySet) SetRebuildFunc(rebuildFunc func() (map[string]string, time.Duration)) HashQuerySeter {
	panic("Not imp")
}

func (q hashQuerySet) Protect(expire time.Duration) HashQuerySeter {
	q.isProtectDB = true
	q.protectExpire = expire
	return &q
}

// ========== 辅助操作 ==========

func (q *hashQuerySet) Rebuilding() error {
	// 重建缓存
	beego.Notice("hashQuerySet.Rebuilding( ", q.Key(), " )")
	if kvMap, expire := q.callRebuildFunc(); len(kvMap) > 0 {
		// 见 issue#1
		cmd := q.Querier().Del(q.Key())
		if cmd.Err() == nil {
			_, err := q.Querier().HMSetExpire(q.Key(), kvMap, expire)
			return err
		}
		return cmd.Err()
	}
	return ErrorCanNotRebuild
}

func (q *hashQuerySet) callRebuildFunc() (map[string]string, time.Duration) {
	if q.rebuildFunc == nil {
		return nil, -1
	}
	return q.rebuildFunc()
}
