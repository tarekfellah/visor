package visor

import (
	"github.com/ha/doozer"
	"net"
)

type ProcessType string
type State string

func Dial(addr string) (*Client, error) {
	conn, err := doozer.Dial(addr)

	if err != nil {
		return nil, err
	}
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)

	return &Client{tcpaddr, conn, "/"}, nil
}
