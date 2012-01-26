package visor

import (
	"github.com/soundcloud/doozer"
	"net"
)

const DEFAULT_ADDR string = "localhost:8046"

type ProcessType string
type Stack string
type State int

func Dial(addr string) (c *Client, err error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	conn, err := doozer.Dial(addr)
	if err != nil {
		return
	}

	rev, err := conn.Rev()
	if err != nil {
		return
	}

	return &Client{tcpaddr, conn, "/", rev}, nil
}
