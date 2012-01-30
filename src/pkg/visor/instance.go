package visor

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
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
func (i *Instance) Register(c *Client) (err error) {
	exists, err := c.Exists(i.Path())
	if err != nil {
		return
	}
	if exists {
		return ErrKeyConflict
	}

	err = c.SetMulti(i.Path(),
		"registered", time.Now().UTC().String(),
		"host", i.Addr.IP.String(),
		"port", strconv.Itoa(i.Addr.Port),
		"process-type", string(i.ProcessType),
		"state", strconv.Itoa(int(i.State)))

	return
}
func (i *Instance) Unregister(c *Client) (err error) {
	return c.Del(i.Path())
}
func (i *Instance) Path() (path string) {
	id := strings.Replace(strings.Replace(i.Addr.String(), ".", "-", -1), ":", "-", -1)

	return i.Rev.Path() + "/" + id
}

func (i *Instance) String() string {
	return fmt.Sprintf("%#v", i)
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

	for i := range names {
		vals, e := c.GetMulti(r.Path()+"/"+names[i], nil)

		if e != nil {
			return nil, e
		}

		s, e := strconv.Atoi(vals["state"])
		if e != nil {
			return nil, e
		}

		addr := vals["host"] + ":" + vals["port"]

		instances[i], err = NewInstance(r, addr, ProcessType(vals["process-type"]), State(s))
		if err != nil {
			return
		}
	}

	return
}
func (c *Client) HostInstances(addr string) ([]Instance, error) {
	return nil, nil
}
