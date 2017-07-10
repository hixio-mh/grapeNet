// 日志系统，简易版本
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/8

package grapeLogger

import (
	"os"

	l4g "github.com/alecthomas/log4go"
)

var isBuild = false

func BuildLogger(logDir, logFile string) {
	if isBuild {
		return
	}

	isBuild = true
	os.Mkdir(logDir, 0777)
	realFile := logDir + "/" + logFile

	l4g.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())          //输出到控制台,级别为DEBUG
	l4g.AddFilter("file", l4g.DEBUG, l4g.NewFileLogWriter(realFile, true)) //输出到文件,级别为DEBUG,每次追加该原文件 滚动文件
}

func BuildFromXML(xmlFile string) {
	if isBuild {
		return
	}

	isBuild = true
	l4g.LoadConfiguration(xmlFile)
}

func INFO(fmt string, v ...interface{}) {
	l4g.Info(fmt, v)
}

func DEBUG(fmt string, v ...interface{}) {
	l4g.Debug(fmt, v...)
}

func CRT(fmt string, v ...interface{}) {
	l4g.Critical(fmt, v...)
}

func WARN(fmt string, v ...interface{}) {
	l4g.Warn(fmt, v...)
}

func ERROR(fmt string, v ...interface{}) {
	l4g.Error(fmt, v...)
}

func TRACE(fmt string, v ...interface{}) {
	l4g.Trace(fmt, v...)
}
