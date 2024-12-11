package codec

/**
* 对 net/rpc 来说，一个函数需要能够被远程调用，需要满足以下五个条件：
*   1. the method's type is exported
*   2. the method is exported
*   3. the method has two arguments, both exported (or builtin) types
*   4. the method's second argument is a pointer
*   5. the method has return type error
*
* 比如 func (t *T) MethodName(argType T1, replyType *T2) error
 */

/**
* 封装结构体 Call 来承载一次 RPC 调用所需要的信息
 */

type Call struct {
	Seq           uint64
	ServiceMethod string      // format "<service>.<method>"
	Args          interface{} // arguments to the function
	Reply         interface{} // reply from the function
	Error         error       // if error occurs, it will be set
	DoneChan      chan *Call
}

// 为了支持异步调用，当调用结束会调用 done 通知调用方
func (call *Call) Done() {
	call.DoneChan <- call
}
