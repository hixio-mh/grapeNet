package kcpNet

import "testing"

func TestNewEmptyKcp(t *testing.T) {

	cnf := NewConfig()
	kcpNet := NewEmptyKcp(cnf)
	if kcpNet == nil {
		t.Fatal("kcp error...")
	}
}

func TestNewConfig(t *testing.T) {
	cnf := NewConfig()
	if cnf == nil {
		t.Fatal("kcp config error...")
	}

}
