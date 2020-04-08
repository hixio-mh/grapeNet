package main

////  监听测试服务器，用于测试网络库

import (
	"fmt"
	logger "github.com/koangel/grapeNet/Logger"
	tcp "github.com/koangel/grapeNet/Net"
	"log"
	"time"

	"net/http"
	_ "net/http/pprof"
)

var (
	totalRecv  = int(0)
	totalCount = int(0)
	singlePack = int(0)
)

func RecvEchoMsg(conn *tcp.TcpConn, Pak []byte) {
	//fmt.Println(string(Pak))

	conn.SendDirect(Pak) // 回执

	//if rand.Intn(100000) <= 3000 {
	//	conn.CloseSocket()
	//}

	totalRecv += len(Pak)
	totalCount++
	singlePack = len(Pak)
}

func OnClosed(conn *tcp.TcpConn) {
	log.Println("连接断开了:", conn.GetSessionId())
}

func main() {

	go func() {
		http.ListenAndServe(":6687", nil)
	}()
	logger.BuildLogger("./logs", "tcpNetv1.log")
	tcpNet, err := tcp.NewTcpServer(tcp.RMReadFull, ":8799")
	if err != nil {
		log.Fatal(err)
		return
	}

	tcpNet.OnClose = OnClosed
	tcpNet.OnHandler = RecvEchoMsg // 绑定callback
	tcpNet.OnAccept = func(conn *tcp.TcpConn) {}
	tcpNet.OnClose = func(conn *tcp.TcpConn) {}
	tcpNet.NetCM.SendMode = 1 // 直接发送模式

	go tcpNet.Runnable() // 占满并跑起来
	newTimer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-newTimer.C:
			fmt.Printf("Server RecvBytes:%v-%v-%v\n", totalRecv, totalCount, singlePack)
		}
	}

	// 连接不必Runable
	for {
		time.Sleep(1 * time.Second)
	}

}
