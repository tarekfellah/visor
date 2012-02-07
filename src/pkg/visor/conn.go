package visor

import (
	"errors"
	"fmt"
	"github.com/soundcloud/doozer"
	"net"
)

type Conn struct {
	Addr *net.TCPAddr
	conn *doozer.Conn
}

func DialConn(addr string) (conn *Conn, rev int64, err error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	dconn, err := doozer.Dial(addr)
	if err != nil {
		return
	}

	rev, err = dconn.Rev()
	if err != nil {
		return
	}

	return &Conn{tcpaddr, dconn}, rev, nil
}

func (c *Conn) Set(path string, rev int64, value []byte) (newrev int64, err error) {
	return c.conn.Set(path, rev, value)
}

func (c *Conn) Exists(path string, rev *int64) (exists bool, pathrev int64, err error) {
	_, pathrev, err = c.conn.Stat(path, rev)
	if err != nil {
		return
	}

	// directories have negative revisions and should also be found
	if pathrev != 0 {
		exists = true
	}

	return
}

func (c *Conn) Create(path string, value []byte) (newrev int64, err error) {
	exists, newrev, err := c.Exists(path, nil)
	if err != nil {
		return
	}
	if exists {
		return newrev, errors.New(fmt.Sprintf("path %s already exists", path))
	}
	return c.Set(path, newrev, value)
}

func (c *Conn) Rev() (int64, error) {
	return c.conn.Rev()
}

func (c *Conn) Get(path string, rev *int64) (value []byte, pathrev int64, err error) {
	return c.conn.Get(path, rev)
}

func (c *Conn) Getdir(path string, rev int64) (keys []string, err error) {
	return c.conn.Getdir(path, rev, 0, -1)
}

func (c *Conn) Wait(path string, rev int64) (event doozer.Event, err error) {
	return c.conn.Wait(path, rev)
}

func (c *Conn) Close() {
	c.conn.Close()
}

func (c *Conn) Del(path string, rev int64) (err error) {
	err = doozer.Walk(c.conn, rev, path, func(path string, f *doozer.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if !f.IsDir {
			e = c.conn.Del(path, rev)
			if e != nil {
				return e
			}
		}

		return nil
	})

	return
}
