# 简述

对Redis的数据操作的二次封装，以应付一些常见的情况，主要核心问题在于与DB层同步。
目前依赖于：[go-redis](https://github.com/go-redis/redis)
暂时不考虑解耦，这块成本太高了。

# 特性

1. 支持同名键的方式创建数据库保护，缓解缓存穿透的问题。
1. 提供一定程度的缓存重建锁，缓解缓存并发的问题。

# 启动及注册

```
option := &redis.Options{
	Addr: "127.0.0.1:6379",
}
RedisClient = redis.NewClient(option)
// 默认情况下请用default注册
rorm.RegistryRedisClient("default", RedisClient)
```

# 思路

根据现在遇到的情况一点点处理吧。

## 结构：Ranking

对ZSet的封装，主要解决有排序需求的排行榜问题。

### Ranking.Count()

统计Ranking中总共有多少元素。

```golang
// 注册过程略
ROrmHandler = rorm.NewROrm()
// 采用链式操作，SetXXX方法都是可选的
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

// 设置如果重构缓存失败，获取默认数量的方法（可选）
qs = qs.SetDefaultCountFunc(func() uint {
			// 可以根据自己的业务情况实现，例如从数据库中读一下数据
			return 0
		})

count = qs.Count()
```

1. DefaultCountFunc会在无法重构缓存的时候被调用，如果不设置则返回0。
1. RebuildFunc会在key不存在的时候被调用，用于重构缓存，如果不设置则会跳过重构的过程。

# 关于日志：

暂时没空处理日志的问题，先用着beego的loger，构建如果有问题，注释掉就好，不影响工作。