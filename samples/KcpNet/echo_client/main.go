package main

import (
	"fmt"
	logger "github.com/koangel/grapeNet/Logger"
	"log"
	"math/rand"
	"time"

	kcpNet "github.com/koangel/grapeNet/KcpNet"
)

var (
	totalRecv  = int(0)
	totalCount = int(0)
	singlePack = int(0)
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func main() {
	logger.BuildLogger("./logs", "kcpNetv1Cli.log")
	cnf := kcpNet.NewConfig()
	cnf.Mode = "aes"
	cnf.Writetimeout = 35
	cnf.Readtimeout = 45
	cnf.Recvmode = kcpNet.RMReadFull

	rand.Seed(time.Now().UnixNano())
	newSendData := RandStringRunes(2048)

	kcpConn := kcpNet.NewEmptyKcp(cnf)
	if kcpConn == nil {
		log.Fatal("kcp create nil...")
	}

	kcpConn.OnConnected = func(conn *kcpNet.KcpConn) {
		/*go func() {
			for i := 0; i < 100; i++ {
				conn.Send([]byte(fmt.Sprintf("this is echo msg:%v", i)))
			}
		}()*/

		go func(c *kcpNet.KcpConn) {
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

	kcpConn.OnClose = func(conn *kcpNet.KcpConn) {
		if conn.TConn == nil {
			return
		}

		log.Printf("Kcp Closed:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr())
	}

	kcpConn.OnHandler = func(conn *kcpNet.KcpConn, ownerPak []byte) {
		totalRecv += len(ownerPak)
		totalCount++
		singlePack = len(ownerPak)
	}

	// 连接建立
	for i := 0; i < 500; i++ {
		_, err := kcpConn.Dial("127.0.0.1:4744", nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	newTimer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-newTimer.C:
			fmt.Printf("RecvBytes:%v-%v-%v\n", totalRecv, totalCount, singlePack)
			kcpConn.NetCM.Broadcast([]byte("tick..."))
		}
	}

	kcpConn.Runnable()
}
