package visor

import (
	"errors"
	"fmt"
	"github.com/soundcloud/doozer"
	"net"
	"reflect"
	"strings"
)

type Client struct {
	Addr       *net.TCPAddr
	conn       *doozer.Conn
	Root       string
	Rev        int64
	pathCodecs []struct {
		path  string
		codec Codec
	}
}

// NewClient creates and returns a new Client object with a ByteCodec registered on /
func NewClient(addr *net.TCPAddr, conn *doozer.Conn, root string, rev int64) *Client {
	c := &Client{addr, conn, root, rev, nil}
	c.RegisterCodec("/", new(ByteCodec))
	return c
}

// Close tears down the internal coordinator connection
func (c *Client) Close() {
	c.conn.Close()
}

// Del deletes a given path in the coordinator
func (c *Client) Del(path string) (err error) {
	rev, err := c.conn.Rev()
	if err != nil {
		return
	}

	c.Rev = rev
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

// Exists checks if the given path is present in the coordinator
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

// Get returns the value for the given path
func (c *Client) Get(path string) (value reflect.Value, err error) {
	rev, err := c.conn.Rev()
	if err != nil {
		return
	}
	path = c.prefixPath(path)

	ebody, rev, err := c.conn.Get(path, &rev)
	if err != nil {
		return
	}
	if rev == 0 {
		err = ErrKeyNotFound

		return
	}

	codec := c.codecForPath(path)
	if codec == nil {
		err = errors.New("couldn't find codec to decode path " + path)
		return
	}

	body, err := codec.Decode(ebody)
	if err != nil {
		return
	}

	c.Rev = rev

	value = reflect.ValueOf(body)

	return
}

// Keys returns all keys for the given path
func (c *Client) Keys(path string) (keys []string, err error) {
	rev, err := c.conn.Rev()
	if err != nil {
		return
	}

	c.Rev = rev

	keys, err = c.conn.Getdir(c.prefixPath(path), c.Rev, 0, -1)
	if err != nil {
		return
	}

	return
}

// Set stores the given body for the given path after encoding
// it with a matching encoder.
func (c *Client) Set(path string, body interface{}) (err error) {
	var ebody []byte

	path = c.prefixPath(path)
	codec := c.codecForPath(path)

	if codec == nil {
		return errors.New("couldn't find codec for path " + path)
	}

	ebody, err = codec.Encode(body)
	if err != nil {
		return
	}

	rev, err := c.conn.Set(path, c.Rev, ebody)
	if err != nil {
		return
	}

	c.Rev = rev

	return
}

// codecForPath iterates over c.pathCodecs in reverse
// order, until it finds a matching path, which it returns
// the Codec for.
func (c *Client) codecForPath(path string) Codec {
	//fmt.Println("looking for " + path)
	//fmt.Printf("%#v\n", c.pathCodecs)
	for i := len(c.pathCodecs) - 1; i >= 0; i-- {
		pc := c.pathCodecs[i]
		if strings.HasPrefix(path, pc.path) {
			return pc.codec
		}
	}
	return nil
}

// GetMulti returns multiple key/value pairs organized in map
func (c *Client) GetMulti(path string, keys []string) (values map[string]reflect.Value, err error) {
	if keys == nil {
		keys, err = c.Keys(path)
	}
	if err != nil {
		return
	}
	values = make(map[string]reflect.Value)

	for i := range keys {
		val, e := c.Get(path + "/" + keys[i])
		if e != nil {
			return nil, e
		}
		values[keys[i]] = val
	}
	return
}

// SetMulti stores mutliple key/value pairs under the given path
func (c *Client) SetMulti(path string, kvs map[string][]byte) (err error) {
	for k, v := range kvs {
		err = c.Set(path+"/"+k, v)
		if err != nil {
			break
		}
	}
	return
}

// RegisterCodec registers codec at the given path
func (c *Client) RegisterCodec(path string, codec Codec) (err error) {
	c.pathCodecs = append(c.pathCodecs, struct {
		path  string
		codec Codec
	}{c.prefixPath(path), codec})
	return
}

// Waits for the first change, on or after rev, to any file matching path
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

	if !strings.HasSuffix(prefix, "/") && !strings.HasPrefix(p, "/") {
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
