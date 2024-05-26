package util

import "github.com/huandu/skiplist"

func IntersectionOfSkipList(lists ...*skiplist.SkipList) *skiplist.SkipList {
	if len(lists) == 0 {
		return nil
	}
	if len(lists) == 1 {
		return lists[0]
	}
	result := skiplist.New(skiplist.Uint64)
	currNodes := make([]*skiplist.Element, len(lists))
	for i, list := range lists {
		if list == nil || list.Len() == 0 {
			return nil
		}
		currNodes[i] = list.Front()
	}
	for {
		maxList := make(map[int]struct{}, len(currNodes))
		var maxValue uint64 = 0
		for i, node := range currNodes {
			if node.Key().(uint64) > maxValue {
				maxValue = node.Key().(uint64)
				maxList = map[int]struct{}{i: {}}
			} else if node.Key().(uint64) == maxValue {
				maxList[i] = struct{}{}
			}
		}
		if len(maxList) == len(currNodes) {
			result.Set(currNodes[0].Key(), currNodes[0].Value)
			for i, node := range currNodes {
				currNodes[i] = node.Next()
				if currNodes[i] == nil {
					return result
				}
			}
		} else {
			for i, node := range currNodes {
				if _, exists := maxList[i]; !exists { // 值大的不动，小的往后移
					currNodes[i] = node.Next() // 不能用node=node.Next()，因为for range取得的是值拷贝
					if currNodes[i] == nil {   // 只要有一条SkipList已走到最后，则说明不会再有新的交集诞生，可以return了
						return result
					}
				}
			}
		}
	}
}

// 求多个SkipList的并集
func UnionsetOfSkipList(lists ...*skiplist.SkipList) *skiplist.SkipList {
	if len(lists) == 0 {
		return nil
	}
	if len(lists) == 1 {
		return lists[0]
	}
	result := skiplist.New(skiplist.Uint64)
	keySet := make(map[any]struct{}, 1000)
	for _, list := range lists {
		if list == nil {
			continue
		}
		node := list.Front()
		for node != nil {
			if _, exists := keySet[node.Key()]; !exists {
				result.Set(node.Key(), node.Value)
				keySet[node.Key()] = struct{}{}
			}
			node = node.Next()
		}
	}
	return result
}