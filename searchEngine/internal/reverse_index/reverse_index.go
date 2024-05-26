package reverseindex

import "github.com/qiancijun/trash/searchEngine/types"

type IReverseIndexer interface {
	Add(doc types.Document)                      // 添加一个 Doc
	Delete(IntId uint64, keyword *types.Keyword) // 删除
	// 搜索
}