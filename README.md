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

# 更多文档参考wiki目录

1. [Main.md](./wiki/1.Main.md)
2. [Ranking.md](./wiki/2.Ranking.md)

# 关于日志：

暂时没空处理日志的问题，先用着beego的loger，构建如果有问题，注释掉就好，不影响工作。