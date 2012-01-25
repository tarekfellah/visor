package visor

import (
	"github.com/soundcloud/doozer"
	"net"
	"testing"
)

func setup(path string) (c *Client, conn *doozer.Conn, rev int64) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", DEFAULT_ADDR)
	if err != nil {
		panic(err)
	}
	conn, err = doozer.Dial(DEFAULT_ADDR)
	if err != nil {
		panic(err)
	}

	rev, err = conn.Rev()
	if err != nil {
		panic(err)
	}

	err = conn.Del(path, rev)
	if err != nil {
		panic(err)
	}

	c = &Client{tcpaddr, conn, "/", rev}

	return
}

func TestExists(t *testing.T) {
	c, conn, _ := setup("/client-test")

	exists, err := c.Exists(path)
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Error("path shouldn't exist")
	}

	_, err = conn.Set(path, 0, []byte{})
	if err != nil {
		t.Error(err)
	}

	exists, err = c.Exists(path)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Error("path doesn't exist")
	}
}
