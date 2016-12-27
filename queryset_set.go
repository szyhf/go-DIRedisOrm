package rorm

import (
	"time"

	"github.com/astaxie/beego"
)

type setQuerySet struct {
	*querySet
	rebuildFunc          func() ([]interface{}, time.Duration)
	defaultCountFunc     func() int64
	defaultIsMembersFunc func(string) bool
	defaultMembersFunc   func() []string
	// 状态标识，防止重构缓存失败后陷入死循环
	isRebuilding bool
}

// ========= 查询接口 =========

func (q *setQuerySet) Count() int64 {
	// 尝试直接从缓存拿
	count, err := q.Querier().SCardIfExist(q.Key())
	if err == nil {
		return count
	}

	// 重建缓存
	if q.rebuildingProcess(q) {
		// 重建成功则重新获取
		return q.Count()
	}

	// 从用户提供的默认方法获取
	return q.callDefaultCountFunc()
}

func (q *setQuerySet) Members() []string {
	// 尝试直接从缓存拿
	members, err := q.Querier().SMembersIfExist(q.Key())
	if err == nil {
		return members
	}
	// 重建缓存
	if q.rebuildingProcess(q) {
		// 重建成功则重新获取
		return q.Members()
	}

	// 从用户提供的默认方法获取
	return q.callDefaultMembersFunc()
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

func (q setQuerySet) SetDefaultCountFunc(defaultCountFunc func() int64) SetQuerySeter {
	q.defaultCountFunc = defaultCountFunc
	return &q
}

func (q setQuerySet) SetDefaultIsMembersFunc(defaultIsMembersFunc func(member string) bool) SetQuerySeter {
	q.defaultIsMembersFunc = defaultIsMembersFunc
	return &q
}

// 默认获取ZSet成员的方法
func (q setQuerySet) SetDefaultMembersFunc(defaultMembersFunc func() []string) SetQuerySeter {
	q.defaultMembersFunc = defaultMembersFunc
	return &q
}

func (q *setQuerySet) callDefaultCountFunc() int64 {
	if q.defaultCountFunc == nil {
		return 0
	}
	return q.defaultCountFunc()
}

func (q *setQuerySet) callDefaultIsMembersFunc(member string) bool {
	if q.defaultIsMembersFunc == nil {
		return false
	}
	return q.defaultIsMembersFunc(member)
}

func (q *setQuerySet) callDefaultMembersFunc() []string {
	if q.defaultMembersFunc == nil {
		return []string{}
	}
	return q.defaultMembersFunc()
}

func (q *setQuerySet) callRebuildFunc() ([]interface{}, time.Duration) {
	if q.rebuildFunc == nil {
		return nil, 0
	}
	return q.rebuildFunc()
}

func (q *setQuerySet) Rebuilding() error {
	// 重建缓存
	beego.Notice("setQuerySet.rebuild(", q.Key(), ")")
	if members, expire := q.callRebuildFunc(); len(members) > 0 {
		return q.rorm.Querier().SAddExpire(q.Key(), members, expire)
	}
	return ErrorCanNotRebuild
}
