package etcddemo

import (
	"context"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// 尝试加锁，获取不到时会立即返回，返回的 bool 为 false 说明没有获得到锁
func tryLock(ctx context.Context, client *clientv3.Client, lockName string) (*concurrency.Session, *concurrency.Mutex, bool) {
	if session, err := concurrency.NewSession(client); err != nil {
		log.Printf("协程%d 创建会话失败:%v\n", ctx.Value("id").(int), err)
		return nil, nil, false
	} else {
		mutex := concurrency.NewMutex(session, lockName)
		if err := mutex.TryLock(ctx); err != nil { // 锁被其他协程或进程持有
			if err != concurrency.ErrLocked { // 如果没获得锁，不打印 error 信息
				log.Printf("协程%d TryLock 异常:%v\n", ctx.Value("id").(int), err)
			}
			return session, mutex, false
		} else {
			return session, mutex, true
		}
	}
}

func testTryLock(ctx context.Context, client *clientv3.Client, lockName string, routineID int) {
	// 协程 id 不应该暴露给应用层，建议通过 context 关联上下文
	ctx = context.WithValue(ctx, "id", routineID)
	session, mutex, success := tryLock(ctx, client, lockName)
	if success {
		log.Printf("协程 %d 获得锁\n", routineID)
		time.Sleep(time.Second) // 执行业务需求
		mutex.Unlock(ctx) // 释放锁
		log.Printf("协程 %d 释放锁\n", routineID)
		session.Close()
	} else {
		log.Printf("协程 %d 锁被其他会话持有\n", routineID)
	}
}

func lockWithTimeout(ctx context.Context, client *clientv3.Client, lockName string) (*concurrency.Session, *concurrency.Mutex) {
	toctx, cancel := context.WithTimeout(ctx, 100 * time.Millisecond) // 设定超时 100ms
	defer cancel()
	if session, err := concurrency.NewSession(client); err != nil {
		log.Printf("协程 %d 创建会话失败:%v\n", ctx.Value("id").(int), err)
		return nil, nil
	} else {
		mutex := concurrency.NewMutex(session, lockName)
		if err := mutex.Lock(toctx); err != nil {
			// toctx 是一个带超时的 context，如果一直获得不到锁，超时后就返回 error
			if err != context.DeadlineExceeded {
				// 如果时超时，不打印 error 信息
				log.Printf("协程 %d Lock 异常:%v\n", ctx.Value("id").(int), err)
			} else {
				log.Printf("协程 %d 指定时间内未获得到锁\n", ctx.Value("id").(int))
			}
			return session, nil
		} else {
			return session, mutex
		}
	}
}