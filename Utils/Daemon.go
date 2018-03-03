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
	"os/signal"
	"syscall"

	"github.com/takama/daemon"
)

type DaemonHandler func() string

var (
	sd    daemon.Daemon
	usage string
)

func serviceRun(handler DaemonHandler) (string, error) {

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return sd.Install()
		case "remove":
			return sd.Remove()
		case "start":
			return sd.Start()
		case "stop":
			return sd.Stop()
		case "status":
			return sd.Status()
		default:
			return usage, nil
		}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	return handler(), nil
}

func RunDaemon(name string, desc string, handler DaemonHandler) error {
	srv, err := daemon.New(name, desc, []string{fmt.Sprintf("%v.service", name)}...)
	if err != nil {
		return err
	}

	usage = fmt.Sprintf("Usage: %v install | remove | start | stop | status", name)
	sd = srv

	show, serr := serviceRun(handler)
	fmt.Println(show)
	return serr
}
