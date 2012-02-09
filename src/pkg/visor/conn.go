package visor

import (
	"errors"
	"fmt"
	"github.com/soundcloud/doozer"
	"net"
	"strings"
)

type Conn struct {
	Addr *net.TCPAddr
	Root string
	conn *doozer.Conn
}

func DialConn(addr string, root string) (s Snapshot, err error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	dconn, err := doozer.Dial(addr)
	if err != nil {
		return
	}

	rev, err := dconn.Rev()
	if err != nil {
		return
	}

	s = Snapshot{rev, &Conn{tcpaddr, root, dconn}}
	return
}

func (c *Conn) Set(path string, rev int64, value []byte) (newrev int64, err error) {
	return c.conn.Set(c.prefixPath(path), rev, value)
}

func (c *Conn) Exists(path string, rev *int64) (exists bool, pathrev int64, err error) {
	_, pathrev, err = c.conn.Stat(c.prefixPath(path), rev)
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
	path = c.prefixPath(path)
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

func (c *Conn) Get(path string, rev *int64) (value []byte, filerev int64, err error) {
	value, filerev, err = c.conn.Get(c.prefixPath(path), rev)
	if filerev == 0 {
		err = ErrKeyNotFound
	}
	return
}

func (c *Conn) Getdir(path string, rev int64) (keys []string, err error) {
	return c.conn.Getdir(c.prefixPath(path), rev, 0, -1)
}

func (c *Conn) Wait(path string, rev int64) (event doozer.Event, err error) {
	path = c.prefixPath(path)
	event, err = c.conn.Wait(path, rev)
	event.Path = strings.Replace(event.Path, c.Root, "", 1)
	return
}

func (c *Conn) Close() {
	c.conn.Close()
}

func (c *Conn) Del(path string, rev int64) (err error) {
	path = c.prefixPath(path)

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

// GetMulti returns multiple key/value pairs organized in map
func (c *Conn) GetMulti(path string, keys []string, rev int64) (values map[string][]byte, err error) {
	if keys == nil {
		keys, err = c.Getdir(path, rev)
	}
	if err != nil {
		return
	}
	values = make(map[string][]byte)

	for i := range keys {
		val, _, e := c.Get(path+"/"+keys[i], &rev)
		if e != nil {
			return nil, e
		}
		values[keys[i]] = val
	}
	return
}

// SetMulti stores mutliple key/value pairs under the given path
func (c *Conn) SetMulti(path string, kvs map[string][]byte, rev int64) (newrev int64, err error) {
	for k, v := range kvs {
		newrev, err = c.Set(path+"/"+k, rev, v)
		if err != nil {
			break
		}
	}
	return
}

func (c *Conn) prefixPath(p string) (path string) {
	prefix := c.Root
	path = p

	if p == "/" {
		return prefix
	}

	if !strings.HasSuffix(prefix, "/") && !strings.HasPrefix(p, "/") {
		prefix = prefix + "/"
	}

	if !strings.HasPrefix(p, c.Root) {
		path = prefix + p
	}

	return path
}
