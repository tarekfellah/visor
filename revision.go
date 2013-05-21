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
	dir        *cp.Dir
	App        *App
	Ref        string
	ArchiveUrl string
}

const revsPath = "revs"

// NewRevision returns a new instance of Revision.
func (s *Store) NewRevision(app *App, ref string) (rev *Revision) {
	rev = &Revision{App: app, Ref: ref}
	rev.dir = cp.NewDir(app.dir.Prefix(revsPath, ref), s.GetSnapshot())

	return
}

func (r *Revision) GetSnapshot() cp.Snapshot {
	return r.dir.Snapshot
}

// Register registers a new Revision with the registry.
func (r *Revision) Register() (*Revision, error) {
	sp, err := r.GetSnapshot().FastForward()
	if err != nil {
		return nil, err
	}

	exists, _, err := sp.Exists(r.dir.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrConflict
	}

	d, err := r.dir.Join(sp).Set("archive-url", r.ArchiveUrl)
	if err != nil {
		return nil, err
	}
	d, err = r.dir.Set("registered", timestamp())
	if err != nil {
		return nil, err
	}

	r.dir = d

	return r, nil
}

// Unregister unregisters a revision from the registry.
func (r *Revision) Unregister() error {
	sp, err := r.GetSnapshot().FastForward()
	if err != nil {
		return err
	}
	return r.dir.Join(sp).Del("/")
}

func (r *Revision) String() string {
	return fmt.Sprintf("Revision<%s:%s>", r.App.Name, r.Ref)
}

func (a *App) GetRevision(ref string) (*Revision, error) {
	sp, err := a.GetSnapshot().FastForward()
	if err != nil {
		return nil, err
	}
	return getRevision(a, ref, sp)

}

// Revisions returns an array of all registered revisions.
func (s *Store) GetRevisions() (revisions []*Revision, err error) {
	apps, err := s.GetApps()
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

func getRevision(app *App, ref string, s cp.Snapshotable) (*Revision, error) {
	sp := s.GetSnapshot()
	path := app.dir.Prefix(revsPath, ref)
	codec := new(cp.StringCodec)

	f, err := sp.GetFile(path+"/archive-url", codec)
	if err != nil {
		if cp.IsErrNoEnt(err) {
			err = errorf(ErrNotFound, "archive-url not found for %s:%s", app.Name, ref)
		}
		return nil, err
	}

	r := &Revision{
		dir:        cp.NewDir(path, sp),
		App:        app,
		Ref:        ref,
		ArchiveUrl: f.Value.(string),
	}

	return r, nil
}
