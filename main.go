package main

import (
	socks "github.com/qiancijun/socks5Demo"
)

func main() {
	runSocks5Demo()
}

func runSocks5Demo() {
	server := socks.NewSocksServer(8888)
	server.Start()
}