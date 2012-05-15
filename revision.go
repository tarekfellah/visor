package visor

import (
	"fmt"
	"path"
	"strconv"
	"time"
)

// A Revision represents an application revision,
// identifiable by its `ref`.
type Revision struct {
	Snapshot
	App        *App
	Ref        string
	ArchiveUrl string
}

const REVS_PATH = "revs"
const SCALE_PATH string = "scale"

// NewRevision returns a new instance of Revision.
func NewRevision(app *App, ref string, snapshot Snapshot) (rev *Revision, err error) {
	rev = &Revision{App: app, Ref: ref, Snapshot: snapshot}
	return
}

func (r *Revision) createSnapshot(rev int64) Snapshotable {
	return &Revision{App: r.App, Ref: r.Ref, Snapshot: Snapshot{rev, r.conn}}
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
	rev, err := r.conn.Set(r.Path()+"/archive-url", r.Rev, []byte(r.ArchiveUrl))
	if err != nil {
		return
	}

	revision = r.FastForward(rev)

	return
}

// Unregister unregisters a revision from the registry.
func (r *Revision) Unregister() (err error) {
	return r.conn.Del(r.Path(), r.Rev)
}

func (r *Revision) SetArchiveUrl(url string) (revision *Revision, err error) {
	rev, err := r.conn.Set(r.Path()+"/archive-url", r.Rev, []byte(url))
	if err != nil {
		return
	}
	revision = r.FastForward(rev)
	return
}

func (r *Revision) Scale(proctype string, factor int) (revision *Revision, err error) {
	op := OpStart
	p := path.Join(ProcPath(r.App.Name, r.Ref, proctype), SCALE_PATH)

	rev, err := r.conn.Set(p, r.Rev, []byte(strconv.Itoa(factor)))
	if err != nil {
		return
	}

	revision = r.FastForward(rev)

	res, _, err := r.conn.Get(p, &r.Rev)
	if err != nil {
		return
	}
	if res != nil {
		current, e := strconv.Atoi(string(res))
		if err != nil {
			return nil, e
		}

		if factor < current {
			op = OpStop
		}

		factor = factor - current

		if factor < 0 {
			factor = -factor
		}
	}

	for i := 0; i < factor; i++ {
		ticket, err := CreateTicket(r.App.Name, r.Ref, ProcessName(proctype), op, revision.Snapshot)
		if err != nil {
			return nil, err
		}

		revision = r.FastForward(ticket.Rev)
	}

	return
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
	return path.Join(r.App.Path(), REVS_PATH, r.Ref)
}

func (r *Revision) String() string {
	return fmt.Sprintf("%#v", r)
}

func GetRevision(s Snapshot, app *App, ref string) (r *Revision, err error) {
	path := app.Path() + "/revs/" + ref
	codec := new(StringCodec)

	f, err := Get(s, path+"/archive-url", codec)
	if err != nil {
		return
	}

	r = &Revision{
		Snapshot:   s,
		App:        app,
		Ref:        ref,
		ArchiveUrl: f.Value.(string),
	}
	return
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
		r, e := GetRevision(s, app, refs[i])
		if e != nil {
			return nil, e
		}

		revisions[i] = r
	}

	return
}
