package grapeNet

import (
	"fmt"
	"testing"
	"time"

	stream "github.com/koangel/grapeNet/Stream"
)

type MCNetwork struct{}

type userData struct {
	userName string
}

// 创建用户DATA
func (c *MCNetwork) CreateUserData() interface{} {
	fmt.Println("create User Data !!!")
	return &userData{}
}

// 通知连接
func (c *MCNetwork) OnAccept(conn *TcpConn) {
	fmt.Println("OnAccept", conn.SessionId)
}
func (c *MCNetwork) OnHandler(conn *TcpConn, ownerPak *stream.BufferIO) {
	fmt.Println("OnHandler", conn.SessionId)
}
func (c *MCNetwork) OnClose(conn *TcpConn) {
	fmt.Println("OnClose", conn.SessionId)
}
func (c *MCNetwork) OnConnected(conn *TcpConn) {
	fmt.Println("OnConnected", conn.SessionId)
}

func (c *MCNetwork) MainProc() {
	fmt.Println("MainProc")
	for {
		time.Sleep(time.Second)
	}
}

// 打包以及加密行为
func (c *MCNetwork) Package(val interface{}) []byte {
	return nil
}

func (c *MCNetwork) Encrypt(data []byte) []byte {
	return data
}

func (c *MCNetwork) Decrypt(data []byte) []byte {
	return data
}

func Test_Listens(t *testing.T) {
	newTcp, err := NewTcpServer(":9234", &MCNetwork{})
	if err != nil {
		t.Error(err)
		return
	}

	newTcp.Runnable()
}
