// Etcd封装库，简化各种操作
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/9/26

package grapeEtcd

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
)

var (
	// 序列化和反序列化时的格式函数
	formatter Formatter = &JsonFormatter{}
	// 监听函数以及监听位置
	watchers map[string]*EtcdWatcher = map[string]*EtcdWatcher{}
	wMux     sync.Mutex
	// etcd内部客户端
	EtcdCli *clientv3.Client = nil
	// Close
	once *sync.Once = nil

	// 验证
	IsAuth       bool   = false
	AuthUserName string = "root"
	AuthPassword string = "123123"
)

const (
	writeTimeout = 5 * time.Second
	readTimeout  = 5 * time.Second

	watchMustArgs = 2
)

func Dial(urls []string) error {
	return DialTimeout(urls, 45*time.Second)
}

func DialTimeout(urls []string, timeout time.Duration) error {

	config := clientv3.Config{
		Endpoints:   urls,
		DialTimeout: timeout,
	}

	// 开启验证
	if IsAuth {
		config.Username = AuthUserName
		config.Password = AuthPassword
	}

	cli, err := clientv3.New(config)
	if err != nil {
		return err
	}
	once = &sync.Once{}
	EtcdCli = cli

	return nil
}

func Close() {
	if once != nil {
		once.Do(func() {
			if EtcdCli != nil {
				EtcdCli.Close()
				EtcdCli = nil
			}
		})
	}
}

func SetFormatter(in Formatter) {
	formatter = in
}

///////////////////////////////////////////////////////////////
// 读写函数
func Read(key string) (body []byte, err error) {
	body = nil
	err = nil
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	resp, verr := EtcdCli.Get(ctx, key)
	cancel()
	if verr != nil {
		err = verr
		return
	}

	if len(resp.Kvs) <= 0 {
		err = errors.New("keys is empty...")
		return
	}

	body = resp.Kvs[0].Value
	return
}

func UnmarshalKey(key string, val interface{}) error {
	resp, err := Read(key)
	if err != nil {
		return err
	}

	return formatter.Unmarshal(formatter.FromString(string(resp)), val)
}

func Write(key string, val []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	_, err := EtcdCli.Put(ctx, key, string(val))
	cancel()
	if err != nil {
		return err
	}

	return nil
}

func MarshalKey(key string, val interface{}) error {
	body, err := formatter.Marshal(val)
	if err != nil {
		return err
	}

	return Write(key, []byte(formatter.ToString(body)))
}
