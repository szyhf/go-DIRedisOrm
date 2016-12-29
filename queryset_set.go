package rorm

import (
	"time"

	"github.com/astaxie/beego"
)

type setQuerySet struct {
	*querySet
	rebuildFunc func() ([]interface{}, time.Duration)
}

// ========= 查询接口 =========

func (q *setQuerySet) Count() (int64, error) {
	// 尝试直接从缓存拿
	count, err := q.Querier().SCardIfExist(q.Key())
	if err == nil {
		return count, nil
	}

	// 重建缓存
	if q.rebuildingProcess(q) {
		// 重建成功则重新获取
		return q.Count()
	}

	// 从用户提供的默认方法获取
	return 0, ErrorCanNotRebuild
}

func (q *setQuerySet) Members() ([]string, error) {
	// 尝试直接从缓存拿
	members, err := q.Querier().SMembersIfExist(q.Key())
	if err == nil {
		return members, nil
	}
	// 重建缓存
	if q.rebuildingProcess(q) {
		// 重建成功则重新获取
		return q.Members()
	}

	// 从用户提供的默认方法获取
	return nil, ErrorCanNotRebuild
}

func (q *setQuerySet) IsMember(member interface{}) (bool, error) {
	val, err := q.Querier().SIsMemberIfExist(q.Key(), member)
	if err == nil {
		return val, nil
	}

	// rebuild cache
	if q.rebuildingProcess(q) {
		return q.IsMember(member)
	}

	return false, ErrorCanNotRebuild
}

func (q *setQuerySet) Rem(member ...interface{}) error {
	cmd := q.Querier().SRem(q.Key(), member...)
	return cmd.Err()
}

// ========= 写入接口 =========
func (q *setQuerySet) AddExpire(member interface{}, expire time.Duration) (int64, error) {
	// 如果不增加过期方法，可能会创建一个不会过期的集合
	num, err := q.Querier().
		SAddExpireIfExist(q.Key(), []interface{}{member}, expire)
	if err == nil {
		return num, nil
	}

	if q.rebuildingProcess(q) {
		return q.AddExpire(member, expire)
	}

	return 0, ErrorCanNotRebuild
}

// ============= 连贯操作 =============

// 防止频繁重建
// expire 保护有效时间
func (q setQuerySet) Protect(expire time.Duration) SetQuerySeter {
	q.isProtectDB = true
	q.protectExpire = expire
	return &q
}

func (q setQuerySet) SetRebuildFunc(rebuildFunc func() ([]interface{}, time.Duration)) SetQuerySeter {
	q.rebuildFunc = rebuildFunc
	return &q
}

func (q *setQuerySet) callRebuildFunc() ([]interface{}, time.Duration) {
	if q.rebuildFunc == nil {
		return nil, -1
	}
	return q.rebuildFunc()
}

func (q *setQuerySet) Rebuilding() error {
	// 重建缓存
	beego.Notice("setQuerySet.rebuild(", q.Key(), ")")
	if members, expire := q.callRebuildFunc(); len(members) > 0 {
		// 见 issue#1
		cmd := q.Querier().Del(q.Key())
		if cmd.Err() == nil {
			_, err := q.Querier().SAddExpire(q.Key(), members, expire)
			return err
		}
	}
	return ErrorCanNotRebuild
}
