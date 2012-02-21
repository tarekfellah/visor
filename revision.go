package visor

import (
	"fmt"
	"time"
)

// A Revision represents an application revision,
// identifiable by its `ref`.
type Revision struct {
	Snapshot
	App *App
	ref string
}

// NewRevision returns a new instance of Revision.
func NewRevision(app *App, ref string, snapshot Snapshot) (rev *Revision, err error) {
	rev = &Revision{App: app, ref: ref, Snapshot: snapshot}
	return
}

func (r *Revision) createSnapshot(rev int64) Snapshotable {
	return &Revision{App: r.App, ref: r.ref, Snapshot: Snapshot{rev, r.conn}}
}

// FastForward advances the revision in time. It returns
// a new instance of Revision with the supplied revision.
func (r *Revision) FastForward(rev int64) *Revision {
	return r.Snapshot.fastForward(r, rev).(*Revision)
}

// Register registers a new Revision with the registry.
func (r *Revision) Register() (revision *Revision, err error) {
	exists, _, err := r.conn.Exists(r.Path(), &r.Rev)
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	_, err = r.conn.Set(r.Path()+"/registered", r.Rev, []byte(time.Now().UTC().String()))
	if err != nil {
		return
	}
	rev, err := r.conn.Set(r.Path()+"/scale", r.Rev, []byte("0"))

	revision = r.FastForward(rev)

	return
}

// Unregister unregisters a revision from the registry.
func (r *Revision) Unregister() (err error) {
	return r.conn.Del(r.Path(), r.Rev)
}
func (r *Revision) Scale(proctype string, factor int) error {
	return nil
}
func (r *Revision) Instances(proctype ProcessName) ([]Instance, error) {
	return nil, nil
}
func (r *Revision) RegisterInstance(proctype ProcessName, address string) (*Instance, error) {
	return nil, nil
}
func (r *Revision) UnregisterInstance(instance *Instance) error {
	return nil
}

// Path returns this.Revision's directory path in the registry.
func (r *Revision) Path() string {
	return r.App.Path() + "/revs/" + r.ref
}

func (r *Revision) String() string {
	return fmt.Sprintf("%#v", r)
}

// Revisions returns an array of all registered revisions.
func Revisions(s Snapshot) (revisions []*Revision, err error) {
	apps, err := Apps(s)
	if err != nil {
		return
	}

	revisions = []*Revision{}

	for i := range apps {
		revs, e := AppRevisions(s, apps[i])
		if e != nil {
			return nil, e
		}
		revisions = append(revisions, revs...)
	}

	return
}

// AppRevisions returns an array of all registered revisions belonging
// to the given application.
func AppRevisions(s Snapshot, app *App) (revisions []*Revision, err error) {
	refs, err := s.conn.Getdir(app.Path()+"/revs", s.Rev)
	if err != nil {
		return
	}
	revisions = make([]*Revision, len(refs))

	for i := range refs {
		r, e := NewRevision(app, refs[i], s)
		if e != nil {
			return nil, e
		}

		revisions[i] = r
	}

	return
}
