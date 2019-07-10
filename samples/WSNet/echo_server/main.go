package main

/// 测试WS连接行为

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	ws "github.com/koangel/grapeNet/Websocket"
)

func RecvEchoMsg(conn *ws.WSConn, Pak []byte) {
	fmt.Println(string(Pak))

	conn.Send(Pak) // 回执
}

func OnClosed(conn *ws.WSConn) {
	log.Println("连接断开了:", conn.GetSessionId())
}

func main() {
	go func() {
		http.ListenAndServe(":6687", nil)
	}()

	wsNet := ws.NewWebsocket(":47892", "", "/ws")

	wsNet.SetTextMessage()
	wsNet.OnHandler = RecvEchoMsg
	wsNet.OnClose = OnClosed

	wsNet.Runnable()
}
