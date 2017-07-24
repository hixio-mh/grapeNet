// 默认Call处理函数
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/24

package grapeNet

import (
	"time"

	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
)

// 创建用户DATA
func defaultCreateUserData() interface{} {
	return nil
}

// 通知连接
func defaultOnAccept(conn *TcpConn) {
	logger.INFO("Default Accept In:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr().String())
}

func defaultOnHandler(conn *TcpConn, ownerPak []byte) {

}

func defaultOnClose(conn *TcpConn) {
	logger.INFO("Default Closed:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr)
}

func defaultOnConnected(conn *TcpConn) {
	logger.INFO("Default Connected:%v", conn.SessionId)
}

// 简易主处理函数
func defaultMainProc() {
	for {
		time.Sleep(time.Second)
	}
}

// 打包以及加密行为
func defaultEncrypt(data []byte) []byte {
	return data
}
func defaultDecrypt(data []byte) []byte {
	return data
}

// 默认的解包数据行为
func DefaultByteData(conn *TcpConn, spak *stream.BufferIO) [][]byte {
	pakData := [][]byte{}

	for {
		pData, err := spak.Unpack(true, conn.ownerNet.Decrypt)
		if err != nil {
			break
		}

		if pData == nil {
			break
		}

		pakData = append(pakData, pData)
	}

	return pakData
}

func DefaultLineData(conn *TcpConn, spak *stream.BufferIO) [][]byte {
	pakData := [][]byte{}

	for {
		pData, err := spak.UnpackLine(true, conn.ownerNet.Decrypt)
		if err != nil {
			break
		}

		if pData == nil {
			break
		}

		pakData = append(pakData, pData)
	}

	return pakData
}
