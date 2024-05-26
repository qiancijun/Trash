package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/qiancijun/trash/searchEngine/util"
	"github.com/huandu/skiplist"
)

func TestIntersectionOfSkipList(t *testing.T) {
	l1 := skiplist.New(skiplist.Uint64)
	l1.Set(uint64(5), 0)
	l1.Set(uint64(1), 0)
	l1.Set(uint64(4), 0)
	l1.Set(uint64(9), 0)
	l1.Set(uint64(11), 0)
	l1.Set(uint64(7), 0)
	//skiplist内部会自动做排序，排完序之后为 1 4 5 7 9 11

	l2 := skiplist.New(skiplist.Uint64)
	l2.Set(uint64(4), 0)
	l2.Set(uint64(5), 0)
	l2.Set(uint64(9), 0)
	l2.Set(uint64(8), 0)
	l2.Set(uint64(2), 0)
	//skiplist内部会自动做排序，排完序之后为 2 4 5 8 9

	l3 := skiplist.New(skiplist.Uint64)
	l3.Set(uint64(3), 0)
	l3.Set(uint64(5), 0)
	l3.Set(uint64(7), 0)
	l3.Set(uint64(9), 0)
	//skiplist内部会自动做排序，排完序之后为 3 5 7 9
	
	inter := util.IntersectionOfSkipList(l1, l2, l3)
	result := []uint64{5, 9}
	for i, j := inter.Front(), 0; i != nil; i, j = i.Next(), j + 1 {
		assert.Equal(t, i.Key(), result[j])
	}
}