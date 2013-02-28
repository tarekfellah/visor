package net

import (
	"io"
	_net "net"
	"time"
)

type Network interface {
	Dial(addr string) (io.ReadWriteCloser, error)
}

type Net struct{}

func (n *Net) Dial(addr string) (io.ReadWriteCloser, error) {
	return _net.DialTimeout("tcp", addr, time.Second)
}
