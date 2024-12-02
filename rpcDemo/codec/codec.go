package codec

import "io"

type Type string

// 客户端和服务端可以通过 Codec 的 Type 得到构造函数，从而创建 Codec 实例。
const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

/**
* 一个 RPC 的请求应该为 client.Call("Arith.Multiply", args, &reply)
* 客户端发送的请求包括服务名 Arith，方法名 Multiply，参数 args
* 服务端的相应包括错误 error，返回值 reply
* 将请求和响应中的参数和返回值抽象为 body，其余信息放在 header 中
 */
type Header struct {
	ServiceMethod string // 服务名和方法名
	Seq           uint64 // 请求的序号
	Error         string // 错误信息，客户端置为空，服务端如果如果发生错误，将错误信息置于 Error 中。
}

// 抽象出对消息体进行编解码的接口 Codec，抽象出接口是为了实现不同的 Codec 实例
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}