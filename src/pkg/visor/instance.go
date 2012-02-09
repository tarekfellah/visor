package visor

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// An Instance represents a running process of a specific type.
// Instances belong to Revisions.
type Instance struct {
	Snapshot
	Rev         *Revision    // Revision the instance belongs to
	Addr        *net.TCPAddr // TCP address of the running instance
	State       State        // Current state of the instance
	ProcessType ProcessType  // Type of process the instance represents
}

const (
	InsStateInitial State = 0
)

// NewInstance creates and returns a new Instance object.
func NewInstance(rev *Revision, addr string, pType ProcessType, state State, snapshot Snapshot) (ins *Instance, err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	ins = &Instance{Rev: rev, Addr: tcpAddr, ProcessType: pType, State: state, Snapshot: snapshot}

	return
}

// FastForward returns a copy of the current instance, with its
// revision set to the supplied one.
func (i *Instance) FastForward(rev int64) *Instance {
	return i.Snapshot.FastForward(i, rev).(*Instance)
}

func (i *Instance) CreateSnapshot(rev int64) Snapshotable {
	return &Instance{Rev: i.Rev, Addr: i.Addr, State: i.State, ProcessType: i.ProcessType, Snapshot: Snapshot{rev: rev, conn: i.conn}}
}

// Register registers an instance with the registry.
func (i *Instance) Register() (instance *Instance, err error) {
	exists, _, err := i.conn.Exists(i.Path(), &i.rev)
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}
	if i.State != InsStateInitial {
		return nil, ErrInvalidState
	}

	rev, err := i.conn.SetMulti(i.Path(), map[string][]byte{
		"registered":   []byte(time.Now().UTC().String()),
		"host":         []byte(i.Addr.IP.String()),
		"port":         []byte(strconv.Itoa(i.Addr.Port)),
		"process-type": []byte(string(i.ProcessType)),
		"state":        []byte(strconv.Itoa(int(i.State)))}, i.rev)

	if err != nil {
		return i, err
	}
	instance = i.FastForward(rev)

	return
}

// Unregister unregisters an instance with the registry.
func (i *Instance) Unregister() (err error) {
	return i.conn.Del(i.Path(), i.rev)
}

// Path returns the instance's directory path in the registry.
func (i *Instance) Path() (path string) {
	id := strings.Replace(strings.Replace(i.Addr.String(), ".", "-", -1), ":", "-", -1)

	return i.Rev.Path() + "/" + id
}

func (i *Instance) String() string {
	return fmt.Sprintf("%#v", i)
}

// Instances returns returns an array of all registered instances.
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

// RevisionInstances returns an array of all registered instances belonging
// to the given revision.
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

		s, e := strconv.Atoi(vals["state"].Value.String())
		if e != nil {
			return nil, e
		}

		addr := vals["host"].Value.String() + ":" + vals["port"].Value.String()

		instances[i], err = NewInstance(r, addr, ProcessType(vals["process-type"].Value.String()), State(s), c.Snapshot)
		if err != nil {
			return
		}
	}

	return
}

// HostInstances returns an array of all registered instances belonging
// to the given host.
func (c *Client) HostInstances(addr string) ([]Instance, error) {
	return nil, nil
}
