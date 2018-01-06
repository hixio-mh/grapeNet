package main

////  监听测试服务器，用于测试网络库

import (
	"log"

	tcp "github.com/koangel/grapeNet/Net"

	"net/http"
	_ "net/http/pprof"
)

func RecvEchoMsg(conn *tcp.TcpConn, Pak []byte) {
	//fmt.Println(string(Pak))

	conn.Send(Pak) // 回执
}

func OnClosed(conn *tcp.TcpConn) {
	log.Println("连接断开了:", conn.GetSessionId())
}

func main() {

	go func() {
		http.ListenAndServe(":6687", nil)
	}()

	tcpNet, err := tcp.NewTcpServer(":8799")
	if err != nil {
		log.Fatal(err)
		return
	}

	tcpNet.OnClose = OnClosed
	tcpNet.OnHandler = RecvEchoMsg // 绑定callback

	tcpNet.Runnable() // 占满并跑起来
}
