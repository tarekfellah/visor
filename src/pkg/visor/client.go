package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"net"
	"strings"
)

type Client struct {
	Addr *net.TCPAddr
	conn *doozer.Conn
	Root string
	rev  int64
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) Del(path string) (err error) {
	rev, err := c.conn.Rev()
	if err != nil {
		return
	}

	c.rev = rev
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

func (c *Client) Exists(path string) (exists bool, err error) {
	_, rev, err := c.conn.Stat(c.prefixPath(path), nil)
	if err != nil {
		return
	}

	if rev != 0 {
		exists = true
	}

	return
}

func (c *Client) Get(path string) (value string, err error) {
	rev, err := c.conn.Rev()
	if err != nil {
		return
	}

	body, rev, err := c.conn.Get(c.prefixPath(path), &rev)
	if err != nil {
		return
	}
	if rev == 0 {
		err = ErrKeyNotFound

		return
	}

	c.rev = rev

	value = string(body)

	return
}

func (c *Client) Keys(path string) (keys []string, err error) {
	rev, err := c.conn.Rev()
	if err != nil {
		return
	}

	c.rev = rev

	keys, err = c.conn.Getdir(c.prefixPath(path), c.rev, 0, -1)
	if err != nil {
		return
	}

	return
}

func (c *Client) Set(path string, body string) (err error) {
	rev, err := c.conn.Set(c.prefixPath(path), c.rev, []byte(body))
	if err != nil {
		return
	}

	c.rev = rev

	return
}

func (c *Client) GetMulti(path string, keys []string) (values map[string]string, err error) {
	if keys == nil {
		keys, err = c.Keys(path)
	}
	if err != nil {
		return
	}
	values = map[string]string{}

	for i := range keys {
		val, e := c.Get(path + "/" + keys[i])
		if e != nil {
			return nil, e
		}
		values[keys[i]] = val
	}
	return
}
func (c *Client) SetMulti(path string, kvs ...string) (err error) {
	var key string

	for i := range kvs {
		if i%2 == 0 {
			key = kvs[i]
		} else {
			err = c.Set(path+"/"+key, kvs[i])
			if err != nil {
				break
			}
		}
	}
	return
}
func (c *Client) Wait(path string, rev int64) (doozer.Event, error) {
	return c.conn.Wait(path, rev)
}
func (c *Client) String() string {
	return fmt.Sprintf("%#v", c)
}

func (c *Client) prefixPath(p string) (path string) {
	prefix := c.Root
	path = p

	if p == "/" {
		return prefix
	}

	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	if !strings.HasPrefix(p, c.Root) {
		path = prefix + p
	}

	return path
}

// TICKETS

func (c *Client) Tickets() ([]Ticket, error) {
	return nil, nil
}
func (c *Client) HostTickets(addr string) ([]Ticket, error) {
	return nil, nil
}
func (c *Client) WatchTicket(listener chan *Ticket) error {
	return nil
}
