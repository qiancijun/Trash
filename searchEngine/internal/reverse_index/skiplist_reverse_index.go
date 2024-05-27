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

func (indexer SkipListReverseIndex) FilterByBits(bits uint64, onFlag uint64, offFlag uint64, orFlags []uint64) bool {
	// on 必须全部命中
	if bits & onFlag != onFlag {
		return false
	}
	// off 所有 bit 必须全部不命中
	if bits & offFlag != 0 {
		return false
	}
	for _, orFlag := range orFlags {
		if orFlag > 0 && bits & orFlag <= 0 {
			return false
		}
	}
	return true
}

func (indexer SkipListReverseIndex) search(q *types.TermQuery,
onFlag uint64, offFlag uint64, orFlags []uint64) *skiplist.SkipList {
	if q.Keyword != nil {
		Keyword := q.Keyword.ToString()
		if value, exists := indexer.table.Get(Keyword); exists {
			result := skiplist.New(skiplist.Uint64)
			list := value.(*skiplist.SkipList)
			node := list.Front()
			for node != nil {
				intId := node.Key().(uint64)
				skv, _ := node.Value.(SkipListValue)
				flag := skv.BitsFeature
				if intId > 0 && indexer.FilterByBits(flag, onFlag, offFlag, orFlags) { //确保有效元素都大于0
					result.Set(intId, skv)
				}
				node = node.Next()
			}
			return result
		}
	} else if len(q.Must) > 0 {
		results := make([]*skiplist.SkipList, 0, len(q.Must))
		for _, q := range q.Must {
			results = append(results, indexer.search(q, onFlag, offFlag, orFlags))
		}
		return util.IntersectionOfSkipList(results...)
	} else if len(q.Should) > 0 {
		results := make([]*skiplist.SkipList, 0, len(q.Should))
		for _, q := range q.Should {
			results = append(results, indexer.search(q, onFlag, offFlag, orFlags))
		}
		return util.UnionsetOfSkipList(results...)
	}
	return nil
}