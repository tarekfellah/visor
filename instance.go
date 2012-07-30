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

// An Instance represents a running process of a specific type.
type Instance struct {
	Path
	Name         string
	AppName      string
	RevisionName string
	ProcessName  ProcessName
	ServiceName  string
	Host         string
	Port         int
	State        State
}

// NewInstance creates and returns a new Instance object.
func NewInstance(pty string, rev string, app string, addr string, snapshot Snapshot) (ins *Instance, err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	ins = &Instance{
		Host:         tcpAddr.IP.String(),
		Port:         tcpAddr.Port,
		ServiceName:  app + "-" + pty,
		AppName:      app,
		ProcessName:  ProcessName(pty),
		RevisionName: rev,
		State:        InsStateInitial,
	}
	ins.Path = Path{snapshot, "/instances/" + ins.Id()}
	ins.Name = ins.Id()

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

func (i *Instance) ProctypePath() string {
	return path.Join(APPS_PATH, i.AppName, PROCS_PATH, string(i.ProcessName), INSTANCES_PATH, i.Id())
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

	_, err = i.Set("info", i.String())
	if err != nil {
		return i, err
	}
	_, err = i.Set("state", string(i.State))
	if err != nil {
		return i, err
	}
	now := time.Now().UTC().String()

	s, err := i.Snapshot.Set(i.ProctypePath(), now)
	instance = i.FastForward(s.Rev)

	return
}

// Unregister unregisters an instance with the registry.
func (i *Instance) Unregister() (err error) {
	err = i.Snapshot.Del(i.ProctypePath())
	if err != nil {
		return
	}
	err = i.Del("/")
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
	return fmt.Sprintf("%s-%d", strings.Replace(i.Host, ".", "-", -1), i.Port)
}

func (i *Instance) String() string {
	return strings.Join([]string{
		i.AppName,
		i.RevisionName,
		string(i.ProcessName),
		i.Host,
		fmt.Sprintf("%d", i.Port),
	}, " ")
}

func (i *Instance) AddrString() string {
	return i.Host + ":" + strconv.Itoa(i.Port)
}

func (i *Instance) RefString() string {
	return fmt.Sprintf("%s:%s@%s", i.AppName, i.ProcessName, i.RevisionName)
}

func (i *Instance) LogString() string {
	return fmt.Sprintf("%s (%s)", i.RefString(), i.AddrString())
}

// GetInstance returns an Instance from the given app, rev, proc and instance ids.
func GetInstance(s Snapshot, insName string) (ins *Instance, err error) {
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

	addr := fields[3] + ":" + fields[4]

	ins, err = NewInstance(fields[2], fields[1], fields[0], addr, s)
	if err != nil {
		return
	}
	ins.State = State(state)

	return
}
