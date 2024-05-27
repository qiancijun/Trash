package etcddemo

import (
	"context"
	"fmt"
	"strconv"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

const ConfigPrefix = "cheryl_"

type GlobalConfig struct {
	Thresh int
	Name   string
}

var Config GlobalConfig // 全局变量或者单例模式

// 从 Etcd 上读取初始的配置参数
func InitGlobalConfig(ctx context.Context, client *etcdv3.Client) {
	if response, err := client.Get(context.Background(), ConfigPrefix + "thresh"); err == nil {
		if len(response.Kvs) > 0 {
			thresh, err := strconv.Atoi(string(response.Kvs[0].Value))
			if err == nil {
				Config.Thresh = thresh
				fmt.Printf("从 Etcd 上获得的 thresh 为 %d\n", Config.Thresh)
			}
		}
	}
	if response, err := client.Get(context.Background(), ConfigPrefix + "name"); err == nil {
		if len(response.Kvs) > 0 {
			Config.Name = string(response.Kvs[0].Value)
			fmt.Printf("从 Etcd 上获得的 name 为 %s\n", Config.Name)
		}
	}
}

func Watch(ctx context.Context, client *etcdv3.Client) {
	// 监听全局配置的变化，每一个修改都会放入管道 ch
	// WithPrefix() 指明了这里的 key 实际上只是前缀
	ch := client.Watch(ctx, ConfigPrefix, etcdv3.WithPrefix()) 
	for response := range ch { // 死循环
		for _, event := range response.Events {
			if "PUT" == event.Type.String() { // 只关心 PUT 事件，即数据更新
				switch string(event.Kv.Key) {
				case ConfigPrefix + "name":
					Config.Name = string(event.Kv.Value)
					fmt.Printf("name 更新为 %s\n", Config.Name)
				case ConfigPrefix + "thresh":
					thresh, err := strconv.Atoi(string(event.Kv.Value))
					if err == nil {
						Config.Thresh = thresh
						fmt.Printf("thresh 更新为 %d\n", Config.Thresh)
					}
				}
			}
		}
	}
}

func UseConfig() {
	fmt.Printf("name=%s, thresh=%d\n", Config.Name, Config.Thresh)
}