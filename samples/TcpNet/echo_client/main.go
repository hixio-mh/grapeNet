package main

/// 测试大量连接的情况下TCP处理性能和队列性能

import (
	"fmt"
	"log"
	"time"

	tcp "github.com/koangel/grapeNet/Net"
)

var (
	totalRecv = int(0)
)

func RecvEchoMsg(conn *tcp.TcpConn, Pak []byte) {
	//fmt.Println(conn.GetSessionId(), string(Pak))
	totalRecv += len(Pak)
}

func OnClose(conn *tcp.TcpConn) {

}

func main() {
	log.Printf("start echo clients...")
	connNet := tcp.NewEmptyTcp() // 空的TCP

	connNet.OnHandler = RecvEchoMsg
	connNet.OnClose = OnClose
	// 连接建立
	for i := 0; i < 5000; i++ {
		_, err := connNet.Dial("127.0.0.1:8799", nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	for i := 0; i < 100; i++ {
		connNet.NetCM.Broadcast([]byte(fmt.Sprintf("this is echo msg:%v", i)))
	}

	newTimer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-newTimer.C:
			fmt.Printf("RecvBytes:%v\n", totalRecv)
			//connNet.NetCM.Broadcast([]byte("tick..."))
		}
	}

	connNet.Runnable()
}
