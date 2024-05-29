package index_service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/qiancijun/trash/searchEngine/util"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/time/rate"
)

type HubProxy struct {
	*ServiceHub
	endpointCache sync.Map
	limiter *rate.Limiter
}

var (
	proxy *HubProxy
	proxyOnce sync.Once
)

func GetServiceHubProxy(etcdServices []string, heartbeatFrequency int64, qps int) *HubProxy {
	if proxy == nil {
		proxyOnce.Do(func () {
			serviceHub := GetServiceHub(etcdServices, heartbeatFrequency)
			proxy = &HubProxy{
				ServiceHub: serviceHub,
				endpointCache: sync.Map{},
				limiter: rate.NewLimiter(rate.Every(time.Duration(1e9 / qps)*time.Nanosecond), qps),
			}
		})
	}
	return proxy
}

func (proxy *HubProxy) watchEndpointsOfService(service string) {
	if _, exists := proxy.watched.LoadOrStore(service, true); exists {
		return
	}
	ctx := context.Background()
	prefix := strings.TrimRight(SERVICE_ROOT_PATH, "/") + "/" + service + "/"
	ch := proxy.client.Watch(ctx, prefix, etcdv3.WithPrefix())
	go func() {
		for response := range ch {
			for _, event := range response.Events {
				util.Log.Printf("etcd event type %s", event.Type)
				path := strings.Split(string(event.Kv.Key), "/")
				if len(path) > 2 {
					service := path[len(path)-2]
					endpoints := proxy.ServiceHub.GetServiceEndpoints(service)
					if len(endpoints) > 0 {
						proxy.endpointCache.Store(service, endpoints)
					} else {
						proxy.endpointCache.Delete(service) // 该 service 下已经没有 endpoint
					}
				}
			}
		}
	}()
}

// 服务发现
func (proxy *HubProxy) GetServiceEndpoints(service string) []string {
	if !proxy.limiter.Allow() {
		return nil
	}
	proxy.watchEndpointsOfService(service)
	if endpoints, exists := proxy.endpointCache.Load(service); exists {
		return endpoints.([]string)
	} else {
		endpoints := proxy.ServiceHub.GetServiceEndpoints(service)
		if len(endpoints) > 0 {
			proxy.endpointCache.Store(service, endpoints)
		}
		return endpoints
	}
}