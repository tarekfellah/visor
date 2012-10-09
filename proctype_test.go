// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
)

func proctypeSetup(ref string) (s Snapshot, app *App) {
	s, err := Dial(DefaultAddr, "/proctype-test")
	if err != nil {
		panic(err)
	}

	r, _ := s.conn.Rev()
	s.conn.Del("/", r)
	s = s.FastForward(-1)

	r, err = Init(s)
	if err != nil {
		panic(err)
	}
	s = s.FastForward(r)

	app = NewApp("rev-test", "git://rev.git", "references", s)

	s = s.FastForward(app.Rev)

	return
}

func TestProcTypeRegister(t *testing.T) {
	s, app := proctypeSetup("reg123")
	pty := NewProcType(app, "whoop", s)

	pty, err := pty.Register()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.conn.Exists(pty.dir.Name)
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Errorf("proctype %s isn't registered", pty)
	}
}

func TestProcTypeRegisterWithInvalidName1(t *testing.T) {
	s, app := proctypeSetup("reg1232")
	pty := NewProcType(app, "who-op", s)

	pty, err := pty.Register()
	if err != ErrBadPtyName {
		t.Errorf("invalid proc type name (who-op) did not raise error")
	}
	if err != ErrBadPtyName && err != nil {
		t.Fatal("wrong error was raised for invalid proc type name")
	}
}

func TestProcTypeRegisterWithInvalidName2(t *testing.T) {
	s, app := proctypeSetup("reg1233")
	pty := NewProcType(app, "who_op", s)

	pty, err := pty.Register()
	if err != ErrBadPtyName {
		t.Errorf("invalid proc type name (who_op) did not raise error")
	}
	if err != ErrBadPtyName && err != nil {
		t.Fatal("wrong error was raised for invalid proc type name")
	}
}

func TestProcTypeUnregister(t *testing.T) {
	s, app := proctypeSetup("unreg123")
	pty := NewProcType(app, "whoop", s)

	pty, err := pty.Register()
	if err != nil {
		t.Error(err)
	}

	err = pty.Unregister()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.exists(pty.dir.Name)
	if check {
		t.Errorf("proctype %s is still registered", pty)
	}
}
