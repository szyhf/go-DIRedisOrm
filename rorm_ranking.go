package rorm // import "go.szyhf.org/di-rorm"

var rankingRegistry = make(map[string]rankingCache)

type rankingCache struct {
	key         string
	rebuildFunc func() bool
}

func RegisterRanking(key string, rebuildFunc func() bool) {
	rankingRegistry[key] = rankingCache{
		key:         key,
		rebuildFunc: rebuildFunc,
	}
}

func (r *rankingCache) Key() string {
	return r.key
}

func (r *rankingCache) RebuildFunc() bool {
	if r.rebuildFunc == nil {
		return false
	}
	return r.rebuildFunc()
}
