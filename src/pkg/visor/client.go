package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"reflect"
	"strings"
)

type Client struct {
	Snapshot
	Root  string
	codec Codec
}

// NewClient creates and returns a new Client object with a ByteCodec registered on /
func NewClient(conn *Conn, root string, rev int64, codec Codec) *Client {
	c := &Client{Root: root, codec: codec, Snapshot: Snapshot{rev, conn}}
	return c
}

// FastForward returns a copy of the current client with its.Revision
// set to the specified revision. If -1 is passed as the revision,
// it will advance to the latest revision.
func (c *Client) FastForward(rev int64) (client *Client, err error) {
	return c.Snapshot.fastForward(c, rev).(*Client), nil
}

func (c *Client) CreateSnapshot(rev int64) Snapshotable {
	return NewClient(c.conn, c.Root, rev, c.codec)
}

// Close tears down the internal coordinator connection
func (c *Client) Close() {
	c.conn.Close()
}

// Del deletes the specified path
func (c *Client) Del(path string) error {
	return c.conn.Del(c.prefixPath(path), c.Rev)
}

// Exists checks if the specified path exists at this client's
// revision.
func (c *Client) Exists(path string) (exists bool, err error) {
	exists, _, err = c.conn.Exists(c.prefixPath(path), &c.Rev)
	return
}

// Get returns the value for the given path
func (c *Client) Get(path string) (file *File, err error) {
	path = c.prefixPath(path)

	evalue, filerev, err := c.conn.Get(path, &c.Rev)
	if err != nil {
		return
	}

	value, err := c.codec.Decode(evalue)
	if err != nil {
		return
	}

	file = NewFile(path, filerev, reflect.ValueOf(value))

	return
}

// Keys returns all keys for the given path
func (c *Client) Keys(path string) (keys []string, err error) {
	keys, err = c.conn.Getdir(c.prefixPath(path), c.Rev)
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

	evalue, err = c.codec.Encode(value)
	if err != nil {
		return
	}

	filerev, err := c.conn.Set(path, c.Rev, evalue)
	if err != nil {
		return
	}

	file = NewFile(path, filerev, reflect.ValueOf(value))
	return
}

// GetMulti returns multiple key/value pairs organized in map
func (c *Client) GetMulti(path string, keys []string) (filevalues map[string]*File, err error) {
	values, err := c.conn.GetMulti(path, keys, c.Rev)
	if err != nil {
		return
	}

	filevalues = make(map[string]*File)

	for i := range keys {
		key := keys[i]
		filevalues[key] = NewFile(path+"/"+keys[i], c.Rev, reflect.ValueOf(values[key]))
	}
	return
}

func (c *Client) SetMulti(path string, kvs map[string][]byte) (newrev int64, err error) {
	return c.conn.SetMulti(path, kvs, c.Rev)
}

// Waits for the first change, on or after rev, to any file matching path
func (c *Client) Wait(path string, rev int64) (ev doozer.Event, newclient *Client, err error) {
	newclient, err = c.FastForward(rev)
	if err != nil {
		return
	}
	ev, err = newclient.conn.Wait(path, newclient.Rev)
	newclient, err = newclient.FastForward(ev.Rev)
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
