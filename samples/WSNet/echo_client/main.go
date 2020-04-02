package main

import (
	"fmt"
	"log"
	"time"

	ws "github.com/koangel/grapeNet/Websocket"
)

/// 测试WS
var (
	totalRecv  = int(0)
	totalCount = int(0)
	singlePack = int(0)
)

func RecvEchoMsg(conn *ws.WSConn, Pak []byte) {
	//fmt.Println(conn.GetSessionId(), string(Pak))
	totalRecv += len(Pak)
	totalCount++
	singlePack = len(Pak)
}

func OnClose(conn *ws.WSConn) {
	log.Println("连接断开了:", conn.GetSessionId())
}

func main() {
	wsNet := ws.NetEmptyWS("test.server.me", "/ws")

	wsNet.OnClose = OnClose
	wsNet.OnHandler = RecvEchoMsg
	wsNet.NetCM.SendMode = 1 // 改为直接发模式测试

	// 连接建立
	for i := 0; i < 100; i++ {
		_, err := wsNet.Dial("localhost:47892")
		if err != nil {
			log.Fatal(err)
		}

	}

	for i := 0; i < 2000; i++ {
		go wsNet.NetCM.Broadcast([]byte(fmt.Sprintf("this is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msgthis is echo msg:%v", i)))
	}

	newTimer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-newTimer.C:
			fmt.Printf("RecvBytes:%v-%v-%v\n", totalRecv, totalCount, singlePack)
		}
	}

	// 连接不必Runable
	for {
		time.Sleep(1 * time.Second)
	}
}
