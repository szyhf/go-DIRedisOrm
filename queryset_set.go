package rorm

import "time"

type setQuerySet struct {
	*querySet
	rebuildFunc func() ([]interface{}, time.Duration)

	// 状态标识，防止重构缓存失败后陷入死循环
	isRebuilding bool
}
