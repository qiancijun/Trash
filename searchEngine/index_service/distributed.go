package index_service

import (
	"context"
	fmt "fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qiancijun/trash/searchEngine/types"
	"github.com/qiancijun/trash/searchEngine/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type Sentinel struct {
	hub      IServiceHub
	connPool sync.Map
}

func NewSentinel(etcdServers []string) *Sentinel {
	return &Sentinel{
		hub:      GetServiceHubProxy(etcdServers, 10, 100),
		connPool: sync.Map{},
	}
}

func (sentinel *Sentinel) GetGrpcConn(endpoint string) *grpc.ClientConn {
	if v, exists := sentinel.connPool.Load(endpoint); exists {
		conn := v.(*grpc.ClientConn)
		if conn.GetState() == connectivity.TransientFailure || conn.GetState() == connectivity.Shutdown {
			util.Log.Printf("connection status to endpoint %s is %s", endpoint, conn.GetState())
			conn.Close()
			sentinel.connPool.Delete(endpoint)
		} else {
			return conn
		}
	}

	// 连接到服务器
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx,
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		util.Log.Printf("dial %s failed: %s", endpoint, err)
		return nil
	}
	util.Log.Printf("connect to grpc server %s", endpoint)
	sentinel.connPool.Store(endpoint, conn)
	return conn
}

// 向集群中添加文档
func (sentinel *Sentinel) AddDoc(doc types.Document) (int, error) {
	endpoint := sentinel.hub.GetServiceEndpoint(INDEX_SERVICE)
	if len(endpoint) == 0 {
		return 0, fmt.Errorf("no alive index worker")
	}
	conn := sentinel.GetGrpcConn(endpoint)
	if conn == nil {
		return 0, fmt.Errorf("connect to worker %s failed", endpoint)
	}
	client := NewIndexServiceClient(conn)
	affected, err := client.AddDoc(context.Background(), &doc)
	if err != nil {
		return 0, err
	}
	util.Log.Printf("add %d doc to worker %s", affected.Count, endpoint)
	return int(affected.Count), nil
}

// 从集群中删除文档
func (sentinel *Sentinel) DeleteDoc(docId string) int {
	endpoints := sentinel.hub.GetServiceEndpoints(INDEX_SERVICE)
	if len(endpoints) == 0 {
		return 0
	}
	var n int32
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			defer wg.Done()
			conn := sentinel.GetGrpcConn(endpoint)
			if conn == nil {
				client := NewIndexServiceClient(conn)
				affected, err := client.DeleteDoc(context.Background(), &DocId{docId})
				if err != nil {
					util.Log.Printf("delete doc %s from worker %s failed: %s", docId, endpoint, err)
				} else {
					if affected.Count > 0 {
						atomic.AddInt32(&n, affected.Count)
						util.Log.Printf("delete %d from worker %s", affected.Count, endpoint)
					}
				}
			}
		}(endpoint)
	}
	wg.Wait()
	return int(atomic.LoadInt32(&n))
}

func (sentinel *Sentinel) Search(query *types.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*types.Document {
	endpoints := sentinel.hub.GetServiceEndpoints(INDEX_SERVICE)
	if len(endpoints) == 0 {
		return nil
	}
	docs := make([]*types.Document, 0, 1000)
	resultCh := make(chan *types.Document, 1000)
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			defer wg.Done()
			conn := sentinel.GetGrpcConn(endpoint)
			if conn != nil {
				client := NewIndexServiceClient(conn)
				result, err := client.Search(context.Background(), &SearchRequest{
					Query:   query,
					OnFlag:  onFlag,
					OffFlag: offFlag,
					OrFlags: orFlags,
				})
				if err != nil {
					util.Log.Printf("search from cluster failed: %s", err)
				} else {
					if len(result.Results) > 0 {
						util.Log.Printf("search %d doc from worker %s", len(result.Results), endpoint)
						for _, doc := range result.Results {
							resultCh <- doc
						}
					}
				}
			}
		}(endpoint)
	}
	receiveFinish := make(chan struct{})
	go func() {
		for {
			doc, ok := <- resultCh
			if !ok {
				break
			}
			docs = append(docs, doc)
		}
		receiveFinish <- struct{}{}
	}()

	wg.Wait()
	close(resultCh)
	<- receiveFinish
	return docs
}

func (sentinel *Sentinel) Count() int {
	var n int32
	endpoints := sentinel.hub.GetServiceEndpoints(INDEX_SERVICE)
	if len(endpoints) == 0 {
		return 0
	}
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			defer wg.Done()
			conn := sentinel.GetGrpcConn(endpoint)
			if conn != nil {
				client := NewIndexServiceClient(conn)
				affected, err := client.Count(context.Background(), new(CountRequest))
				if err != nil {
					util.Log.Printf("get doc count from worker %s failed: %s", endpoint, err)
				} else {
					if affected.Count > 0 {
						atomic.AddInt32(&n, affected.Count)
						util.Log.Printf("worker %s have %d documents", endpoint, affected.Count)
					}
				}
			}
		}(endpoint)
	}
	return int(n)
}

// 关闭各个grpc client connection，关闭etcd client connection
func (sentinel *Sentinel) Close() (err error) {
	sentinel.connPool.Range(func(key, value any) bool {
		conn := value.(*grpc.ClientConn)
		err = conn.Close()
		return true
	})
	sentinel.hub.Close()
	return
}