# 简述

对Redis的数据操作的二次封装，以应付一些常见的情况，主要核心问题在于与DB层同步。
目前依赖于：[go-redis](https://github.com/go-redis/redis)
暂时不考虑解耦，这块成本太高了。

# 特性

1. 支持同名键的方式创建数据库保护，缓解缓存穿透的问题。
1. 提供一定程度的缓存重建锁，缓解缓存并发的问题。

# 启动及注册

```golang
option := &redis.Options{
	Addr: "127.0.0.1:6379",
}
RedisClient = redis.NewClient(option)
// 默认情况下请用default注册
rorm.RegistryRedisClient("default", RedisClient)
```

# 思路

根据现在遇到的情况一点点处理吧。

## 接口：QuerySet

最基础的QuerySet接口，提供一些基础方法。

获取QuerySet的方式由于各个扩展不尽相同，这里先不罗列了。

```golang
// 缓存穿透保护（连贯操作）
// 表示对本次查询的key做180s的缓存穿透保护
qs = qs.Protect(180*time.Second)
```

## 接口：RankingQuerySet

QuerySet的Ranking扩展版，支持QuerySet的所有方法。
对ZSet的封装，主要解决有排序需求的排行榜问题。

### 基础调用

```golang
// 注册过程略
ROrmHandler = rorm.NewROrm()

// 生成RankingQuerySet
qs := ROrmHandler.QueryRanking("Key to your ZSet").

// 设置如果缓存不存在，重构缓存的方法（可选）
qs = qs.SetRebuildFunc(func() ([]redis.Z, time.Duration) {
			// 从DB读取数据(根据自己的业务情况)
			ary := FromDB()

			// 生成ZSet成员
			members := make([]redis.Z, len(ary))
			for i, v := range ary {
				members[i] = redis.Z{
					Score:  v.Score,
					Member: v.Name,
				}
			}

			// 要写入ZSet的成员及该key过期的时间
			return members, 30 *time.Seconds
		})


```

### Ranking.RangeASC(start,stop int64)[]string

根据正序，获取指定索引区间内的成员。

```golang
// 如果重构缓存失败，默认获取指定区间成员的方法（可选）
qs = qs.SetDefaultRangeASCFunc(
			func(start, stop int64) []string {
				return DB("XX").MustStringArray()
			}).
memberInRange := qs.RangeASC(start,stop)
```

### Ranking.RangeDESC(start,stop int64)[]string

根据正序，获取指定索引区间内的成员。

```golang
// 如果重构缓存失败，默认获取指定区间成员的方法（可选）
qs = qs.SetDefaultRangeDESCFunc(
			func(start, stop int64) []string {
				return DB("XX").OrderByDECS("ID").MustStringArray()
			}).
memberInRange := qs.RangeDESC(start,stop)
```

### Ranking.IsMembers(member string)

判断member是否在当前集合中。

```golang
// 设置如果重构缓存失败，判断member是否在当前集合中（可选）
qs := SetDefaultIsMembersFunc(
			func(member string) bool {
				return DB("XX").Exist()
			})

isMembers := qs.IsMembers("MEMBER")
```

### Ranking.Count()

统计Ranking中总共有多少元素。

```golang
// 设置如果重构缓存失败，获取默认数量的方法（可选）
qs = qs.SetDefaultCountFunc(func() uint {
			// 可以根据自己的业务情况实现，例如从数据库中读一下数据
			return 0
		})

// 获取Ranking中元素的数量
count := qs.Count()
```

1. DefaultCountFunc会在无法重构缓存的时候被调用，如果不设置则返回0。
1. RebuildFunc会在key不存在的时候被调用，用于重构缓存，如果不设置则会跳过重构的过程。

# 关于日志：

暂时没空处理日志的问题，先用着beego的loger，构建如果有问题，注释掉就好，不影响工作。