package index_service

import (
	"bytes"
	"encoding/gob"
	"strings"
	"sync/atomic"

	kvdb "github.com/qiancijun/trash/searchEngine/internal/kv_db"
	reverseindex "github.com/qiancijun/trash/searchEngine/internal/reverse_index"
	"github.com/qiancijun/trash/searchEngine/types"
	"github.com/qiancijun/trash/searchEngine/util"
)

// 外观模式，把正排和倒排封装到了一起
type Indexer struct {
	forwardIndex kvdb.IKeyValueDB
	reverseIndex reverseindex.IReverseIndexer
	maxIntId uint64
}

var _ IIndexer = (*Indexer)(nil)

func (indexer *Indexer) Init(docNumEstimate int, dbtype int, dataDir string) error {
	db, err := kvdb.GetKvDB(dbtype, dataDir)
	if err != nil {
		return err
	}
	indexer.forwardIndex = db
	indexer.reverseIndex = reverseindex.NewSkipListReverseIndex(docNumEstimate)
	return nil
}

// 系统重启时，直接从索引文件里加载数据
func (indexer *Indexer) LoadFromIndexFile() int {
	reader := bytes.NewReader([]byte{})
	n := indexer.forwardIndex.IterDB(func (k, v []byte) error {
		reader.Reset(v)
		decoder := gob.NewDecoder(reader)
		var doc types.Document
		err := decoder.Decode(&doc)
		if err != nil {
			util.Log.Printf("gob decode document failed: %v", err)
			return nil
		}	
		indexer.reverseIndex.Add(doc)
		return err
	})
	util.Log.Printf("load %d data from forward index %s", n, indexer.forwardIndex.GetDbPath())
	return int(n)
}

func (indexer *Indexer) Close() error {
	return indexer.forwardIndex.Close()
}

// 更新与添加操作
func (indexer *Indexer) AddDoc(doc types.Document) (int, error) {
	docId := strings.TrimSpace(doc.Id)
	if len(docId) == 0 {
		return 0, nil
	}

	// 先从正排和倒排索引上删除 doc
	indexer.DeleteDoc(docId)
	
	doc.IntId = atomic.AddUint64(&indexer.maxIntId, 1)
	// 写入正排索引
	var value bytes.Buffer
	encoder := gob.NewEncoder(&value)
	if err := encoder.Encode(doc); err == nil {
		indexer.forwardIndex.Set([]byte(docId), value.Bytes())
	} else {
		return 0, err
	}
	// 写入倒排索引
	indexer.reverseIndex.Add(doc)
	return 1, nil
}


func (indexer *Indexer) DeleteDoc(docId string) int {
	n := 0
	forwardKey := []byte(docId)
	docBs, err := indexer.forwardIndex.Get(forwardKey)
	if err == nil {
		reader := bytes.NewReader([]byte{})
		if len(docBs) > 0 {
			n = 1
			reader.Reset(docBs)
			decoder := gob.NewDecoder(reader)
			var doc types.Document
			err := decoder.Decode(&doc)
			if err != nil {
				// 遍历每一个 keyword，从倒排索引上删除
				for _, kw := range doc.Keywords {
					indexer.reverseIndex.Delete(doc.IntId, kw)
				}
			}
		}
	}
	indexer.forwardIndex.Delete(forwardKey)
	return n
}

func (indexer *Indexer) Search(query *types.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*types.Document {
	docIds := indexer.reverseIndex.Search(query, onFlag, offFlag, orFlags)
	if len(docIds) == 0 {
		return nil
	}
	keys := make([][]byte, 0, len(docIds))
	for _, docId := range docIds {
		keys = append(keys, []byte(docId))
	}
	docs, err := indexer.forwardIndex.BatchGet(keys)
	if err != nil {
		util.Log.Printf("read kvdb failed: %s", err)
		return nil
	}
	result := make([]*types.Document, 0, len(docs))
	reader := bytes.NewReader([]byte{})
	for _, docBs := range docs {
		if len(docBs) > 0 {
			reader.Reset(docBs)
			decoder := gob.NewDecoder(reader)
			var doc types.Document
			if err := decoder.Decode(&doc); err == nil {
				result = append(result, &doc)
			}
		}
	}
	return result
}

func (indexer *Indexer) Count() int {
	n := 0
	indexer.forwardIndex.IterKey(func(k []byte) error {
		n++
		return nil
	})
	return n
}