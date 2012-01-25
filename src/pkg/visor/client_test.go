package visor

import (
	"github.com/soundcloud/doozer"
	"net"
	"testing"
)

func setup(path string) (c *Client, conn *doozer.Conn) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", DEFAULT_ADDR)
	if err != nil {
		panic(err)
	}
	conn, err = doozer.Dial(DEFAULT_ADDR)
	if err != nil {
		panic(err)
	}

	c = &Client{tcpaddr, conn, "/", 0}
	c.Del(path)

	return
}

func TestDel(t *testing.T) {
	path := "/del-test"
	c, conn := setup(path)

	_, err := conn.Set(path+"/deep/blue", 0, []byte{})
	if err != nil {
		t.Error(err)
	}

	err = c.Del(path)
	if err != nil {
		t.Error(err)
	}

	_, rev, err := conn.Stat(path, nil)
	if err != nil {
		t.Error(err)
	}

	if rev != 0 {
		t.Error("path isn't deleted")
	}
}

func TestExists(t *testing.T) {
	path := "/exists-test"
	c, conn := setup(path)

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
