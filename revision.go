// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"time"
)

// A Revision represents an application revision,
// identifiable by its `ref`.
type Revision struct {
	dir
	App        *App
	Ref        string
	ArchiveUrl string
}

const REVS_PATH = "revs"

// NewRevision returns a new instance of Revision.
func NewRevision(app *App, ref string, snapshot Snapshot) (rev *Revision) {
	rev = &Revision{App: app, Ref: ref}
	rev.dir = dir{snapshot, app.dir.prefix(REVS_PATH, ref)}

	return
}

func (r *Revision) createSnapshot(rev int64) snapshotable {
	tmp := *r
	tmp.Snapshot = Snapshot{rev, r.conn}
	return &tmp
}

// FastForward advances the revision in time. It returns
// a new instance of Revision with the supplied revision.
func (r *Revision) FastForward(rev int64) *Revision {
	return r.Snapshot.fastForward(r, rev).(*Revision)
}

// Register registers a new Revision with the registry.
func (r *Revision) Register() (revision *Revision, err error) {
	exists, _, err := r.conn.Exists(r.dir.Name)
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	rev, err := r.set("archive-url", r.ArchiveUrl)
	if err != nil {
		return
	}
	rev, err = r.set("registered", time.Now().UTC().String())
	if err != nil {
		return
	}

	revision = r.FastForward(rev)

	return
}

// Unregister unregisters a revision from the registry.
func (r *Revision) Unregister() (err error) {
	return r.del("/")
}

func (r *Revision) SetArchiveUrl(url string) (revision *Revision, err error) {
	rev, err := r.set("archive-url", url)
	if err != nil {
		return
	}
	revision = r.FastForward(rev)
	return
}

func (r *Revision) String() string {
	return fmt.Sprintf("Revision<%s:%s>", r.App.Name, r.Ref)
}

func (r *Revision) Inspect() string {
	return fmt.Sprintf("%#v", r)
}

func GetRevision(s Snapshot, app *App, ref string) (r *Revision, err error) {
	path := app.dir.prefix(REVS_PATH, ref)
	codec := new(StringCodec)

	f, err := s.getFile(path+"/archive-url", codec)
	if err != nil {
		return
	}

	r = &Revision{
		dir:        dir{s, path},
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
	refs, err := s.getdir(app.dir.prefix("revs"))
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
