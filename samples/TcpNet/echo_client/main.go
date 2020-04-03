package main

/// 测试大量连接的情况下TCP处理性能和队列性能

import (
	"fmt"
	logger "github.com/koangel/grapeNet/Logger"
	"log"
	"math/rand"
	"net/http"
	"time"

	tcp "github.com/koangel/grapeNet/Net"
	_ "net/http/pprof"
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
func main() {
	go func() {
		http.ListenAndServe(":6688", nil)
	}()
	logger.BuildLogger("./logs", "tcpNetv1Cli.log")
	log.Printf("start echo clients...")
	connNet := tcp.NewEmptyTcp(tcp.RMReadFull) // 空的TCP

	rand.Seed(time.Now().UnixNano())
	newSendData := RandStringRunes(2048)

	connNet.OnHandler = RecvEchoMsg
	connNet.OnClose = OnClose
	// 连接建立
	for i := 0; i < 3500; i++ {
		conn, err := connNet.Dial("localhost:8799", nil)
		if err != nil {
			log.Fatal(err)
		}

		go func(c *tcp.TcpConn) {
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

	for i := 0; i < 1000; i++ {
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
