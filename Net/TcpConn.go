// 连接对象
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/10

package grapeNet

import (
	"net"
	"time"

	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
)

type TcpConn struct {
	cm.Conn
	ownerNet *TCPNetwork
	TConn    net.Conn    // tcp连接
	UserData interface{} // 用户对象
	LastPing time.Time

	mainStream stream.BufferIO

	send     chan []byte
	recvPack chan *stream.BufferIO // 单独的一个数据包
}

//////////////////////////////////////
// 创建新连接
func NewConn(tn *TCPNetwork, conn net.Conn, UData interface{}) *TcpConn {
	newConn := &TcpConn{
		ownerNet: tn,
		TConn:    conn,
		UserData: UData,
		LastPing: time.Now(),
	}

	newConn.Done = make(chan int, 1)
	newConn.SessionId = cm.CreateUUID(1)
	newConn.Type = cm.ESERVER_TYPE

	return newConn
}

func NewDial(tn *TCPNetwork, addr string, UData interface{}) (conn *TcpConn, err error) {
	err = nil
	conn = nil
	dconn, derr := net.DialTimeout("tcp", addr, time.Second*300)
	if err != nil {
		logger.ERROR(err.Error())
		err = derr
		return
	}

	conn = &TcpConn{
		ownerNet: tn,
		UserData: UData,
		LastPing: time.Now(),
	}

	conn.TConn = dconn
	conn.Done = make(chan int, 1)
	conn.SessionId = cm.CreateUUID(2)
	conn.Type = cm.ECLIENT_TYPE

	return
}

//////////////////////////////////////////////
// 成员函数
func (c *TcpConn) startProc() {
	go c.writePump()
	go c.recvPump()
}

func (c *TcpConn) recvPump() {
	defer func() {
		c.Done <- 1
		c.Close()
	}()

	var buffer []byte = make([]byte, 65535)

	for {
		rn, err := c.TConn.Read(buffer)
		if err != nil {
			logger.ERROR("Session %v Recv Error:%v", c.SessionId, err)
			return
		}

		if rn == 0 {
			logger.ERROR("Session %v Recv Len:%v", c.SessionId, rn)
			return
		}

		c.mainStream.Write(buffer, rn)
		for {
			buf, berr := c.mainStream.Unpack(true)
			if berr != nil {
				break
			}

			c.recvPack <- buf // 缓冲期
		}
	}
}

func (c *TcpConn) writePump() {

}

func (c *TcpConn) processHandler() {

}

func (c *TcpConn) Send(data []byte, len int) int {
	return -1
}

func (c *TcpConn) Close() {

}

func (c *TcpConn) RemoveData() {

}

func (c *TcpConn) InitData() {

}
