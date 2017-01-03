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

const (
	// 表示正无穷
	InfinityPositive = "+inf"
	// 表示负无穷
	InfinityNegative = "-inf"
)

var (
	ErrorKeyNotExist    = fmt.Errorf("key not exist.")
	ErrorCanNotRebuild  = fmt.Errorf("rebuild failed.")
	ErrorMemberNotExist = fmt.Errorf("member not exist.")
)

type ROrmer interface {
	// 构造Hash查询构造器
	QueryHash(key string) HashQuerySeter
	// 构造String查询构造器
	QueryString(key string) StringQuerySeter
	// 构造ZSet查询构造器
	QueryZSet(key string) ZSetQuerySeter
	// 构造Set查询构造器
	QuerySet(key string) SetQuerySeter
	// 构造Keys查询构造器（对符合pattern的keys执行批量操作）
	QueryKeys(pattern string) KeysQuerySeter
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

// 对所有符合模式pattern的keys执行批量操作
type KeysQuerySeter interface {
	// 查找所有符合给定模式pattern（正则表达式）的 key 。
	Keys() ([]string, error)
	Del() error
}

type StringQuerySeter interface {
	QuerySeter
	// ========= 连贯操作接口 =========
	// 保护数据库
	Protect(expire time.Duration) StringQuerySeter
	// 重构String的方法
	SetRebuildFunc(rebuildFunc func() (interface{}, time.Duration)) StringQuerySeter

	// ======== 读取接口 ========
	// 获取键值
	Get() (string, error)
	// 将值写入传入实例
	Scan(value interface{}) error

	// ========= 写入接口 =========
	// 设置值（如果为实例，则调用encoding/binary接口）
	Set(value interface{}, expire time.Duration) error
	// 移除当前key
	Del() error
	// 增加指定的数值
	IncrBy(incr int64) (int64, error)
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
	// 获取指定成员的分数
	Score(member string) (float64, error)
	// 按分数升序获取排名第start到stop的所有成员
	RangeASC(start, stop int64) ([]string, error)
	// 按分数降序获取排名第start到stop的所有成员
	RangeDESC(start, stop int64) ([]string, error)
	// 按分数降序获取排名第start到stop的所有成员及成员分数
	RangeASCWithScores(start, stop int64) ([]redis.Z, error)
	// 按分数降序获取排名第start到stop的所有成员及成员分数
	RangeDESCWithScores(start, stop int64) ([]redis.Z, error)
	// 按分数升序获取指定分数区间内的成员
	// max,min除了数字外，可取"+inf"或"-inf"表示无限大或无限小
	// 默认情况下，区间的取值使用闭区间(小于等于或大于等于)，你也可以通过给参数前增加'('符号来使用可选的开区间(小于或大于)。
	// 例如：ZRANGEBYSCORE zset (1 5
	// 表示：所有符合条件 1<score<=5 的成员
	RangeByScoreASC(min, max string, offset, count int64) ([]string, error)
	// 按分数降序获取指定分数区间内的成员
	// max,min除了数字外，可取"+inf"或"-inf"表示无限大或无限小
	// 默认情况下，区间的取值使用闭区间(小于等于或大于等于)，你也可以通过给参数前增加'('符号来使用可选的开区间(小于或大于)。
	// 例如：ZREVRANGEBYSCORE zset 5 (1
	// 表示：所有符合条件 5>score>=1的成员
	RangeByScoreDESC(min, max string, offset, count int64) ([]string, error)

	// ========= 写入接口 =========
	// 向集合中增加一个成员，并设置其过期时间
	AddExpire(member interface{}, score float64, expire time.Duration) (int64, error)
	// 从集合中移除n个成员
	Rem(member ...interface{}) error
}

type SetQuerySeter interface {
	QuerySeter

	// ========= 连贯操作接口 =========
	// 保护数据库
	Protect(expire time.Duration) SetQuerySeter
	// 重构ZSet的方法
	SetRebuildFunc(rebuildFunc func() ([]interface{}, time.Duration)) SetQuerySeter

	// ========== 读取接口 ==========
	// 获取成员数量
	Count() (int64, error)
	// 获取所有成员
	Members() ([]string, error)
	// 判断目标成员是否是榜单的成员（按value判断）
	IsMember(member interface{}) (bool, error)

	// ========== 写入接口 ==========
	// 向集合中增加一个成员，并设置其过期时间
	AddExpire(member interface{}, expire time.Duration) (int64, error)
	// 从集合中移除n个成员
	Rem(member ...interface{}) error
}

type HashQuerySeter interface {
	QuerySeter

	// ========= 连贯操作接口 =========
	// 保护数据库
	Protect(expire time.Duration) HashQuerySeter
	// 重构ZSet的方法
	SetRebuildFunc(rebuildFunc func() (map[string]string, time.Duration)) HashQuerySeter

	// ========= 读取接口 =========
	Get(field string) (string, error)
	MutiGet(field ...string) ([]string, error)
	Exist(field string) (bool, error)

	// ========== 写入接口 ==========
	SetExpire(field string, value interface{}, expire time.Duration) (bool, error)
	MutiSet(kvMap map[string]string) (bool, error)
}

// 查询器接口
// 直接与Redis相连，隔离Redis与其它工具的关系
type Querier interface {
	redis.Cmdable

	// ==== String ====
	IncrByIfExist(key string, incr int64) (int64, error)

	// ==== Set ====
	SAddExpire(key string, members []interface{}, expire time.Duration) (int64, error)
	SAddExpireIfExist(key string, members []interface{}, expire time.Duration) (int64, error)
	SCardIfExist(key string) (int64, error)
	SMembersIfExist(key string) ([]string, error)
	SIsMemberIfExist(key string, member interface{}) (bool, error)

	// ==== ZSet ====
	ZAddExpire(key string, members []redis.Z, expire time.Duration) (int64, error)
	ZAddExpireIfExist(key string, members []redis.Z, expire time.Duration) (int64, error)

	ZIsMemberIfExist(key string, member string) (bool, error)
	ZCardIfExist(key string) (int64, error)
	ZScoreIfExist(key string, member string) (float64, error)

	ZRangeIfExist(key string, start, stop int64) ([]string, error)
	ZRevRangeIfExist(key string, start, stop int64) ([]string, error)

	ZRangeWithScoresIfExist(key string, start int64, stop int64) ([]redis.Z, error)
	ZRevRangeWithScoresIfExist(key string, start, stop int64) ([]redis.Z, error)

	ZRangeByScoreIfExist(key string, opt redis.ZRangeBy) ([]string, error)
	ZRevRangeByScoreIfExist(key string, opt redis.ZRangeBy) ([]string, error)

	// ==== Hash ====
	HSetExpire(key string, field string, value interface{}, expire time.Duration) (bool, error)
	HSetExpireIfExist(key string, field string, value interface{}, expire time.Duration) (bool, error)
	HMSetExpire(key string, kvMap map[string]string, expire time.Duration) (bool, error)

	HGetIfExist(key string, field string) (string, error)
}
