package kcpNet

import (
	"encoding/json"
	"os"
)

const (
	interJson = `
	{
		"crypt": "xor",
		"mode": "fast3",
		"key": "2b6f5b9eeb34f5a1f69caa00558ea833",
		"autoexpire":0,
		"scavengettl":600,
		"mtu": 1350,
		"sndwnd": 256,
		"rcvwnd": 256,
		"datashard": 10,
		"parityshard": 3,
		"dscp": 0,
		"nocomp": false,
		"acknodelay": false,
		"nodelay": 0,
		"interval": 50,
		"resend": 0,
		"nc": 0,
		"sockbuf": 4194304,
		"keepalive": 10,
		"log": true,
		"snmplog": "",
		"snmpperiod": 60,
		"pprof": false,
		"quiet": true
	}
	`
)

// Config for server
type KcpConfig struct {
	AutoExpire   int    `json:"autoexpire"`
	ScavengeTTL  int    `json:"scavengettl"`
	Crypt        string `json:"crypt"`
	CryptKey     string `json:"key"`
	Mode         string `json:"mode"`
	MTU          int    `json:"mtu"`
	SndWnd       int    `json:"sndwnd"`
	RcvWnd       int    `json:"rcvwnd"`
	DataShard    int    `json:"datashard"`
	ParityShard  int    `json:"parityshard"`
	DSCP         int    `json:"dscp"`
	NoComp       bool   `json:"nocomp"`
	AckNodelay   bool   `json:"acknodelay"`
	NoDelay      int    `json:"nodelay"`
	Interval     int    `json:"interval"`
	Resend       int    `json:"resend"`
	NoCongestion int    `json:"nc"`
	SockBuf      int    `json:"sockbuf"`
	KeepAlive    int    `json:"keepalive"`
	Log          bool   `json:"log"`
	SnmpLog      string `json:"snmplog"`
	SnmpPeriod   int    `json:"snmpperiod"`
	Pprof        bool   `json:"pprof"`
	Quiet        bool   `json:"quiet"`
}

func NewConfig() *KcpConfig {
	conf := &KcpConfig{}
	err := json.Unmarshal([]byte(interJson), conf)
	if err != nil {
		return nil
	}

	return conf
}

func ParseJSONConfig(config *KcpConfig, path string) error {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}
