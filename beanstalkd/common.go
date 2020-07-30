package beanstalkd

import (
	"github.com/kr/beanstalk"
	"sync"
)

type AddrList struct {
	Addrs []string `toml:"addrs" json:"addrs"`
}

type Conns struct {
	sync.RWMutex
	conns map[string]*beanstalk.Conn
}

func (cs *Conns) Get(addr string) *beanstalk.Conn {
	cs.RLock()
	defer cs.RUnlock()

	return cs.conns[addr]
}

func (cs *Conns) Set(addr string, conn *beanstalk.Conn) {
	cs.Lock()
	defer cs.Unlock()

	cs.conns[addr] = conn
}

func reconnect(addr string) (conn *beanstalk.Conn, err error) {
	return beanstalk.Dial("tcp", addr)
}
