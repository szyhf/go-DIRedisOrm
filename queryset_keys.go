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

// 移除当前key
func (q *keysQuerySet) Del() error {
	cmd := q.Querier().Del(q.Key())
	return cmd.Err()
}
