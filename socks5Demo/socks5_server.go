package socks5demo

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

type SocksServer struct {
	port int
}

func NewSocksServer(p int) *SocksServer {
	return &SocksServer{
		port: p,
	}
}

func (server *SocksServer) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", server.port))
	if err != nil {
		panic("Listen at port " + strconv.Itoa(server.port) + "error, " + err.Error())
	}
	log.Println("socks server run on port: " + strconv.Itoa(server.port))
	id := 0 // 连接的索引号，供调试用
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept connection error: " + err.Error())
			continue
		}
		id++
		// 成功监听到一个连接，启用一个协程处理它
		// 这部分代码的逻辑参考 RFC 1928 关于 socks5 协议的文档
		go func(client net.Conn) {
			index := id
			defer client.Close()
			buf := make([]byte, 1024)
			// 连接成功后，客户端会发送一条 version identifier/method selection message
			// 这部分数据包我们只接受，不处理
			_, err := client.Read(buf)
			if err != nil {
				log.Println("read data from client error, " + err.Error())
				return
			}
			// 收到来自客户端的数据包之后，服务端需要回写一条 version ans method selection message
			// 这个数据包的字节格式参考 RFC 文档的第三页
			// 第一个字节是协议的版本，第二个字节是具体的 METHOD
			_, err = client.Write([]byte{0x5, 0x0})
			if err != nil {
				log.Println("write data to client error, " + err.Error())
				return
			}
			// 随后客户端和服务端都会进入一个特定于方法的协商阶段
			// 协商完成之后，客户端会发送详细的请求信息
			n, err := client.Read(buf)
			if err != nil {
				log.Println("read data from client error, " + err.Error())
				return
			}
			// 客户端发来的详细信息主要包含了：
			// 1. 请求的目标地址的类型：IPv4，域名，IPv6
			// 2. 目标地址
			// 3. 目标地址的端口号
			domainName := string(buf[5:n-2])
			domainPort := int(buf[n-2]) * 256 + int(buf[n-1])
			// 服务端通常会根据源地址和目标地址来评估请求
			// 并且根据请求类型返回一个或者多个消息
			// 这里就简单回复一个消息
			// 版本, 回复类型（0代表success）, 保留字段, IPv4
			_, err = client.Write([]byte{0x5, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
			// 向客户端需要访问的远程服务端建立连接
			remote, err := net.Dial("tcp", fmt.Sprintf("%s:%d", domainName, domainPort))
			if err != nil {
				log.Println("dial to remote server " + fmt.Sprintf("%s:%d", domainName, domainPort) + " error, " + err.Error())
				return
			}
			log.Printf("%d connect to remote server %s success\n", index, fmt.Sprintf("%s:%d", domainName, domainPort))
			defer remote.Close()

			done := make(chan struct{}, 1)
			// 建立完成之后，为双方传递数据
			handleFunc := func(origin, dest net.Conn) {
				buf := make([]byte, 4096)
				for {
					n1, err := origin.Read(buf)
					if err != nil {
						log.Println("read data from origin occurs some error: " + err.Error())
						break
					}
					// 读出来多少，写过去多少
					n2, err := dest.Write(buf[0:n1])
					if err != nil || n2 != n1 {
						log.Println("write data to destnation occurs some error: " + err.Error())
						break
					}
				}
				done <- struct{}{}
			}
			go handleFunc(client, remote)
			go handleFunc(remote, client)
			<- done
			log.Printf("%d has been closed\n", index)
		}(conn)
	}
}