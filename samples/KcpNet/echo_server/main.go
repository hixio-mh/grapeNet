package main

import (
	"fmt"
	"github.com/koangel/grapeNet/KcpNet"
	logger "github.com/koangel/grapeNet/Logger"
	"log"
	"time"
)

var (
	totalRecv  = int(0)
	totalCount = int(0)
	singlePack = int(0)
)

func main() {
	logger.BuildLogger("./logs", "kcpNetv1.log")
	cnf := kcpNet.NewConfig()
	cnf.Mode = "aes"
	cnf.Writetimeout = 35
	cnf.Readtimeout = 45
	cnf.Recvmode = kcpNet.RMReadFull

	// 需要自主加解密算法的使用
	//cnf.Mode = "none"

	kcpNetwork, err := kcpNet.NewKcpServer(":4744", cnf)
	if err != nil {
		log.Fatal(err)
	}

	kcpNetwork.OnAccept = func(conn *kcpNet.KcpConn) {
		log.Printf("Kcp Accept In:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr().String())
	}

	kcpNetwork.OnHandler = func(conn *kcpNet.KcpConn, ownerPak []byte) {
		//log.Printf("Kcp Accept In:%v From:%v Recv:%v", conn.SessionId, conn.TConn.RemoteAddr().String(), string(ownerPak))
		conn.Send(ownerPak)

		totalRecv += len(ownerPak)
		totalCount++
		singlePack = len(ownerPak)
	}

	kcpNetwork.OnClose = func(conn *kcpNet.KcpConn) {
		if conn.TConn == nil {
			return
		}

		log.Printf("Kcp Closed:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr)
	}

	go kcpNetwork.Runnable()
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
