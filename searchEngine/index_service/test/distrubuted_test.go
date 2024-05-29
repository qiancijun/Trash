package test

import (
	"fmt"
	"net"
	"strconv"

	"github.com/qiancijun/trash/searchEngine/index_service"
	kvdb "github.com/qiancijun/trash/searchEngine/internal/kv_db"
	"github.com/qiancijun/trash/searchEngine/util"
	"google.golang.org/grpc"
)

var (
	workPorts = []int{5678, 5679, 5680}
	etcdServers = []string{"127.0.0.1:2379"}
	workers []*index_service.IndexServiceWorker
)

func StartWorkers() {
	workers = make([]*index_service.IndexServiceWorker, 0, len(workPorts))
	for i, port := range workPorts {
		lis, err := net.Listen("tcp", "127.0.0.1" + strconv.Itoa(port))
		if err != nil {
			panic(err)
		}
		server := grpc.NewServer()
		service := new(index_service.IndexServiceWorker)
		service.Init(50000, kvdb.BADGER, util.RootPath + "data/local_db/book_badge_" + strconv.Itoa(i))
		service.Indexer.LoadFromIndexFile()

		index_service.RegisterIndexServiceServer(server, service)
		service.Regist(etcdServers, port)
		go func(port int) {
			fmt.Printf("start grpc server or port %d\n", port)
			err = server.Serve(lis)
			if err != nil {
				service.Close()
				fmt.Printf("start grpc server on port %d failed: %s\n", port, err)
			} else {
				workers = append(workers, service)
			}
		}(port)
	}
}

func StopWorkers() {
	for _, worker := range workers {
        worker.Close()
    }
}
