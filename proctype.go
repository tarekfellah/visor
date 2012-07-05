// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"fmt"
	"path"
	"time"
)

// ProcType represents a process type with a certain scale.
type ProcType struct {
	Snapshot
	Name        ProcessName
	App         *App
	Port        int
}

const PROCS_PATH = "procs"

func NewProcType(app *App, name ProcessName, s Snapshot) *ProcType {
	return &ProcType{Name: name, App: app, Snapshot: s}
}

func (p *ProcType) createSnapshot(rev int64) Snapshotable {
	tmp := *p
	tmp.Snapshot = Snapshot{rev, p.conn}
	return &tmp
}

// FastForward advances the instance in time. It returns
// a new instance of Instance with the supplied revision.
func (p *ProcType) FastForward(rev int64) *ProcType {
	return p.Snapshot.fastForward(p, rev).(*ProcType)
}

// Register registers a proctype with the registry.
func (p *ProcType) Register() (ptype *ProcType, err error) {
	exists, _, err := p.conn.Exists(p.Path())
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	p.Port, err = ClaimNextPort(p.Snapshot)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't claim port: %s", err.Error()))
	}

	port := &File{p.Snapshot, p.Path() + "/port", p.Port, new(IntCodec)}

	port, err = port.Create()
	if err != nil {
		return p, err
	}

	rev, err := p.conn.Set(p.Path()+"/registered", p.Rev, []byte(time.Now().UTC().String()))

	if err != nil {
		return p, err
	}
	ptype = p.FastForward(rev)

	return
}

// Unregister unregisters a proctype from the registry.
func (p *ProcType) Unregister() (err error) {
	return p.conn.Del(p.Path(), p.Rev)
}

func (p *ProcType) Path() string {
	return path.Join(p.App.Path(), PROCS_PATH, string(p.Name))
}

func (p *ProcType) InstancePath(Id string) string {
	return path.Join(p.InstancesPath(), Id)
}

func (p *ProcType) InstancesPath() string {
	return path.Join(p.Path(), INSTANCES_PATH)
}

func (p *ProcType) GetInstanceNames() (ins []string, err error) {
	exists, _, err := p.conn.Exists(p.InstancesPath())
	if err != nil || !exists {
		return
	}

	ins, err = p.conn.Getdir(p.InstancesPath(), p.Snapshot.FastForward(-1).Rev)
	if err != nil {
		return
	}

	return
}

func (p *ProcType) GetInstanceInfos() (ins []*InstanceInfo, err error) {
	insNames, err := p.GetInstanceNames()
	if err != nil {
		return
	}

	for _, insName := range insNames {
		var i *InstanceInfo

		i, err = GetInstanceInfo(p.Snapshot, insName)
		if err != nil {
			return
		}

		ins = append(ins, i)
	}

	return
}

// GetProcType fetches a ProcType from the coordinator
func GetProcType(s Snapshot, app *App, name ProcessName) (p *ProcType, err error) {
	path := path.Join(app.Path(), PROCS_PATH, string(name))

	port, err := Get(s, path+"/port", new(IntCodec))
	if err != nil {
		return
	}

	p = &ProcType{
		Name:     name,
		Snapshot: s,
		App:      app,
		Port:     port.Value.(int),
	}
	return
}

func (p *ProcType) String() string {
	return fmt.Sprintf("ProcType<%s:%s>", p.App.Name, p.Name)
}

func (p *ProcType) Inspect() string {
	return fmt.Sprintf("%#v", p)
}
