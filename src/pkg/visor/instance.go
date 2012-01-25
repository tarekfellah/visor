package visor

import (
	"net"
	"strings"
)

type Instance struct {
	Rev         *Revision
	Addr        *net.TCPAddr
	State       State
	ProcessType ProcessType
}

func (i *Instance) String() string {
	return "<instance>"
}
func (i *Instance) Register(c *Client) (err error) {
	exists, err := c.Exists(i.Path())
	if err != nil {
		return
	}
	if exists {
		return ErrKeyConflict
	}

	err = c.Set(i.Path()+"/host", string(i.Addr.IP))
	if err != nil {
		return
	}
	err = c.Set(i.Path()+"/port", string(i.Addr.Port))
	if err != nil {
		return
	}
	err = c.Set(i.Path()+"/process-type", string(i.ProcessType))
	if err != nil {
		return
	}
	err = c.Set(i.Path()+"/state", string(i.State))

	return
}
func (i *Instance) Unregister(c *Client) (err error) {
	return c.Del(i.Path())
}
func (i *Instance) Path() (path string) {
	id := strings.Replace(strings.Replace(i.Addr.String(), ".", "-", -1), ":", "-", -1)

	return i.Rev.Path() + "/" + id
}
