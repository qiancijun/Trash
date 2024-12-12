package service

import (
	"reflect"
	"sync/atomic"
)

/**
* 假设客户端发送一个请求，包含 ServiceName 和 Argv
* {
*     "ServiceMethod": "T.MethodName"
*     "Argv": ""
* }
* 通过 T.MethodName 可以确定调用的是类型 T 的 MethodName
* 通过反射能够非常容易获取某个结构体的所有方法，并且能够通过方法，
* 获取到该方法的所有的参数类型与返回值
 */

/**
* 通过反射实现结构体与服务的映射关系
* 每一个 MethodType 实例包含了一个方法的完整信息
*   - method: 方法本身
*   - ArgType: 第一个参数的类型
*   - ReplyType: 第二个参数的类型
*   - numCalls: 后续统计方法调用次数用
 */
type MethodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64
}

func (m *MethodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

// 用户创建对应类型的实例，指针类型和值类型实例创建有细微区别
func (m *MethodType) NewArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *MethodType) NewReply() reflect.Value {
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}
