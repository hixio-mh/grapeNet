// 封装的快速构建服务的体系
// 超简单的启动服务基于github.com/takama/daemon
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2018/03/02

package Utils

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/koangel/daemon"
)

type DaemonHandler func(signal chan os.Signal) string

var (
	sd    daemon.Daemon
	usage string

	workRoot string = ""
)

func getCurrPath() string {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		file = os.Args[0]
	}

	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	ret := path[:index]
	return ret
}

func serviceRun(handler DaemonHandler) (string, error) {

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			if len(os.Args) > 2 {
				return sd.Install(os.Args[2:]...) // 附加参数
			}
			return sd.Install()
		case "remove":
			return sd.Remove()
		case "start":
			return sd.Start()
		case "stop":
			return sd.Stop()
		case "status":
			return sd.Status()
		case "restart":
			sd.Stop()
			return sd.Start()
		default:
			return usage, nil
		}
	}

	// 设置工作目录
	if workRoot == "" {
		// 使用当前程序目录
		os.Chdir(getCurrPath())
	} else {
		os.Chdir(workRoot)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	return handler(interrupt), nil
}

func RunDaemon(name string, desc string, sWork string, handler DaemonHandler) error {
	srv, err := daemon.New(name, desc, []string{fmt.Sprintf("%v.service", name)}...)
	if err != nil {
		return err
	}

	usage = fmt.Sprintf("Usage: %v install | remove | start | stop | status", name)
	sd = srv
	workRoot = sWork

	show, serr := serviceRun(handler)
	fmt.Println(show)
	return serr
}
