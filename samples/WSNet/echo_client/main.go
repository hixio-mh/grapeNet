package main

import (
	"fmt"
	logger "github.com/koangel/grapeNet/Logger"
	"log"
	"math/rand"
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func main() {
	logger.BuildLogger("./logs", "wsnetcli.log")
	wsNet := ws.NetEmptyWS("test.server.me", "/ws")

	wsNet.OnClose = OnClose
	wsNet.OnHandler = RecvEchoMsg
	//wsNet.NetCM.SendMode = 1 // 改为直接发模式测试
	rand.Seed(time.Now().UnixNano())
	newSendData := RandStringRunes(2048)

	// 连接建立
	for i := 0; i < 1000; i++ {
		conn, err := wsNet.Dial("localhost:47892")
		if err != nil {
			log.Fatal(err)
		}

		go func(c *ws.WSConn) {
			log.Println("start tick send...")
			defer log.Println("stop tick send...")
			for {
				if c.IsClosed == 1 {
					break
				}
				c.SendDirect([]byte(newSendData))
				time.Sleep(time.Second)
			}
		}(conn)
	}

	for i := 0; i < 2000; i++ {
		wsNet.NetCM.Broadcast([]byte(newSendData))
	}

	newTimer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-newTimer.C:
			fmt.Printf("RecvBytes:%v-%v-%v\n", totalRecv, totalCount, singlePack)

			for i := 0; i < 2000; i++ {
				wsNet.NetCM.Broadcast([]byte(newSendData))
			}
		}
	}

	// 连接不必Runable
	for {
		time.Sleep(1 * time.Second)
	}
}
