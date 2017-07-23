// 日志系统，简易版本
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/10

package grapeLogger

import (
	"fmt"
	"os"

	slog "github.com/cihub/seelog"
)

var isBuild = false

func BuildLogger(logDir, logFile string) {
	if isBuild {
		return
	}

	isBuild = true
	os.Mkdir(logDir, 0777)
	realFile := logDir + "/" + logFile
	realErrorFile := logDir + "/error.log"
	realDebugFile := logDir + "/debug.log"

	sConfig := fmt.Sprintf(`
	<seelog type="asynctimer" asyncinterval="1000">
		<outputs formatid="main">  
			<filter levels="info,warn">   
				<console />    
				<file path="%v" />
				<rollingfile type="size"  filename="%v" maxsize="102400" maxrolls="5" />    
			</filter>
			<filter levels="critical,error">
				<console />   
				<file path="%v" />
				<rollingfile type="size"  filename="%v" maxsize="102400" maxrolls="5" />   
			</filter>
			<filter levels="debug">
				<console />   
				<file path="%v" />
				<rollingfile type="size" filename="%v" maxsize="102400" maxrolls="5" />   
			</filter>
		</outputs>
		<formats>
			<format id="main" format="[%%Date %%Time] [%%File.%%Line] [%%LEVEL] %%Msg%%n"/>   
		</formats>
	</seelog>
	`, realFile, realFile, realErrorFile, realErrorFile, realDebugFile, realDebugFile)

	elog, err := slog.LoggerFromConfigAsString(sConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	elog.SetAdditionalStackDepth(1)
	slog.UseLogger(elog)
}

func BuildFromXML(xmlFile string) {
	if isBuild {
		return
	}

	isBuild = true
	elog, err := slog.LoggerFromConfigAsFile(xmlFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	elog.SetAdditionalStackDepth(1)
	slog.UseLogger(elog)
}

func INFO(fmt string, v ...interface{}) {
	slog.Infof(fmt, v...)
}

func DEBUG(fmt string, v ...interface{}) {
	slog.Debugf(fmt, v...)
}

func CRT(fmt string, v ...interface{}) {
	slog.Criticalf(fmt, v...)
}

func WARN(fmt string, v ...interface{}) {
	slog.Warnf(fmt, v...)
}

func ERROR(fmt string, v ...interface{}) {
	slog.Errorf(fmt, v...)
}

func TRACE(fmt string, v ...interface{}) {
	slog.Tracef(fmt, v...)
}

func FLUSH() {
	slog.Flush()
}
