package rorm // import "go.szyhf.org/di-rorm"
import (
	"fmt"
	"time"

	redis "gopkg.in/redis.v5"
)

const (
	OrderASC uint8 = iota
	OrderDESC
)

var (
	ErrorKeyNotExist   = fmt.Errorf("key not exist.")
	ErrorCanNotRebuild = fmt.Errorf("rebuild failed.")
)

type ROrmer interface {
	// 构造String查询构造器
	QueryString(key string) StringQuerySeter
	// 构造ZSet查询构造器
	QueryZSet(key string) ZSetQuerySeter
	// 构造Set查询构造器
	QuerySet(key string) SetQuerySeter
	// 设置使用的Redis链接
	Using(alias string) ROrmer
	// 当前生效的查询器
	Querier() Querier
}

type QuerySeter interface {
	// 获取当前查询的key
	Key() string
	// 重构缓存的接口
	Rebuilding() error
	// 查询器的引用
	Querier() Querier
	// ROrmer的引用
	ROrmer() ROrmer
}

type StringQuerySeter interface {
	QuerySeter
	// ========= 连贯操作接口 =========
	// 保护数据库
	Protect(expire time.Duration) StringQuerySeter
	// 重构String的方法
	SetRebuildFunc(rebuildFunc func() (interface{}, time.Duration)) StringQuerySeter
	// 设置默认的扫描方法
	SetDefaultScanFunc(scanFunc func(val interface{}) error) StringQuerySeter
	// ======== 读取接口 ========
	// 获取键值
	Get() string
	// 将值写入传入实例
	Scan(value interface{}) error

	// ========= 写入接口 =========
	// 设置值（如果为实例，则调用encoding/binary接口）
	Set(value interface{}, expire time.Duration) error
	// 移除当前key
	Del() error
}

type ZSetQuerySeter interface {
	QuerySeter
	// ========= 连贯操作接口 =========
	// 保护数据库
	Protect(expire time.Duration) ZSetQuerySeter
	// 重构ZSet的方法
	SetRebuildFunc(rebuildFunc func() ([]redis.Z, time.Duration)) ZSetQuerySeter

	// ========= 查询接口 =========
	// 判断目标成员是否是榜单的成员（按value判断）
	IsMember(member string) (bool, error)
	// 获取成员数量
	Count() (int64, error)
	// 获取所有成员
	Members() ([]string, error)
	// 按分数升序获取排名第start到stop的所有成员
	RangeASC(start, stop int64) ([]string, error)
	// 按分数降序获取排名第start到stop的所有成员
	RangeDESC(start, stop int64) ([]string, error)

	// ========= 写入接口 =========
	// 向集合中增加一个成员，并设置其过期时间
	AddExpire(member interface{}, score float64, expire time.Duration) error
	// 从集合中移除n个成员
	Rem(member ...interface{}) error
}

type SetQuerySeter interface {
	QuerySeter
	Protect(expire time.Duration) SetQuerySeter
	SetRebuildFunc(rebuildFunc func() ([]interface{}, time.Duration)) SetQuerySeter
	SetDefaultMembersFunc(defaultMembersFunc func() []string) SetQuerySeter
	SetDefaultIsMemberFunc(defaultIsMemberFunc func(member interface{}) bool) SetQuerySeter

	// ========== 读取接口 ==========
	Count() int64
	Members() []string
	IsMember(member interface{}) bool

	// ========== 写入接口 ==========
	Rem(member ...interface{}) error
}

// 查询器接口
// 直接与Redis相连，隔离Redis与其它工具的关系
type Querier interface {
	redis.Cmdable

	// ==== Set ====
	SCardIfExist(key string) (int64, error)
	SAddExpire(key string, members []interface{}, expire time.Duration) error
	SMembersIfExist(key string) ([]string, error)
	SIsMemberIfExist(key string, member interface{}) (bool, error)

	// ==== ZSet ====
	ZAddExpire(key string, members []redis.Z, expire time.Duration) error
	ZCardIfExist(key string) (int64, error)
	ZIsMember(key string, member string) (bool, error)
	ZRangeIfExist(key string, start, stop int64) ([]string, error)
	ZRevRangeIfExist(key string, start, stop int64) ([]string, error)
}
