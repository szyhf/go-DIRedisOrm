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
	val, err := q.Querier().HGetIfExist(q.Key(), field)
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
func (q *hashQuerySet) SetExpire(field string, value interface{}, expire time.Duration) (bool, error) {
	// 如果不增加过期方法，可能会创建一个不会过期的集合
	ok, err := q.Querier().
		HSetExpireIfExist(q.Key(), field, value, expire)
	if err == nil {
		return ok, nil
	}

	if q.rebuildingProcess(q) {
		return q.SetExpire(field, value, expire)
	}

	return false, ErrorCanNotRebuild
}

func (q *hashQuerySet) Del(field string) error {
	cmd := q.Querier().HDel(q.Key(), field)
	return cmd.Err()
}

func (q *hashQuerySet) MutiSet(kvMap map[string]string) (bool, error) {
	panic("Not imp")
}

// ========== 连贯操作 ==========

// 重构Hash的方法
func (q hashQuerySet) SetRebuildFunc(rebuildFunc func() (map[string]string, time.Duration)) HashQuerySeter {
	q.rebuildFunc = rebuildFunc
	return &q
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
