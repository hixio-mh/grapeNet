package main

/// 测试WS连接行为

import (
	"fmt"
	logger "github.com/koangel/grapeNet/Logger"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	ws "github.com/koangel/grapeNet/Websocketv2"
)

var (
	totalRecv  = int(0)
	totalCount = int(0)
	singlePack = int(0)

	dataMap   sync.Map
	lastPrint = time.Now().Unix()
)

func RecvEchoMsg(conn *ws.WSConn, Pak []byte) {
	//fmt.Println(string(Pak))
	if time.Now().Unix() > lastPrint {
		fmt.Println(string(Pak))
		lastPrint = time.Now().Add(2 * time.Minute).Unix()
	}
	conn.Send(Pak) // 回执

	totalRecv += len(Pak)
	totalCount++
	singlePack = len(Pak)

	/*packCount, has := dataMap.LoadOrStore(conn.SessionId, int(1))
	if has {
		v, ok := packCount.(int)
		if ok {
			dataMap.Store(conn.SessionId, v+1)
		}
	}*/

}

func OnClosed(conn *ws.WSConn) {
	dataMap.Delete(conn.SessionId)
	log.Println("连接断开了:", conn.GetSessionId())
}

func main() {
	go func() {
		http.ListenAndServe(":6687", nil)
	}()

	logger.BuildLogger("./logs", "wsnetv2.log")
	wsNet := ws.NewWebsocket(":47892", "", "/ws")

	wsNet.OnHandler = RecvEchoMsg
	wsNet.OnClose = OnClosed
	//wsNet.SetTextMessage()

	go wsNet.Runnable()

	newTimer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-newTimer.C:
			fmt.Printf("Server RecvBytes:%v-%v-%v\n", totalRecv, totalCount, singlePack)

			//dataMap.Range(func(key, value interface{}) bool {
			//	fmt.Println("session:", key, ",", value)
			//	return true
			//})
		}
	}

	// 连接不必Runable
	for {
		time.Sleep(1 * time.Second)
	}
}
