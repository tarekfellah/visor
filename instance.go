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
	ProcType *ProcType    // ProcType the instance belongs to
	Addr     *net.TCPAddr // TCP address of the running instance
	State    State        // Current state of the instance
}

const (
	InsStateInitial State = 0
)

// NewInstance creates and returns a new Instance object.
func NewInstance(ptype *ProcType, addr string, state State, snapshot Snapshot) (ins *Instance, err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	ins = &Instance{Addr: tcpAddr, ProcType: ptype, State: state, Snapshot: snapshot}

	return
}

// FastForward advances the instance in time. It returns
// a new instance of Instance with the supplied revision.
func (i *Instance) FastForward(rev int64) *Instance {
	return i.Snapshot.fastForward(i, rev).(*Instance)
}

func (i *Instance) createSnapshot(rev int64) Snapshotable {
	return &Instance{Addr: i.Addr, State: i.State, ProcType: i.ProcType, Snapshot: Snapshot{rev, i.conn}}
}

// Register registers an instance with the registry.
func (i *Instance) Register() (instance *Instance, err error) {
	exists, _, err := i.conn.Exists(i.Path(), &i.Rev)
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
		"host":  []byte(i.Addr.IP.String()),
		"port":  []byte(strconv.Itoa(i.Addr.Port)),
		"state": []byte(strconv.Itoa(int(i.State)))}, i.Rev)
	if err != nil {
		return i, err
	}

	rev, err = i.conn.Set(i.Path()+"/registered", i.Rev, []byte(time.Now().UTC().String()))
	if err != nil {
		return i, err
	}
	instance = i.FastForward(rev)

	return
}

// Unregister unregisters an instance with the registry.
func (i *Instance) Unregister() (err error) {
	return i.conn.Del(i.Path(), i.Rev)
}

// Path returns the instance's directory path in the registry.
func (i *Instance) Path() (path string) {
	id := strings.Replace(strings.Replace(i.Addr.String(), ".", "-", -1), ":", "-", -1)

	return i.ProcType.Path() + "/instances/" + id
}

func (i *Instance) String() string {
	return fmt.Sprintf("%#v", i)
}

// Instances returns returns an array of all registered instances.
func Instances(s Snapshot) (instances []*Instance, err error) {
	ptypes, err := ProcTypes(s)
	if err != nil {
		return
	}

	instances = []*Instance{}

	for i := range ptypes {
		iss, e := ProcTypeInstances(s, ptypes[i])
		if e != nil {
			return nil, e
		}
		instances = append(instances, iss...)
	}

	return
}

// ProcTypeInstances returns an array of all registered instances belonging
// to the given ProcType.
func ProcTypeInstances(s Snapshot, ptype *ProcType) (instances []*Instance, err error) {
	path := ptype.Path() + "/instances"
	names, err := s.conn.Getdir(path, s.Rev)
	if err != nil {
		return
	}

	instances = make([]*Instance, len(names))

	for i := range names {
		vals, e := s.conn.GetMulti(path+"/"+names[i], nil, s.Rev)

		if e != nil {
			return nil, e
		}

		state, e := strconv.Atoi(string(vals["state"]))
		if e != nil {
			return nil, e
		}

		addr := string(vals["host"]) + ":" + string(vals["port"])

		instances[i], err = NewInstance(ptype, addr, State(state), s)
		if err != nil {
			return
		}
	}

	return
}

// HostInstances returns an array of all registered instances belonging
// to the given host.
func HostInstances(s Snapshot, addr string) ([]Instance, error) {
	return nil, nil
}
