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

func (s *setQuerySet) Count() int64 {
	// 尝试直接从缓存拿
	ro := s.rorm
	qr := ro.Querier()
	count, err := qr.SCardIfExist(s.Key())
	if err == nil {
		return count
	}

	// 重建缓存
	if s.rebuild() {
		// 重建成功则重新获取
		return s.Count()
	}

	// 从用户提供的默认方法获取
	return s.callDefaultCountFunc()
}

func (s *setQuerySet) Members() []string {
	// 尝试直接从缓存拿
	ro := s.rorm
	qr := ro.Querier()
	members, err := qr.SMembersIfExist(s.Key())
	if err == nil {
		return members
	}
	// 重建缓存
	if s.rebuild() {
		// 重建成功则重新获取
		return s.Members()
	}

	// 从用户提供的默认方法获取
	return s.callDefaultMembersFunc()
}

// ============= 连贯操作 =============

// 防止频繁重建
// expire 保护有效时间
func (s setQuerySet) Protect(expire time.Duration) SetQuerySeter {
	s.isProtectDB = true
	s.protectExpire = expire
	return &s
}

func (s setQuerySet) SetRebuildFunc(rebuildFunc func() ([]interface{}, time.Duration)) SetQuerySeter {
	s.rebuildFunc = rebuildFunc
	return &s
}

func (s setQuerySet) SetDefaultCountFunc(defaultCountFunc func() int64) SetQuerySeter {
	s.defaultCountFunc = defaultCountFunc
	return &s
}

func (s setQuerySet) SetDefaultIsMembersFunc(defaultIsMembersFunc func(member string) bool) SetQuerySeter {
	s.defaultIsMembersFunc = defaultIsMembersFunc
	return &s
}

// 默认获取ZSet成员的方法
func (s setQuerySet) SetDefaultMembersFunc(defaultMembersFunc func() []string) SetQuerySeter {
	s.defaultMembersFunc = defaultMembersFunc
	return &s
}

func (s *setQuerySet) callDefaultCountFunc() int64 {
	if s.defaultCountFunc == nil {
		return 0
	}
	return s.defaultCountFunc()
}

func (s *setQuerySet) callDefaultIsMembersFunc(member string) bool {
	if s.defaultIsMembersFunc == nil {
		return false
	}
	return s.defaultIsMembersFunc(member)
}

func (s *setQuerySet) callDefaultMembersFunc() []string {
	if s.defaultMembersFunc == nil {
		return []string{}
	}
	return s.defaultMembersFunc()
}

func (s *setQuerySet) callRebuildFunc() ([]interface{}, time.Duration) {
	if s.rebuildFunc == nil {
		return nil, 0
	}
	return s.rebuildFunc()
}

func (s *setQuerySet) rebuild() bool {
	if s.isRebuilding {
		// 防止重构缓存失败陷入死循环
		return false
	}

	s.isRebuilding = true
	// 获取缓存重建锁
	if s.tryGetRebuildLock(s.Key()) {
		defer s.tryReleaseRebuildLock(s.Key())
		// 重建缓存
		beego.Notice("setQuerySet.rebuild(", s.Key(), ")")
		if members, expire := s.callRebuildFunc(); len(members) > 0 {
			s.rorm.Querier().SAddExpire(s.Key(), members, expire)
			return true
		} else {
			// 失败了，建立缓存保护盾保护DB
			if s.isProtectDB {
				s.tryProtectDB(s.Key())
			}
		}
	}
	return false
}
