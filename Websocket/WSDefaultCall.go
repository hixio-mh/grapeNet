package grapeWSNet

// Websocket Call
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/8/3

import (
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
	"gopkg.in/mgo.v2/bson"
)

// 创建用户DATA
func defaultCreateUserData() interface{} {
	return nil
}

// 通知连接
func defaultOnAccept(conn *WSConn) {
	logger.INFO("Default Accept In:%v From:%v", conn.SessionId, conn.WConn.RemoteAddr().String())
}

func defaultOnClose(conn *WSConn) {
	logger.INFO("Default Closed:%v From:%v", conn.SessionId, conn.WConn.RemoteAddr)
}

func defaultOnConnected(conn *WSConn) {
	logger.INFO("Default Connected:%v", conn.SessionId)
}

// 打包以及加密行为
func defaultEncrypt(data []byte) []byte {
	return data
}
func defaultDecrypt(data []byte) []byte {
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

func defaultPanic(conn *WSConn, src string) {
	logger.ERROR(src)
}

// 默认的解包数据行为
func defaultByteData(conn *WSConn, spak *stream.BufferIO) (data [][]byte, err error) {
	data = [][]byte{}
	total := int(0)

	for {
		pData, uerr := spak.Unpack(conn.ownerNet.Decrypt)
		if uerr != nil {
			err = uerr
			break
		}

		if pData == nil {
			break
		}

		total += len(pData)
		data = append(data, pData)
	}

	spak.Reset()

	return
}

func defaultLineData(conn *WSConn, spak *stream.BufferIO) (data [][]byte, err error) {
	data = [][]byte{}

	for {
		pData, uerr := spak.UnpackLine(conn.ownerNet.Decrypt)
		if uerr != nil {
			err = uerr
			break
		}

		if pData == nil {
			break
		}

		data = append(data, pData)
	}

	spak.Reset()

	return
}
