package etcddemo_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

func getConn() (*etcdv3.Client, error) {
	return etcdv3.New(
		etcdv3.Config{
			Endpoints: []string{"127.0.0.1:2379"},
			DialTimeout: 3 * time.Second,
		},
	)
}

func TestConn(t *testing.T) {
	client, err := etcdv3.New(
		etcdv3.Config{
			Endpoints: []string{"127.0.0.1:2379"},
			DialTimeout: 3 * time.Second,
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestReadAndWrite(t *testing.T) {
	conn, err := getConn()
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	ctx := context.Background()
	_, err = conn.Put(ctx, "name", "cheryl")
	assert.NoError(t, err)
	response, err := conn.Get(ctx, "name")
	assert.NoError(t, err)
	for _, kv := range response.Kvs {
		fmt.Printf("%s=%s\n", kv.Key, kv.Value)
	}
}

func TestReadRange(t *testing.T) {
	client, err := getConn()
	assert.NoError(t, err)
	assert.NotNil(t, client)
	for i := 1; i <= 3; i++ {
		_, err = client.Put(context.Background(), strconv.Itoa(i), strconv.Itoa(i))
		assert.NoError(t, err)
	}
	res, err := client.Get(context.Background(), "1", etcdv3.WithRange("3"))
	assert.NoError(t, err)
	for _, kv := range res.Kvs {
		fmt.Printf("%s=%s\n", kv.Key, kv.Value)
	}
}

func TestLease(t *testing.T) {
	client, err := getConn()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// 第二个参数 ttl 的单位为秒
	lease, err := client.Grant(context.Background(), 2)
	assert.NoError(t, err)

	_, err = client.Put(context.Background(), "name", "cheryl", etcdv3.WithLease(lease.ID))
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
	res, err := client.Get(context.Background(), "name")
	assert.NoError(t, err)
	for _, kv := range res.Kvs {
		fmt.Printf("%s=%s\n", kv.Key, kv.Value)
	}

	// 必须在到期之前续，否则找不到对应的租约
	// 通过 KeepAlive 使租约永久生效
	_, err = client.KeepAlive(context.Background(), lease.ID)
	assert.NoError(t, err)
	time.Sleep(2 * time.Second)
	res, err = client.Get(context.Background(), "name")
	assert.NoError(t, err)
	for _, kv := range res.Kvs {
		fmt.Printf("%s=%s\n", kv.Key, kv.Value)
	}
}

func TestKeppAliveOnce(t *testing.T) {
	client, err := getConn()
	assert.NoError(t, err)
	assert.NotNil(t, client)
	// 第二个参数 ttl 的单位为秒
	lease, err := client.Grant(context.Background(), 2)
	assert.NoError(t, err)

	_, err = client.Put(context.Background(), "name", "cheryl", etcdv3.WithLease(lease.ID))
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
	res, err := client.Get(context.Background(), "name")
	assert.NoError(t, err)
	for _, kv := range res.Kvs {
		fmt.Printf("%s=%s\n", kv.Key, kv.Value)
	}

	// 本来租约有效期只有 2 秒，通过 KeepAliveOnce 再续 2 秒
	// 之前剩余的时间会清零
	_, err = client.KeepAliveOnce(context.Background(), lease.ID)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
	res, err = client.Get(context.Background(), "name")
	assert.NoError(t, err)
	for _, kv := range res.Kvs {
		fmt.Printf("%s=%s\n", kv.Key, kv.Value)
	}
}
