package rorm

type keysQuerySet struct {
	*querySet
}

func (q *keysQuerySet) Keys() ([]string, error) {
	cmd := q.Querier().Keys(q.Key())
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Val(), nil
}

// 移除符合模式的所有key
func (q *keysQuerySet) Del() error {
	keys, _ := q.Keys()
	cmd := q.Querier().Del(keys...)
	return cmd.Err()
}
