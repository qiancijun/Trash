package index_service

import "github.com/qiancijun/trash/searchEngine/types"

type IIndexer interface {
	AddDoc(doc types.Document) (int, error)
	DeleteDoc(docId string) int
	Search(query *types.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*types.Document
	Count() int
	Close() error
}