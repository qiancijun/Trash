package index_service

import (
	"context"
	"fmt"
	"time"

	"github.com/qiancijun/trash/searchEngine/types"
	"github.com/qiancijun/trash/searchEngine/util"
)

const (
	INDEX_SERVICE = "index_service"
)

type IndexServiceWorker struct {
	Indexer  *Indexer
	hub      *ServiceHub
	selfAddr string
}

var _ IndexServiceServer = (*IndexServiceWorker)(nil)

func (service *IndexServiceWorker) Init(docNumEstimate int, dbtype int, dataDir string) error {
	service.Indexer = new(Indexer)
	return service.Indexer.Init(docNumEstimate, dbtype, dataDir)
}

func (service *IndexServiceWorker) Regist(etcdServers []string, servicePort int) error {
	if len(etcdServers) > 0 {
		if servicePort < 1024 {
			return fmt.Errorf("invalid listen port %d", servicePort)
		}
		selfLocalIp, err := util.GetLocalIP()
		if err != nil {
			panic(err)
		}
		selfLocalIp = "127.0.0.1" // 单机模拟分布式，把 IP 地址写死
		service.selfAddr = fmt.Sprintf("%s:%d", selfLocalIp, servicePort)
		var heartBeat int64 = 3
		hub := GetServiceHub(etcdServers, heartBeat)
		leaseId, err := hub.Regist(INDEX_SERVICE, service.selfAddr, 0)
		if err != nil {
			panic(err)
		}
		service.hub = hub
		go func() {
			for {
				hub.Regist(INDEX_SERVICE, service.selfAddr, leaseId)
				time.Sleep(time.Duration(heartBeat) * time.Second - 100 * time.Millisecond)
			}
		}()
	}
	return nil
}

func (service *IndexServiceWorker) Close() error {
	if service.hub != nil {
		service.hub.UnRegist(INDEX_SERVICE, service.selfAddr)
	}
	return service.Indexer.Close()
}

func (service *IndexServiceWorker) AddDoc(ctx context.Context, doc *types.Document) (*AffectedCount, error) {
	n, err := service.Indexer.AddDoc(*doc)
	return &AffectedCount{int32(n)}, err
}

func (service *IndexServiceWorker) DeleteDoc(ctx context.Context, docId *DocId) (*AffectedCount, error) {
	return &AffectedCount{int32(service.Indexer.DeleteDoc(docId.DocId))}, nil
}

func (service *IndexServiceWorker) Search(ctx context.Context, request *SearchRequest) (*SearchResult, error) {
	result := service.Indexer.Search(request.Query, request.OnFlag, request.OffFlag, request.OrFlags)
	return &SearchResult{result}, nil
}

func (service *IndexServiceWorker) Count(ctx context.Context, request *CountRequest) (*AffectedCount, error) {
	return &AffectedCount{int32(service.Indexer.Count())}, nil
}