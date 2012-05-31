// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"path"
	"time"
)

// ProcType represents a process type with a certain scale.
type ProcType struct {
	Snapshot
	Name        ProcessName
	Revision    *Revision
	Heartbeat   *Heartbeat
	MaxRestarts int
}

type Heartbeat struct {
	Interval     int
	Treshold     int
	InitialDelay int
}

const PROCS_PATH = "procs"

var (
	HEARTBEAT_INTERVAL      = 30
	HEARTBEAT_TRESHOLD      = 2
	HEARTBEAT_INITIAL_DELAY = 1
)

var DEFAULT_HEARTBEAT = &Heartbeat{
	Interval:     HEARTBEAT_INTERVAL,
	Treshold:     HEARTBEAT_TRESHOLD,
	InitialDelay: HEARTBEAT_INITIAL_DELAY,
}

func NewProcType(revision *Revision, name ProcessName, s Snapshot) (*ProcType, error) {
	return &ProcType{Name: name, Revision: revision, Snapshot: s, Heartbeat: DEFAULT_HEARTBEAT}, nil
}

func (p *ProcType) createSnapshot(rev int64) Snapshotable {
	return &ProcType{Name: p.Name, Revision: p.Revision, Snapshot: Snapshot{rev, p.conn}}
}

// FastForward advances the instance in time. It returns
// a new instance of Instance with the supplied revision.
func (p *ProcType) FastForward(rev int64) *ProcType {
	return p.Snapshot.fastForward(p, rev).(*ProcType)
}

// Register registers a proctype with the registry.
func (p *ProcType) Register() (ptype *ProcType, err error) {
	exists, _, err := p.conn.Exists(p.Path(), &p.Rev)
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	if p.Heartbeat == nil {
		p.Heartbeat = DEFAULT_HEARTBEAT
	}

	attrs := &File{p.Snapshot, p.Path() + "/attrs", map[string]int{
		"heartbeat-interval":      p.Heartbeat.Interval,
		"heartbeat-treshold":      p.Heartbeat.Treshold,
		"heartbeat-initial-delay": p.Heartbeat.InitialDelay,
		"max-restarts":            p.MaxRestarts,
	}, new(JSONCodec)}

	attrs, err = attrs.Create()
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
	return path.Join(p.Revision.Path(), PROCS_PATH, string(p.Name))
}

func (p *ProcType) InstancePath(Id string) string {
	return path.Join(p.InstancesPath(), Id)
}

func (p *ProcType) InstancesPath() string {
	return path.Join(p.Revision.Path(), PROCS_PATH, string(p.Name), INSTANCES_PATH)
}

// ProcTypes returns an array of all registered proctypes belonging to the specified revision.
func RevisionProcTypes(s Snapshot, revision *Revision) (ptypes []*ProcType, err error) {
	path := revision.Path() + "/procs"
	names, err := s.conn.Getdir(path, s.Rev)
	if err != nil {
		return
	}

	ptypes = make([]*ProcType, len(names))

	for i := range names {
		name := ProcessName(names[i])

		ptypes[i] = &ProcType{Name: name, Revision: revision, Snapshot: s}
	}
	return
}

// ProcTypes returns an array of all registered proctypes.
func ProcTypes(s Snapshot) (ptypes []*ProcType, err error) {
	revs, err := Revisions(s)
	if err != nil {
		return
	}

	ptypes = []*ProcType{}

	for i := range revs {
		revps, e := RevisionProcTypes(s, revs[i])
		if e != nil {
			return nil, e
		}
		ptypes = append(ptypes, revps...)
	}
	return
}

// GetProcType fetches a ProcType from the coordinator
func GetProcType(s Snapshot, revision *Revision, name ProcessName) (p *ProcType, err error) {
	path := path.Join(revision.Path(), PROCS_PATH, string(name))

	f, err := Get(s, path+"/attrs", new(JSONCodec))
	if err != nil {
		return
	}
	value := f.Value.(map[string]interface{})

	p = &ProcType{
		Name:        name,
		Snapshot:    s,
		Revision:    revision,
		MaxRestarts: int(value["max-restarts"].(float64)),
		Heartbeat: &Heartbeat{
			Interval:     int(value["heartbeat-interval"].(float64)),
			Treshold:     int(value["heartbeat-treshold"].(float64)),
			InitialDelay: int(value["heartbeat-initial-delay"].(float64)),
		},
	}
	return
}

func (p *ProcType) String() string {
	return fmt.Sprintf("ProcType<%s:%s:%s>", p.Revision.App.Name, p.Revision.Ref, p.Name)
}

func (p *ProcType) Inspect() string {
	return fmt.Sprintf("%#v", p)
}
