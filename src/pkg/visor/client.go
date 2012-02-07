package visor

import (
	"errors"
	"fmt"
	"github.com/soundcloud/doozer"
	"reflect"
	"strings"
)

type Client struct {
	conn       *Conn
	Root       string
	rev        int64
	pathCodecs []struct {
		path  string
		codec Codec
	}
}

// NewClient creates and returns a new Client object with a ByteCodec registered on /
func NewClient(conn *Conn, root string, rev int64) *Client {
	c := &Client{conn, root, rev, nil}
	c.RegisterCodec("/", new(ByteCodec))
	return c
}

// FastForward returns a copy of the current client with its revision
// set to the specified revision. If -1 is passed as the revision,
// it will advance to the latest revision.
func (c *Client) FastForward(rev int64) (client *Client, err error) {
	if rev == -1 {
		rev, err = c.conn.Rev()
		if err != nil {
			return c, err
		}
	} else if rev < c.rev {
		return c, errors.New("revision must be greater than client rev")
	}

	client = NewClient(c.conn, c.Root, rev)
	// TODO: use a reference instead
	client.pathCodecs = c.pathCodecs
	return client, nil
}

// Close tears down the internal coordinator connection
func (c *Client) Close() {
	c.conn.Close()
}

// Del deletes the specified path
func (c *Client) Del(path string) error {
	return c.conn.Del(c.prefixPath(path), c.rev)
}

// Exists checks if the specified path exists at this client's
// revision.
func (c *Client) Exists(path string) (exists bool, err error) {
	exists, _, err = c.conn.Exists(c.prefixPath(path), &c.rev)
	return
}

// Get returns the value for the given path
func (c *Client) Get(path string) (file *File, err error) {
	path = c.prefixPath(path)

	codec := c.codecForPath(path)
	if codec == nil {
		err = errors.New("couldn't find codec to decode path " + path)
		return
	}

	evalue, filerev, err := c.conn.Get(path, &c.rev)
	if err != nil {
		return
	}
	if filerev == 0 {
		err = ErrKeyNotFound
		return
	}

	value, err := codec.Decode(evalue)
	if err != nil {
		return
	}

	file = NewFile(path, filerev, reflect.ValueOf(value))

	return
}

// Keys returns all keys for the given path
func (c *Client) Keys(path string) (keys []string, err error) {
	keys, err = c.conn.Getdir(c.prefixPath(path), c.rev)
	if err != nil {
		return
	}

	return
}

// Set stores the given value for the given path after encoding
// it with a matching encoder.
func (c *Client) Set(path string, value interface{}) (file *File, err error) {
	var evalue []byte

	path = c.prefixPath(path)
	codec := c.codecForPath(path)
	if codec == nil {
		return nil, errors.New("couldn't find codec for path " + path)
	}

	evalue, err = codec.Encode(value)
	if err != nil {
		return
	}

	filerev, err := c.conn.Set(path, c.rev, evalue)
	if err != nil {
		return
	}

	file = NewFile(path, filerev, reflect.ValueOf(value))
	return
}

// codecForPath iterates over c.pathCodecs in reverse
// order, until it finds a matching path, which it returns
// the Codec for.
func (c *Client) codecForPath(path string) Codec {
	for i := len(c.pathCodecs) - 1; i >= 0; i-- {
		pc := c.pathCodecs[i]
		if strings.HasPrefix(path, pc.path) {
			return pc.codec
		}
	}
	return nil
}

// GetMulti returns multiple key/value pairs organized in map
func (c *Client) GetMulti(path string, keys []string) (values map[string]*File, err error) {
	if keys == nil {
		keys, err = c.Keys(path)
	}
	if err != nil {
		return
	}
	values = make(map[string]*File)

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
func (c *Client) SetMulti(path string, kvs map[string][]byte) (file *File, err error) {
	for k, v := range kvs {
		file, err = c.Set(path+"/"+k, v)
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
func (c *Client) Wait(path string, rev int64) (ev doozer.Event, newclient *Client, err error) {
	newclient, err = c.FastForward(rev)
	if err != nil {
		return
	}
	ev, err = newclient.conn.Wait(path, newclient.rev)
	return
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
