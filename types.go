package rorm // import "go.szyhf.org/di-rorm"
import redis "gopkg.in/redis.v5"
import "time"
import "fmt"

const (
	OrderASC uint8 = iota
	OrderDESC
)

var (
	ErrorKeyNotExist = fmt.Errorf("key not exist.")
)

type ROrmer interface {
	QueryZSet(key string) ZSetQuerySeter
	Using(alias string) ROrmer
	Querier() *RedisQuerier
}

type QuerySeter interface {
	Key() string
}

type ZSetQuerySeter interface {
	// ========= 连贯操作接口 =========
	// 保护数据库
	Protect(expire time.Duration) ZSetQuerySeter
	// 重构ZSet的方法
	SetRebuildFunc(rebuildFunc func() ([]redis.Z, time.Duration)) ZSetQuerySeter
	// 默认获取ZSet数量的方法
	SetDefaultCountFunc(defaultCountFunc func() uint) ZSetQuerySeter
	// 默认判断目标是否ZSet成员的方法
	SetDefaultIsMembersFunc(defaultIsMembersFunc func(member string) bool) ZSetQuerySeter
	// 默认获取ZSet某区段成员的方法
	SetDefaultRangeASCFunc(defaultRangeASC func(start, stop int64) []string) ZSetQuerySeter
	// 默认获取ZSet某区段成员的方法
	SetDefaultRangeDESCFunc(defaultRangeDESC func(start, stop int64) []string) ZSetQuerySeter

	// ========= 查询接口 =========
	// 获取成员数量
	Count() uint
	// 按分数升序获取排名第start到stop的所有成员
	RangeASC(start, stop int64) []string
	// 按分数降序获取排名第start到stop的所有成员
	RangeDESC(start, stop int64) []string
	// 判断目标成员是否是榜单的成员（按value判断）
	IsMembers(member string) bool

	// ========= 写入接口 =========
	// 向集合中增加一个成员，并设置其过期时间
	AddExpire(member interface{}, score float64, expire time.Duration) error
	// 从集合中移除n个成员
	Rem(member ...interface{}) error
}

type SetQuerySeter interface {
}

type Querier interface {
	redis.Cmdable
	ZAddExpire(key string, members []redis.Z, expire time.Duration) error
	ZCardIfExist(key string) (int64, error)
	ZIsMembers(key string, member string) (bool, error)
}
