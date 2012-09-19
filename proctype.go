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
	dir
	Name string
	App  *App
	Port int
}

const procsPath = "procs"

func NewProcType(app *App, name string, s Snapshot) *ProcType {
	return &ProcType{
		Name: name,
		App:  app,
		dir: dir{
			s, app.dir.prefix(procsPath, string(name)),
		},
	}
}

func (p *ProcType) createSnapshot(rev int64) snapshotable {
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
	exists, _, err := p.conn.Exists(p.dir.Name)
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

	port := &file{p.Snapshot, -1, p.dir.prefix("port"), p.Port, new(intCodec)}

	port, err = port.Create()
	if err != nil {
		return p, err
	}

	rev, err := p.set("registered", time.Now().UTC().String())

	if err != nil {
		return p, err
	}
	ptype = p.FastForward(rev)

	return
}

// Unregister unregisters a proctype from the registry.
func (p *ProcType) Unregister() (err error) {
	return p.del("/")
}

func (p *ProcType) instancesPath() string {
	return p.dir.prefix(instancesPath)
}

func (p *ProcType) GetInstanceNames() (ins []string, err error) {
	ins, err = p.getdir(p.instancesPath())
    if IsErrNoEnt(err) {
        err = nil
    }
	return
}

func (p *ProcType) GetInstances() (ins []*Instance, err error) {
	names, err := p.GetInstanceNames()
	if err != nil || len(names) == 0 {
		return
	}

	insch := make(chan *Instance, 1)
	errch := make(chan error, 1)

	for _, name := range names {
		go func(name string) {
			i, err := GetInstance(p.Snapshot, name)
			if err != nil {
				errch <- err
			} else {
				insch <- i
			}
		}(name)
	}
	for {
		select {
		case i := <-insch:
			ins = append(ins, i)

			if len(ins) == len(names) {
				return
			}
		case err = <-errch:
			return
		}
	}
	return
}

// GetProcType fetches a ProcType from the coordinator
func GetProcType(s Snapshot, app *App, name string) (p *ProcType, err error) {
	path := app.dir.prefix(procsPath, string(name))

	port, err := s.getFile(path+"/port", new(intCodec))
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
