// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
)

func revSetup() (s Snapshot, app *App) {
	s, err := Dial(DefaultAddr, "/revision-test")
	if err != nil {
		panic(err)
	}

	r, _ := s.conn.Rev()
	s.conn.Del("/", r)
	s = s.FastForward(-1)

	rev, err := Init(s)
	if err != nil {
		panic(err)
	}
	s = s.FastForward(rev)

	app = NewApp("rev-test", "git://rev.git", "references", s)

	return
}

func TestRevisionRegister(t *testing.T) {
	s, app := revSetup()
	rev := NewRevision(app, "stable", app.Snapshot)

	check, _, err := s.conn.Exists(rev.dir.Name)
	if err != nil {
		t.Error(err)
		return
	}
	if check {
		t.Error("Revision already registered")
		return
	}

	rev, err = rev.Register()
	if err != nil {
		t.Error(err)
		return
	}

	check, _, err = s.conn.Exists(rev.dir.Name)
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Error("Revision registration failed")
	}

	_, err = rev.Register()
	if err == nil {
		t.Error("Revision allowed to be registered twice")
	}
}

func TestRevisionUnregister(t *testing.T) {
	s, app := revSetup()
	rev := NewRevision(app, "master", app.Snapshot)

	rev, err := rev.Register()
	if err != nil {
		t.Error(err)
	}

	err = rev.Unregister()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.conn.Exists(rev.dir.Name)
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("Revision still registered")
	}
}
