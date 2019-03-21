package main

import (
	"fmt"
	"github.com/koangel/grapeNet/KcpNet"
	"log"
	"time"
)

var (
	totalRecv = int(0)
)

func main() {
	cnf := kcpNet.NewConfig()
	cnf.Mode = "aes"

	kcpConn := kcpNet.NewEmptyKcp(cnf)
	if kcpConn == nil {
		log.Fatal("kcp create nil...")
	}

	kcpConn.OnClose = func(conn *kcpNet.KcpConn) {
		if conn.TConn == nil {
			return
		}

		log.Printf("Kcp Closed:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr)
	}

	kcpConn.OnHandler = func(conn *kcpNet.KcpConn, ownerPak []byte) {
		totalRecv += len(ownerPak)
	}

	// 连接建立
	for i := 0; i < 5000; i++ {
		_, err := kcpConn.Dial("127.0.0.1:4744", nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	for i := 0; i < 100; i++ {
		kcpConn.NetCM.Broadcast([]byte(fmt.Sprintf("this is echo msg:%v", i)))
	}

	newTimer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-newTimer.C:
			fmt.Printf("RecvBytes:%v\n", totalRecv)
			//connNet.NetCM.Broadcast([]byte("tick..."))
		}
	}

	kcpConn.Runnable()
}
