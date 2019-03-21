package main

import (
	"github.com/koangel/grapeNet/KcpNet"
	"log"
)

func main() {
	cnf := kcpNet.NewConfig()
	cnf.Mode = "aes"

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
		log.Printf("Kcp Accept In:%v From:%v Recv:%v", conn.SessionId, conn.TConn.RemoteAddr().String(), string(ownerPak))
	}

	kcpNetwork.OnClose = func(conn *kcpNet.KcpConn) {
		if conn.TConn == nil {
			return
		}

		log.Printf("Kcp Closed:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr)
	}

	kcpNetwork.Runnable()
}
