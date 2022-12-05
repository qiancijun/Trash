package compress

import (
	"container/heap"
)

type Node struct {
	val    byte
	freq   int
	isLeaf bool
	child  []*Node // [0]: 左孩子 [1]: 右孩子
}

func NewNode(val byte, fre int, isLeaf bool) *Node {
	return &Node{
		val:    val,
		freq:   fre,
		isLeaf: isLeaf,
		child:  make([]*Node, 2),
	}
}

/* 根据频率表生成对应的哈夫曼树
*  贪心思想：将频率从大到小排序，每次选出两个出现最多的字节进行合并
*/
func NewHuffmanTree(fre []int) *Node {
	pq := hp{}
	for i := 1; i < len(fre); i++ {
		if fre[i] != 0 {
			heap.Push(&pq, NewNode(byte(i), fre[i], true))
		}
	}
	// 最后一个节点作为 root
	for len(pq) > 1 {
		left, right := heap.Pop(&pq).(*Node), heap.Pop(&pq).(*Node)
		// 合并成一个新的节点
		newNode := NewNode(0, left.freq + right.freq, false)
		newNode.child[0], newNode.child[1] = left, right
		heap.Push(&pq, newNode)
	}
	return heap.Pop(&pq).(*Node)
}