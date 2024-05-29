package index_service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/qiancijun/trash/searchEngine/util"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	SERVICE_ROOT_PATH = "/cheryl/index" // etcd key 的前缀
)

type ServiceHub struct {
	client             *etcdv3.Client
	heartbeatFrequency int64
	watched            sync.Map
	loadBalancer       LoadBalancer
}

var (
	serviceHub *ServiceHub
	hubOnce    sync.Once
	_ IServiceHub = (*ServiceHub)(nil)
)

func GetServiceHub(etcdServers []string, heartbeanFreqency int64) *ServiceHub {
	if serviceHub == nil {
		hubOnce.Do(func() {
			if client, err := etcdv3.New(
				etcdv3.Config{
					Endpoints:   etcdServers,
					DialTimeout: 3 * time.Second,
				},
			); err != nil {
				util.Log.Fatalf("Couldn't connect to etcd: %v", err)
			} else {
				serviceHub = &ServiceHub{
					client:             client,
					heartbeatFrequency: heartbeanFreqency, // 租约有效期
					loadBalancer: &RoundRobin{},
				}
			}
		})
	}
	return serviceHub
}

// 注册服务，第一次注册向 etcd 写一个 key，后续都是在续约
func (hub *ServiceHub) Regist(service string, endpoint string, leaseID etcdv3.LeaseID) (etcdv3.LeaseID, error) {
	ctx := context.Background()
	if leaseID <= 0 {
		if lease, err := hub.client.Grant(ctx, hub.heartbeatFrequency); err != nil {
			util.Log.Printf("创建租约失败: %v", err)
			return 0, err
		} else {
			key := strings.TrimRight(SERVICE_ROOT_PATH, "/") + "/" + service + "/" + endpoint
			if _, err = hub.client.Put(ctx, key, "", etcdv3.WithLease(lease.ID)); err != nil {
				util.Log.Printf("写入服务%s对应的节点%s失败: %v", service, endpoint, err)
				return lease.ID, err
			} else {
				return lease.ID, nil
			}
		}
	} else {
		// 租约
		if _, err := hub.client.KeepAliveOnce(ctx, leaseID); err == rpctypes.ErrLeaseNotFound { //续约一次，到期后还得再续约
			return hub.Regist(service, endpoint, 0) //找不到租约，走注册流程(把leaseID置为0)
		} else if err != nil {
			util.Log.Printf("续约失败:%v", err)
			return 0, err
		} else {
			// util.Log.Printf("服务%s对应的节点%s续约成功", service, endpoint)
			return leaseID, nil
		}
	}
}

// 注销服务
func (hub *ServiceHub) UnRegist(service string, endpoint string) error {
	ctx := context.Background()
	key := strings.TrimRight(SERVICE_ROOT_PATH, "/") + "/" + service + "/" + endpoint
	if _, err := hub.client.Delete(ctx, key); err != nil {
		util.Log.Printf("注销服务%s对应的节点%s失败: %v", service, endpoint, err)
		return err
	} else {
		util.Log.Printf("注销服务%s对应的节点%s", service, endpoint)
		return nil
	}
}

func (hub *ServiceHub) GetServiceEndpoints(service string) []string {
	ctx := context.Background()
	prefix := strings.TrimRight(SERVICE_ROOT_PATH, "/") + "/" + service + "/"
	if resp, err := hub.client.Get(ctx, prefix, etcdv3.WithPrefix()); err != nil {
		util.Log.Printf("获取服务%s节点失败: %v", service, err)
		return nil
	} else {
		endpoints := make([]string, 0, len(resp.Kvs))
		for _, kv := range resp.Kvs {
			path := strings.Split(string(kv.Key), "/")
			endpoints = append(endpoints, path[len(path) - 1])
		}
		return endpoints
	}
}

// 负载均衡
func (hub *ServiceHub) GetServiceEndpoint(service string) string {
	return hub.loadBalancer.Take(hub.GetServiceEndpoints(service))
}

// 关闭 etcd client
func (hub *ServiceHub) Close() {
	hub.client.Close()
}