// Websocket Call
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/8/3
package grapeWSNet

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
func defaultOnAccept(conn *WSConn) {
	logger.INFO("Default Accept In:%v From:%v", conn.SessionId, conn.WConn.RemoteAddr().String())
}

func defaultOnHandler(conn *WSConn, ownerPak []byte) {

}

func defaultOnClose(conn *WSConn) {
	logger.INFO("Default Closed:%v From:%v", conn.SessionId, conn.WConn.RemoteAddr)
}

func defaultOnConnected(conn *WSConn) {
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
func DefaultByteData(conn *WSConn, spak *stream.BufferIO) [][]byte {
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

func DefaultLineData(conn *WSConn, spak *stream.BufferIO) [][]byte {
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
