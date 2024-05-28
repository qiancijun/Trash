package index_service

import (
	"math/rand"
	"sync/atomic"
)

type LoadBalancer interface {
	Take([]string) string
}

type RoundRobin struct {
	acc int64
}

var _ LoadBalancer = (*RoundRobin)(nil)

func (b *RoundRobin) Take(endpoints []string) string {
	if len(endpoints) == 0 {
		return ""
	}
	n := atomic.AddInt64(&b.acc, 1)
	index := int(n % int64(len(endpoints)))
	return endpoints[index]
}

type RandomSelect struct {

}

var _ LoadBalancer = (*RandomSelect)(nil)

func (b *RandomSelect) Take(endpoints []string) string {
	if len(endpoints) == 0 {
		return ""
	}
	index := rand.Intn(len(endpoints))
	return endpoints[index]
}