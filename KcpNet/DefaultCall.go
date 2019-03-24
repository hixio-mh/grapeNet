package kcpNet

import (
	"bytes"
	"time"

	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
	bson "gopkg.in/mgo.v2/bson"
)

// 创建用户DATA
func defaultCreateUserData() interface{} {
	return nil
}

// 通知连接
func defaultOnAccept(conn *KcpConn) {
	logger.INFO("Default Accept In:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr().String())
}

func defaultOnClose(conn *KcpConn) {
	if conn.TConn == nil {
		return
	}

	logger.INFO("Default Closed:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr)
}

func defaultOnConnected(conn *KcpConn) {
	logger.INFO("Default Connected:%v", conn.SessionId)
}

// 简易主处理函数
func defaultMainProc() {
	for {
		time.Sleep(time.Second)
	}
}

// 打包以及加密行为
func defaultEncrypt(data, key []byte) []byte {
	return data
}
func defaultDecrypt(data, key []byte) []byte {
	return data
}

func defaultBytePacker(val interface{}) (data []byte, err error) {
	data = []byte{}
	err = nil
	sbuf, serr := bson.Marshal(val)
	if serr != nil {
		logger.ERRORV(serr)
		err = serr
		return
	}

	data = sbuf
	return
}

func defaultPanic(conn *KcpConn, src string) {

}

// 默认的解包数据行为
func defaultByteData(conn *KcpConn, spak *stream.BufferIO) (data [][]byte, err error) {
	data = [][]byte{}
	total := int(0)
	err = nil

	for {
		pData, uerr := spak.Unpack(conn.ownerNet.Decrypt, conn.CryptKey)
		if uerr != nil {
			if total <= 0 {
				err = uerr
			}
			break
		}

		if pData == nil {
			break
		}

		total += len(pData)
		data = append(data, pData)
	}

	if total > 0 {
		spak.Reset()
	}

	return
}

func defaultLineData(conn *KcpConn, spak *stream.BufferIO) (data [][]byte, err error) {
	data = [][]byte{}

	for {
		pData, uerr := spak.UnpackLine(conn.ownerNet.Decrypt, conn.CryptKey)
		if uerr != nil {
			err = uerr
			break
		}

		if pData == nil {
			break
		}

		data = append(data, pData)
	}

	if len(data) > 0 {
		spak.Reset()
	}

	return
}

func defaultPing(conn *KcpConn) {
	conn.Send([]byte("ping"))
}

func defalutPong(conn *KcpConn, ping []byte) bool {
	if bytes.HasSuffix(ping, []byte("pong")) {
		return true
	}

	if bytes.HasSuffix(ping, []byte("ping")) {
		conn.Send([]byte("pong"))
		return true
	}

	return false
}
