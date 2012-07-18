// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"fmt"
	"time"
)

// ProcType represents a process type with a certain scale.
type ProcType struct {
	Path
	Name ProcessName
	App  *App
	Port int
}

const PROCS_PATH = "procs"

func NewProcType(app *App, name ProcessName, s Snapshot) *ProcType {
	return &ProcType{
		Name: name,
		App:  app,
		Path: Path{
			s, app.Path.Prefix(PROCS_PATH, string(name)),
		},
	}
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
	exists, _, err := p.conn.Exists(p.Path.Dir)
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

	port := &File{p.Snapshot, -1, p.Path.Prefix("port"), p.Port, new(IntCodec)}

	port, err = port.Create()
	if err != nil {
		return p, err
	}

	rev, err := p.Set("registered", time.Now().UTC().String())

	if err != nil {
		return p, err
	}
	ptype = p.FastForward(rev)

	return
}

// Unregister unregisters a proctype from the registry.
func (p *ProcType) Unregister() (err error) {
	return p.Del("/")
}

func (p *ProcType) InstancePath(id string) string {
	return p.Path.Prefix(INSTANCES_PATH, id)
}

func (p *ProcType) InstancesPath() string {
	return p.Path.Prefix(INSTANCES_PATH)
}

func (p *ProcType) GetInstanceNames() (ins []string, err error) {
	exists, _, err := p.conn.Exists(p.InstancesPath())
	if err != nil || !exists {
		return
	}

	ins, err = p.FastForward(-1).Getdir(p.InstancesPath())
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
	path := app.Path.Prefix(PROCS_PATH, string(name))

	port, err := Get(s, path+"/port", new(IntCodec))
	if err != nil {
		return
	}
	p = NewProcType(app, name, s)
	p.Port = port.Value.(int)

	return
}

func (p *ProcType) String() string {
	return fmt.Sprintf("ProcType<%s:%s>", p.App.Name, p.Name)
}

func (p *ProcType) Inspect() string {
	return fmt.Sprintf("%#v", p)
}
