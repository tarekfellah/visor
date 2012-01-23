package visor

import (
	"github.com/ha/doozer"
	"net"
)

const DEFAULT_ADDR string = "localhost:8046"

type ProcessType string
type State string

func Dial(addr string) (c *Client, err error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
	  return
  }

	conn, err := doozer.Dial(addr)
	if err != nil {
		return
	}

	return &Client{tcpaddr, conn, "/"}, nil
}
