package visor

import (
	"fmt"
	"path"
	"time"
)

// ProcType represents a process type with a certain scale.
type ProcType struct {
	Snapshot
	Name     ProcessName
	Revision *Revision
}

const PROCS_PATH = "procs"

func NewProcType(revision *Revision, name ProcessName, s Snapshot) (*ProcType, error) {
	return &ProcType{Name: name, Revision: revision, Snapshot: s}, nil
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

	rev, err := p.conn.SetMulti(p.Path(), map[string][]byte{
		"registered": []byte(time.Now().UTC().String()),
		"scale":      []byte("0")}, p.Rev)

	if err != nil {
		return p, err
	}
	p = p.FastForward(rev)

	return
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

func (p *ProcType) String() string {
	return fmt.Sprintf("%#v", p)
}
