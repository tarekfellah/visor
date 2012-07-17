// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	InsStateInitial State = "initial"
	InsStateStarted       = "started"
	InsStateFailed        = "failed"
	InsStateDead          = "dead"
	InsStateExited        = "exited"
)

const INSTANCES_PATH = "instances"

// InstanceInfo represents instance information as ids,
// when it's impractical to use the full Instance type.
type InstanceInfo struct {
	Name         string
	AppName      string
	RevisionName string
	ProcessName  ProcessName
	ServiceName  string
	Host         string
	Port         int
	State        State
}

func (i InstanceInfo) AddrString() string {
	return i.Host + ":" + strconv.Itoa(i.Port)
}
func (i InstanceInfo) RevString() string {
	return i.AppName + "-" + i.RevisionName
}
func (i InstanceInfo) LogString() string {
	return fmt.Sprintf("%s (%s)", i.RevString(), i.AddrString())
}

// An Instance represents a running process of a specific type.
type Instance struct {
	Path
	ProcType *ProcType // ProcType the instance belongs to
	Revision *Revision
	Addr     *net.TCPAddr // TCP address of the running instance
	State    State        // Current state of the instance
}

// NewInstance creates and returns a new Instance object.
func NewInstance(pty *ProcType, rev *Revision, addr string, snapshot Snapshot) (ins *Instance, err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	ins = &Instance{
		Addr:     tcpAddr,
		ProcType: pty,
		Revision: rev,
		State:    InsStateInitial,
	}
	ins.Path = Path{snapshot, "/instances/" + ins.Id()}

	return
}

// FastForward advances the instance in time. It returns
// a new instance of Instance with the supplied revision.
func (i *Instance) FastForward(rev int64) *Instance {
	return i.Snapshot.fastForward(i, rev).(*Instance)
}

func (i *Instance) createSnapshot(rev int64) Snapshotable {
	tmp := *i
	tmp.Snapshot = Snapshot{rev, i.conn}
	return &tmp
}

// Register registers an instance with the registry.
func (i *Instance) Register() (instance *Instance, err error) {
	exists, _, err := i.conn.Exists(i.Path.Dir)
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	rev, err := i.Set("info", i.String())
	if err != nil {
		return i, err
	}
	rev, err = i.Set("state", string(i.State))
	if err != nil {
		return i, err
	}
	now := time.Now().UTC().String()

	rev, err = i.Snapshot.Set(i.ProcType.InstancePath(i.Id()), now)
	instance = i.FastForward(rev)

	return
}

// Unregister unregisters an instance with the registry.
func (i *Instance) Unregister() (err error) {
	err = i.Snapshot.Del(i.ProcType.InstancePath(i.Id()))
	if err != nil {
		return
	}
	err = i.Del("/")
	return
}

func (i *InstanceInfo) Unregister(s Snapshot) (err error) {
	p := path.Join(APPS_PATH, i.AppName, PROCS_PATH, string(i.ProcessName), INSTANCES_PATH, i.Name)

	err = s.Del(p)
	if err != nil {
		return
	}
	err = s.Del(path.Join(INSTANCES_PATH, i.Name))
	return
}

// UpdateState updates the instance's state file in
// the coordinator to the given value.
func (i *Instance) UpdateState(s State) (ins *Instance, err error) {
	newrev, err := i.Set("state", string(s))
	if err != nil {
		return
	}
	ins = i.FastForward(newrev)
	ins.State = s

	return
}

func (i *Instance) Id() string {
	return strings.Replace(strings.Replace(i.Addr.String(), ".", "-", -1), ":", "-", -1)
}

func (i *Instance) String() string {
	return strings.Join([]string{
		i.ProcType.App.Name,
		i.Revision.Ref,
		string(i.ProcType.Name),
		i.Addr.IP.String(),
		fmt.Sprintf("%d", i.Addr.Port),
	}, " ")
}

// GetInstanceInfo returns an InstanceInfo from the given app, rev, proc and instance ids.
func GetInstanceInfo(s Snapshot, insName string) (ins *InstanceInfo, err error) {
	p := path.Join(INSTANCES_PATH, insName)

	state, _, err := s.conn.Get(p+"/state", nil)
	if err != nil {
		return
	}

	info, _, err := s.conn.Get(p+"/info", nil)
	if err != nil {
		return
	}
	fields := strings.Fields(string(info))

	port, err := strconv.Atoi(string(fields[4]))
	if err != nil {
		return
	}

	ins = &InstanceInfo{
		Name:         insName,
		AppName:      fields[0],
		RevisionName: fields[1],
		ProcessName:  ProcessName(fields[2]),
		Host:         fields[3],
		Port:         port,
		ServiceName:  fields[0] + "-" + fields[2],
		State:        State(state),
	}

	return
}
