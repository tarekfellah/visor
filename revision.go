// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
)

// A Revision represents an application revision,
// identifiable by its `ref`.
type Revision struct {
	Dir        dir
	App        *App
	Ref        string
	ArchiveUrl string
}

const revsPath = "revs"

// NewRevision returns a new instance of Revision.
func NewRevision(app *App, ref string, snapshot Snapshot) (rev *Revision) {
	rev = &Revision{App: app, Ref: ref}
	rev.Dir = dir{snapshot, app.Dir.prefix(revsPath, ref)}

	return
}

func (r *Revision) createSnapshot(rev int64) snapshotable {
	tmp := *r
	tmp.Dir.Snapshot = Snapshot{rev, r.Dir.Snapshot.conn}
	return &tmp
}

// FastForward advances the revision in time. It returns
// a new instance of Revision with the supplied revision.
func (r *Revision) FastForward(rev int64) *Revision {
	return r.Dir.Snapshot.fastForward(r, rev).(*Revision)
}

// Register registers a new Revision with the registry.
func (r *Revision) Register() (revision *Revision, err error) {
	exists, _, err := r.Dir.Snapshot.conn.Exists(r.Dir.Name)
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	rev, err := r.Dir.set("archive-url", r.ArchiveUrl)
	if err != nil {
		return
	}
	rev, err = r.Dir.set("registered", timestamp())
	if err != nil {
		return
	}

	revision = r.FastForward(rev)

	return
}

// Unregister unregisters a revision from the registry.
func (r *Revision) Unregister() (err error) {
	return r.Dir.del("/")
}

func (r *Revision) SetArchiveUrl(url string) (revision *Revision, err error) {
	rev, err := r.Dir.set("archive-url", url)
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
	path := app.Dir.prefix(revsPath, ref)
	codec := new(stringCodec)

	f, err := s.getFile(path+"/archive-url", codec)
	if err != nil {
		return
	}

	r = &Revision{
		Dir:        dir{s, path},
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
		revs, e := apps[i].GetRevisions()
		if e != nil {
			return nil, e
		}
		revisions = append(revisions, revs...)
	}

	return
}
