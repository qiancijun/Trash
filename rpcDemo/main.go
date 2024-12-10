package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/qiancijun/rpcDemo/codec"
	rpcServer "github.com/qiancijun/rpcDemo/rpc_server"
)

func startServer(addr chan string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	rpcServer.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	// 一个简单的 client
	conn, _ := net.Dial("tcp", <-addr)
	defer func() { _ = conn.Close() }()

	time.Sleep(time.Second)

	// 先发送 JSON 编码的 Option
	_ = json.NewEncoder(conn).Encode(codec.DefaultOption)
	cc := codec.NewGobCodec(conn)

	// 发送 request 接受 response
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.sum",
			Seq:           uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("rpc req %d", h.Seq))
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply:", reply)
	}
}
