// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	cp "github.com/soundcloud/cotterpin"
)

// A Revision represents an application revision,
// identifiable by its `ref`.
type Revision struct {
	Dir        cp.Dir
	App        *App
	Ref        string
	ArchiveUrl string
}

const revsPath = "revs"

// NewRevision returns a new instance of Revision.
func (s *Store) NewRevision(app *App, ref string) (rev *Revision) {
	rev = &Revision{App: app, Ref: ref}
	rev.Dir = cp.Dir{s.GetSnapshot(), app.Dir.Prefix(revsPath, ref)}

	return
}

func (r *Revision) GetSnapshot() cp.Snapshot {
	return r.Dir.Snapshot
}

// Join advances the Revision in time. It returns a new
// instance of Revision at the rev of the supplied
// cp.Snapshotable.
func (r *Revision) Join(s cp.Snapshotable) *Revision {
	tmp := *r
	tmp.Dir.Snapshot = s.GetSnapshot()
	return &tmp
}

// Register registers a new Revision with the registry.
func (r *Revision) Register() (revision *Revision, err error) {
	// Explicit FastForward to assure existence
	// check against latest state
	s, err := r.Dir.Snapshot.FastForward()
	if err != nil {
		return nil, err
	}
	r = r.Join(s)

	exists, _, err := r.Dir.Snapshot.Exists(r.Dir.Name)
	if err != nil {
		return
	}
	if exists {
		return nil, cp.ErrKeyConflict
	}

	_, err = r.Dir.Set("archive-url", r.ArchiveUrl)
	if err != nil {
		return
	}
	d, err := r.Dir.Set("registered", timestamp())
	if err != nil {
		return
	}

	revision = r.Join(d)

	return
}

// Unregister unregisters a revision from the registry.
func (r *Revision) Unregister() (err error) {
	return r.Dir.Del("/")
}

func (r *Revision) SetArchiveUrl(url string) (revision *Revision, err error) {
	d, err := r.Dir.Set("archive-url", url)
	if err != nil {
		return
	}
	revision = r.Join(d)
	return
}

func (r *Revision) String() string {
	return fmt.Sprintf("Revision<%s:%s>", r.App.Name, r.Ref)
}

func (r *Revision) Inspect() string {
	return fmt.Sprintf("%#v", r)
}

func (s *Store) GetRevision(app *App, ref string) (*Revision, error) {
  sp, err := s.GetSnapshot().FastForward()
  if err != nil {
    return nil, err
  }

	path := app.Dir.Prefix(revsPath, ref)
	codec := new(cp.StringCodec)

	f, err := sp.GetFile(path+"/archive-url", codec)
	if err != nil {
	  if cp.IsErrNoEnt(err) {
	    err = errorf(ErrNotFound, "archive-url not found for %s:%s", app.Name, ref)
    }
		return nil, err
	}

  r := &Revision{
		Dir:        cp.Dir{sp, path},
		App:        app,
		Ref:        ref,
		ArchiveUrl: f.Value.(string),
	}

	return r, err
}

// Revisions returns an array of all registered revisions.
func (s *Store) Revisions() (revisions []*Revision, err error) {
	apps, err := s.Apps()
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
