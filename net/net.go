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
	c, err := _net.DialTimeout("tcp", addr, time.Second)

	if err != nil {
		return c, err
	}

	return &conn{c}, nil
}

type conn struct {
	_net.Conn
}

func (c *conn) Write(b []byte) (int, error) {
	if err := c.Conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		return 0, err
	}

	n, err := c.Conn.Write(b)

	if err == nil {
		c.Conn.SetWriteDeadline(time.Time{})
	}

	return n, err
}

func (c *conn) Read(b []byte) (int, error) {
	if err := c.Conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		return 0, err
	}

	n, err := c.Conn.Read(b)

	if err == nil {
		c.Conn.SetReadDeadline(time.Time{})
	}

	return n, err
}
