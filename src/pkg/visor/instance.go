package visor

import (
	"net"
	"strconv"
	"strings"
)

type Instance struct {
	Rev         *Revision
	Addr        *net.TCPAddr
	State       State
	ProcessType ProcessType
}

func NewInstance(rev *Revision, addr string, pType ProcessType, state State) (ins *Instance, err error) {

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	ins = &Instance{Rev: rev, Addr: tcpAddr, ProcessType: pType, State: state}

	return
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

	err = c.Set(i.Path()+"/host", i.Addr.IP.String())
	if err != nil {
		return
	}
	err = c.Set(i.Path()+"/port", strconv.Itoa(i.Addr.Port))
	if err != nil {
		return
	}
	err = c.Set(i.Path()+"/process-type", string(i.ProcessType))
	if err != nil {
		return
	}
	err = c.Set(i.Path()+"/state", strconv.Itoa(int(i.State)))

	return
}
func (i *Instance) Unregister(c *Client) (err error) {
	return c.Del(i.Path())
}
func (i *Instance) Path() (path string) {
	id := strings.Replace(strings.Replace(i.Addr.String(), ".", "-", -1), ":", "-", -1)

	return i.Rev.Path() + "/" + id
}

func Instances(c *Client) (instances []*Instance, err error) {
	revs, err := Revisions(c)
	if err != nil {
		return
	}

	instances = []*Instance{}

	for i := range revs {
		iss, e := RevisionInstances(c, revs[i])
		if e != nil {
			return nil, e
		}
		instances = append(instances, iss...)
	}

	return
}
func RevisionInstances(c *Client, r *Revision) (instances []*Instance, err error) {
	names, err := c.Keys(r.Path())
	if err != nil {
		return
	}

	instances = make([]*Instance, len(names))

	var (
		host  string
		port  string
		pType string
		state string
		s     int
	)

	for i := range names {
		iPath := r.Path() + "/" + names[i]
		host, err = c.Get(iPath + "/host")
		if err != nil {
			return
		}
		port, err = c.Get(iPath + "/port")
		if err != nil {
			return
		}
		pType, err = c.Get(iPath + "/process-type")
		if err != nil {
			return
		}
		state, err = c.Get(iPath + "/state")
		if err != nil {
			return
		}

		s, err = strconv.Atoi(state)
		if err != nil {
			return
		}

		instances[i], err = NewInstance(r, host+":"+port, ProcessType(pType), State(s))
		if err != nil {
			return
		}
	}

	return
}
func (c *Client) HostInstances(addr string) ([]Instance, error) {
	return nil, nil
}
