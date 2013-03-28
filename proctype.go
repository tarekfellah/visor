// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"fmt"
	cp "github.com/soundcloud/cotterpin"
	"regexp"
	"strconv"
	"time"
)

var reProcName = regexp.MustCompile("^[[:alnum:]]+$")

// ProcType represents a process type with a certain scale.
type ProcType struct {
	Dir   cp.Dir
	Name  string
	App   *App
	Port  int
	Attrs ProcTypeAttrs
}

// Mutable extra ProcType attributes.
type ProcTypeAttrs struct {
	Limits ResourceLimits `json:"limits"`
}

// Per-proctype resource limits.
type ResourceLimits struct {
	// Maximum memory allowance in MB for an instance of this ProcType.
	MemoryLimitMb *int `json:"memory-limit-mb,omitempty"`
}

const (
	procsPath      = "procs"
	procsPortPath  = "port"
	procsAttrsPath = "attrs"
)

func (s *Store) NewProcType(app *App, name string) *ProcType {
	return &ProcType{
		Name: name,
		App:  app,
		Dir: cp.Dir{
			s.GetSnapshot(), app.Dir.Prefix(procsPath, string(name)),
		},
	}
}

func (p *ProcType) GetSnapshot() cp.Snapshot {
	return p.Dir.Snapshot
}

// Join advances the instance in time. It returns
// a new instance of ProcType at rev of the supplied
// cp.Snapshotable.
func (p *ProcType) Join(s cp.Snapshotable) *ProcType {
	tmp := *p
	tmp.Dir.Snapshot = s.GetSnapshot()
	return &tmp
}

// Register registers a proctype with the registry.
func (p *ProcType) Register() (pty *ProcType, err error) {
	// Explicit FastForward to assure existence
	// check against latest state
	s, err := p.Dir.Snapshot.FastForward()
	if err != nil {
		return nil, err
	}
	p = p.Join(s)

	exists, _, err := p.Dir.Snapshot.Exists(p.Dir.Name)
	if err != nil {
		return
	}
	if exists {
		return nil, cp.ErrKeyConflict
	}

	if !reProcName.MatchString(p.Name) {
		return nil, ErrBadPtyName
	}

	p.Port, err = claimNextPort(p.Dir.Snapshot)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't claim port: %s", err.Error()))
	}

	port := &cp.File{p.Dir.Snapshot, p.Dir.Prefix("port"), p.Port, new(cp.IntCodec)}

	port, err = port.Save()
	if err != nil {
		return p, err
	}

	d, err := p.Dir.Set("registered", timestamp())
	if err != nil {
		return p, err
	}
	pty = p.Join(d)

	return
}

// Unregister unregisters a proctype from the registry.
func (p *ProcType) Unregister() (err error) {
	return p.Dir.Del("/")
}

func (p *ProcType) instancesPath() string {
	return p.Dir.Prefix(instancesPath)
}

func (p *ProcType) failedInstancesPath() string {
	return p.Dir.Prefix(failedPath)
}

func (p *ProcType) NumInstances() (int, error) {
	revs, err := p.Dir.Snapshot.Getdir(p.Dir.Prefix("instances"))
	if err != nil {
		return -1, err
	}
	total := 0

	for _, rev := range revs {
		// TODO: remove this once all old instances are cleaned up
		if len(rev) != 7 {
			continue
		}
		size, _, err := p.Dir.Snapshot.Stat(p.Dir.Prefix("instances", rev), &p.Dir.Snapshot.Rev)
		if err != nil {
			return -1, err
		}
		total += size
	}
	return total, nil
}

func (p *ProcType) InstanceIds() (ids []string, err error) {
	revs, err := p.Dir.Snapshot.Getdir(p.Dir.Prefix("instances"))
	if err != nil {
		return
	}
	for _, rev := range revs {
		// TODO: remove this once all old instances are cleaned up
		if len(rev) != 7 {
			continue
		}

		iids, e := p.Dir.Snapshot.Getdir(p.Dir.Prefix("instances", rev))
		if e != nil {
			return nil, e
		}
		ids = append(ids, iids...)
	}
	return
}

func (p *ProcType) GetFailedInstances() (ins []*Instance, err error) {
	ids, err := p.Dir.Snapshot.Getdir(p.failedInstancesPath())
	if err != nil {
		return
	}
	return getInstances(storeFromSnapshotable(p), ids)
}

func (p *ProcType) GetInstances() (ins []*Instance, err error) {
	ids, err := p.InstanceIds()
	if err != nil {
		return
	}
	return getInstances(storeFromSnapshotable(p), ids)
}

func (p *ProcType) StoreAttrs() (ptype *ProcType, err error) {
	attrs := cp.NewFile(p.Dir.Prefix(procsAttrsPath), p.Attrs, new(cp.JsonCodec), p.Dir.Snapshot)
	attrs, err = attrs.Save()
	if err != nil {
		return
	}

	ptype = p.Join(attrs)
	return
}

func (p *ProcType) String() string {
	return fmt.Sprintf("ProcType<%s:%s>", p.App.Name, p.Name)
}

func (p *ProcType) Inspect() string {
	return fmt.Sprintf("%#v", p)
}

// TODO consider moving to (*App).GetProcType(name)

// GetProcType fetches a ProcType from the coordinator
func (s *Store) GetProcType(app *App, name string) (p *ProcType, err error) {
	path := app.Dir.Prefix(procsPath, string(name))

	dir := cp.Dir{
		Snapshot: s.GetSnapshot(),
		Name:     path,
	}

	port, err := dir.GetFile(procsPortPath, new(cp.IntCodec))
	if err != nil {
		return nil, errorf(ErrNotFound, "proctype %s not found for %s", name, app.Name)
	}
	p = s.NewProcType(app, name)
	p.Port = port.Value.(int)

	_, err = s.GetSnapshot().GetFile(dir.Prefix(procsAttrsPath), &cp.JsonCodec{DecodedVal: &p.Attrs})
	if cp.IsErrNoEnt(err) {
		err = nil
	}

	return
}

func getInstances(s *Store, ids []string) (ins []*Instance, err error) {
	ch, errch := cp.GetSnapshotables(ids, func(idstr string) (cp.Snapshotable, error) {
		id, err := strconv.ParseInt(idstr, 10, 64)
		if err != nil {
			return nil, err
		}
		return s.GetInstance(id)
	})
	for i := 0; i < len(ids); i++ {
		select {
		case r := <-ch:
			ins = append(ins, r.(*Instance))
		case err := <-errch:
			return nil, err
		}
	}
	return
}

func claimNextPort(s cp.Snapshot) (int, error) {
	for {
		var err error
		s, err = s.FastForward()
		if err != nil {
			return -1, err
		}

		f, err := s.GetFile(nextPortPath, new(cp.IntCodec))
		if err == nil {
			port := f.Value.(int)

			f, err = f.Set(port + 1)
			if err == nil {
				return port, nil
			} else {
				time.Sleep(time.Second / 10)
			}
		} else {
			return -1, err
		}
	}

	return -1, nil
}
