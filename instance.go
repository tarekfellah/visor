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
	InsStateReady         = "ready"
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
	Snapshot
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
		Snapshot: snapshot}

	return
}

// FastForward advances the instance in time. It returns
// a new instance of Instance with the supplied revision.
func (i *Instance) FastForward(rev int64) *Instance {
	return i.Snapshot.fastForward(i, rev).(*Instance)
}

func (i *Instance) createSnapshot(rev int64) Snapshotable {
	return &Instance{Addr: i.Addr, State: i.State, ProcType: i.ProcType, Revision: i.Revision, Snapshot: Snapshot{rev, i.conn}}
}

// Register registers an instance with the registry.
func (i *Instance) Register() (instance *Instance, err error) {
	exists, _, err := i.conn.Exists(i.Path())
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	rev, err := i.conn.SetMulti(i.Path(), map[string][]byte{
		"host":     []byte(i.Addr.IP.String()),
		"port":     []byte(strconv.Itoa(i.Addr.Port)),
		"state":    []byte(i.State),
		"app":      []byte(i.ProcType.App.Name),
		"proctype": []byte(string(i.ProcType.Name)),
		"revision": []byte(i.Revision.Ref),
	}, i.Rev)
	if err != nil {
		return i, err
	}
	now := []byte(time.Now().UTC().String())

	rev, err = i.conn.Set(i.Path()+"/registered", i.Rev, now)
	if err != nil {
		return i, err
	}
	rev, err = i.conn.Set(i.ProcType.InstancePath(i.Id()), i.Rev, now)
	instance = i.FastForward(rev)

	return
}

// Unregister unregisters an instance with the registry.
func (i *Instance) Unregister() (err error) {
	rev := i.Rev

	err = i.conn.Del(i.ProcType.InstancePath(i.Id()), rev)
	if err != nil {
		return
	}
	err = i.conn.Del(i.Path(), rev)
	return
}

// UpdateState updates the instance's state file in
// the coordinator to the given value.
func (i *Instance) UpdateState(s State) (ins *Instance, err error) {
	newrev, err := i.conn.Set(i.Path()+"/state", i.Rev, []byte(s))
	if err != nil {
		return
	}
	ins = i.FastForward(newrev)
	ins.State = s

	return
}

// Path returns the instance's directory path in the registry.
func (i *Instance) Path() (path string) {
	return "/instances/" + i.Id()
}

func (i *Instance) Id() string {
	return strings.Replace(strings.Replace(i.Addr.String(), ".", "-", -1), ":", "-", -1)
}

func (i *Instance) String() string {
	return fmt.Sprintf("%#v", i)
}

// GetInstanceInfo returns an InstanceInfo from the given app, rev, proc and instance ids.
func GetInstanceInfo(s Snapshot, insName string) (ins *InstanceInfo, err error) {
	keys := []string{"app", "host", "port", "proctype", "revision", "state"}
	p := path.Join(INSTANCES_PATH, insName)

	vals, err := s.conn.GetMulti(p, keys, s.FastForward(-1).Rev)
	if err != nil {
		return
	}

	port, err := strconv.Atoi(string(vals["port"]))
	if err != nil {
		return
	}

	ins = &InstanceInfo{
		Name:         insName,
		AppName:      string(vals["app"]),
		Host:         string(vals["host"]),
		Port:         port,
		ProcessName:  ProcessName(vals["proctype"]),
		RevisionName: string(vals["revision"]),
		ServiceName:  string(vals["app"]) + "-" + string(vals["proctype"]),
		State:        State(string(vals["state"])),
	}

	return
}
