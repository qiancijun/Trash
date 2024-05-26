package reverseindex

import (
	"runtime"
	"sync"

	"github.com/huandu/skiplist"
	farmhash "github.com/leemcloughlin/gofarmhash"
	"github.com/qiancijun/trash/searchEngine/types"
	"github.com/qiancijun/trash/searchEngine/util"
)

type SkipListReverseIndex struct {
	table *util.ConcurrentHashMap
	locks []sync.RWMutex // 相同的 key 需要去竞争一把锁
}

type SkipListValue struct {
	Id string
	BitsFeature uint64
}

func NewSkipListReverseIndex(docNumEstimate int) *SkipListReverseIndex {
	indexer := new(SkipListReverseIndex)
	indexer.table = util.NewConcurrentHashMap(runtime.NumCPU(), docNumEstimate)
	indexer.locks = make([]sync.RWMutex, 1000)
	return indexer
}

func (indexer SkipListReverseIndex) getLock(key string) *sync.RWMutex {
	n := int(farmhash.Hash32WithSeed([]byte(key), 0))
	return &indexer.locks[n%len(indexer.locks)]
}

func (indexer *SkipListReverseIndex) Add(doc types.Document) {
	for _, keyword := range doc.Keywords {
		key := keyword.ToString()
		lock := indexer.getLock(key)
		lock.Lock()
		defer lock.Unlock()
		skValue := SkipListValue{doc.Id, doc.BitsFeature}
		if value, exists := indexer.table.Get(key); exists {
			list := value.(*skiplist.SkipList)
			list.Set(doc.IntId, skValue)
		} else {
			list := skiplist.New(skiplist.Uint64)
			list.Set(doc.IntId, skValue)
			indexer.table.Set(key, list)
		}
	}
}

func (indexer *SkipListReverseIndex) Delete(intId uint64, keyword *types.Keyword) {
	key := keyword.ToString()
	lock := indexer.getLock(key)
	lock.Lock()
	defer lock.Unlock()
	if value, exists := indexer.table.Get(key); exists {
		list := value.(*skiplist.SkipList)
		list.Remove(intId)
	}
}