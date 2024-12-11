package rpcClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/qiancijun/rpcDemo/codec"
)

/**
* cc: 消息编解码器，和服务端类似，用来序列化要发送出去的请求，以及反序列化接受到的相应
* sending: 互斥锁，和服务端类似，为了保证请求的有序发送，即防止多个请求报文混淆
* header: 每个请求的消息头，header 只有在请求发送时候才需要，而请求发送是互斥的，因此每个客户端只需要一个
* seq: 用户给发送的请求编号，每个请求拥有唯一编号
* pending: 用于存储未处理完的请求，键是编号，值是 Call 实例
* closing 和 shutdown 任意一个值为 true，则表示 client 处于不可用的状态。closing 一般是主动关闭的，shutdown 一般有错误发生
 */
type Client struct {
	cc       codec.Codec
	opt      *codec.Option
	sending  sync.Mutex
	header   codec.Header
	mu       sync.Mutex
	seq      uint64
	pending  map[uint64]*codec.Call
	closing  bool // user has called Close
	shutdown bool // server has told be stoped
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

/**
* 创建 Client 实例首先需要完成一开始的协议交换，即发送 Option
* 信息给服务端。协商好消息的编解码方式之后，再创建协程接受响应
 */
func NewClient(conn net.Conn, opt *codec.Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error:", err)
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(cc codec.Codec, opt *codec.Option) *Client {
	client := &Client{
		seq:     1,
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*codec.Call),
	}
	go client.receive()
	return client
}

// 便于用户传入网络地址来创建客户端
func Dial(network, address string, opts ...*codec.Option) (client *Client, err error) {
	opt, err := codec.ParseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}

func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}

// 将参数 call 添加到 client.pending 中，并更新 client.seq
func (client *Client) registerCall(call *codec.Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing || client.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

// 根据 seq 从 client.pending 中删除对应的 call 并返回
func (client *Client) removeCall(seq uint64) *codec.Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

// 服务端或者客户端发生错误时调用，shutdown 设置为 true
func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.Done()
	}
}

/**
* 接受请求，接收到的响应有三种情况：
*   1. call 不存在，可能是请求没有发送完整，或者因为其他情况被取消，但是服务端依旧处理了
*   2. call 存在，但是服务端处理出错，即 h.Error 不为空
*   3. call 存在，服务端处理正常，那么需要从 body 中读取 reply 的值
 */
func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			err = client.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			call.Done()
		default:
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.Done()
		}
	}
	client.terminateCalls(err)
}

/**
* Go 和 Call 是客户端暴露给用户的两个 RPC 服务调用接口
*   - Go 是一个异步接口，返回 Call 实例
*   - Call 是对 Go 的封装，阻塞 call.DoneChan 等待响应返回，是一个同步接口
 */
func (client *Client) send(call *codec.Call) {
	client.sending.Lock()
	defer client.sending.Unlock()

	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.Done()
		return
	}

	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		if call != nil {
			call.Error = err
			call.Done()
		}
	}
}

func (client *Client) Go(serviceMethod string, args, reply interface{}, done chan *codec.Call) *codec.Call {
	if done == nil {
		done = make(chan *codec.Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}

	call := &codec.Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		DoneChan:      done,
	}
	client.send(call)
	return call
}

func (client *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-client.Go(serviceMethod, args, reply, make(chan *codec.Call, 1)).DoneChan
	return call.Error
}
