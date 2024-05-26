package test

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/qiancijun/trash/searchEngine/util"
)

var conMp = util.NewConcurrentHashMap(8, 1000)
var synMp = sync.Map{}

func readConMap() {
	for i := 0; i < 10000; i++ {
		key := strconv.Itoa(int(rand.Int63()))
		conMp.Get(key)
	}
}

func readSynMap() {
	for i := 0; i < 10000; i++ {
		key := strconv.Itoa(int(rand.Int63()))
		synMp.Load(key)
	}
}

func writeConMap() {
	for i := 0; i < 10000; i++ {
		key := strconv.Itoa(int(rand.Int63()))
		conMp.Set(key, 1)
	}
}

func writeSynMap() {
	for i := 0; i < 10000; i++ {
		key := strconv.Itoa(int(rand.Int63()))
		synMp.Store(key, 1)
	}
}

func BenchmarkConMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		const P = 300
		wg := sync.WaitGroup{}
		wg.Add(2 * P)
		for i := 0; i < P; i++ {
			go func() {
				defer wg.Done()
				readConMap()
			}()
		}
		for i := 0; i < P; i++ {
			go func() {
				defer wg.Done()
				writeConMap()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkSynMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		const P = 300
		wg := sync.WaitGroup{}
		wg.Add(2 * P)
		for i := 0; i < P; i++ { //300个协程一直读
			go func() {
				defer wg.Done()
				readSynMap()
			}()
		}
		for i := 0; i < P; i++ { //300个协程一直写
			go func() {
				defer wg.Done()
				writeSynMap()
				// time.Sleep(100 * time.Millisecond)
			}()
		}
		wg.Wait()
	}
}

func TestConcurrentHashMapIterator(t *testing.T) {
	for i := 0; i < 10; i++ {
		conMp.Set(strconv.Itoa(i), i)
	}
	iterator := conMp.CreateIterator()
	entry := iterator.Next()
	for entry != nil {
		fmt.Println(entry.Key, entry.Value)
		entry = iterator.Next()
	}
}