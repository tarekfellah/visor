package visor

import (
	"github.com/soundcloud/doozer"
	"net"
	"strings"
)

const DEFAULT_ADDR string = "localhost:8046"

type ProcessType string
type Stack string
type State int

func Dial(path string) (c *Client, err error) {
	parts := strings.SplitN(path, "/", 2)
	addr, root := parts[0], "/"

	if len(parts) == 2 {
		root += parts[1]
	}

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

	return &Client{tcpaddr, conn, root, rev}, nil
}
