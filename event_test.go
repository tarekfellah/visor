// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func eventSetup() (s Snapshot, l chan *Event) {
	s, err := Dial(DefaultAddr, "/event-test")
	if err != nil {
		panic(err)
	}
	r, _ := s.conn.Rev()
	err = s.conn.Del("/", r)

	s = s.FastForward(-1)

	rev, err := Init(s)
	if err != nil {
		panic(err)
	}

	s = s.FastForward(rev)

	l = make(chan *Event)

	return
}

func eventAppSetup(name string, s Snapshot) *App {
	return NewApp(name, "git://"+name, name+"stack", s)
}

func expectEvent(etype EventType, s snapshotable, l chan *Event, t *testing.T) (event *Event) {
	for {
		select {
		case event = <-l:
			if event.Type == etype {
				if reflect.TypeOf(event.Source) != reflect.TypeOf(s) {
					t.Errorf("types are not equal %#v != %#v", event.Source, s)
				}
			} else {
				t.Errorf("received incorrect event type: expected %s got %s %s", etype, event, event.Type)
			}
			return
		case <-time.After(time.Second):
			t.Errorf("expected event type %s got timeout", etype)
			return
		}
	}
	return
}

func TestEventAppRegistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("regcat", s)

	go WatchEvent(s, l)

	_, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	ev := expectEvent(EvAppReg, app, l, t)
	if ev.Path.App == nil || (*ev.Path.App != app.Name) {
		t.Error("event.Path doesn't contain expected data")
	}
}

func TestEventAppUnregistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("unregcat", s)

	app, err := app.Register()
	if err != nil {
		t.Error(err)
		return
	}

	s = s.FastForward(app.Dir.Snapshot.Rev)

	go WatchEvent(s, l)

	err = app.Unregister()
	if err != nil {
		t.Error(err)
	}

	ev := expectEvent(EvAppUnreg, nil, l, t)
	if ev.Path.App == nil || (*ev.Path.App != app.Name) {
		t.Error("event.Path doesn't contain expected data")
	}
}

func TestEventRevRegistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("regdog", s)

	app, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(app.Dir.Snapshot.Rev)

	rev := NewRevision(app, "stable", s)
	rev = rev.FastForward(s.Rev)

	go WatchEvent(s, l)

	_, err = rev.Register()
	if err != nil {
		t.Error(err)
	}

	ev := expectEvent(EvRevReg, rev, l, t)
	if ev.Path.Revision == nil || (*ev.Path.Revision != rev.Ref) {
		t.Error("event.Path doesn't contain expected data")
	}
	if ev.Path.App == nil || (*ev.Path.App != app.Name) {
		t.Error("event.Path doesn't contain expected data")
	}
}

func TestEventRevUnregistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("unregdog", s)

	app, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(app.Dir.Snapshot.Rev)

	rev := NewRevision(app, "stable", s)
	rev, err = rev.FastForward(s.Rev).Register()
	if err != nil {
		t.Error(err)
		return
	}
	s = s.FastForward(rev.Dir.Snapshot.Rev)

	go WatchEvent(s, l)

	err = rev.Unregister()
	if err != nil {
		t.Error(err)
	}

	ev := expectEvent(EvRevUnreg, nil, l, t)
	if ev.Path.Revision == nil || (*ev.Path.Revision != rev.Ref) {
		t.Error("event.Path doesn't contain expected data")
	}
}

func TestEventProcTypeRegistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("regstar", s)

	app, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(app.Dir.Snapshot.Rev)

	rev := NewRevision(app, "bang", s)
	rev, err = rev.FastForward(s.Rev).Register()
	if err != nil {
		t.Error(err)
		return
	}
	s = s.FastForward(rev.Dir.Snapshot.Rev)

	pty := NewProcType(app, "all", s)

	go WatchEvent(s, l)

	_, err = pty.Register()
	if err != nil {
		t.Error(err)
	}

	ev := expectEvent(EvProcReg, pty, l, t)
	if ev.Path.Proctype == nil || (*ev.Path.Proctype != pty.Name) {
		t.Error("event.Path doesn't contain expected data")
	}
	if ev.Path.App == nil || (*ev.Path.App != app.Name) {
		t.Error("event.Path doesn't contain expected data")
	}
}

func TestEventProcTypeUnregistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("unregstar", s)
	pty := NewProcType(app, "all", s)

	pty, err := pty.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(pty.Dir.Snapshot.Rev)

	go WatchEvent(s, l)

	err = pty.Unregister()
	if err != nil {
		t.Error(err)
	}

	ev := expectEvent(EvProcUnreg, nil, l, t)

	if ev.Path.Proctype == nil || (*ev.Path.Proctype != pty.Name) {
		t.Error("event.Path doesn't contain expected data")
	}
}

func TestEventInstanceRegistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("regmouse", s)

	go WatchEvent(s, l)

	ins, err := RegisterInstance(app.Name, "stable", "web", s)
	if err != nil {
		t.Fatal(err)
	}

	ev := expectEvent(EvInsReg, ins, l, t)

	if ev.Path.Instance == nil || (*ev.Path.Instance != strconv.FormatInt(ins.Id, 10)) {
		t.Error("event.Path doesn't contain expected data")
	}
}

func TestEventInstanceUnregistered(t *testing.T) {
	s, l := eventSetup()

	ins, err := RegisterInstance("unregmouse", "stable", "web", s)
	if err != nil {
		t.Fatal(err)
	}
	s = s.FastForward(ins.Dir.Snapshot.Rev)

	go WatchEvent(s, l)

	err = ins.Unregister()
	if err != nil {
		t.Fatal(err)
	}

	ev := expectEvent(EvInsUnreg, nil, l, t)
	if ev.Path.Instance == nil || (*ev.Path.Instance != strconv.FormatInt(ins.Id, 10)) {
		t.Error("event.Path doesn't contain expected data")
	}
}

func TestEventInstanceStateChange(t *testing.T) {
	ip := "10.0.0.1"
	port := 9999
	host := "mouse.org"
	s, l := eventSetup()

	ins, err := RegisterInstance("statemouse", "stable-state", "web-state", s)
	if err != nil {
		t.Fatal(err)
	}
	s = s.FastForward(ins.Dir.Snapshot.Rev)

	ins, err = ins.Claim(ip)
	if err != nil {
		t.Fatal(err)
	}

	go WatchEvent(s, l)

	ins, err = ins.Started(ip, port, host)
	if err != nil {
		t.Error(err)
	}
	ev := expectEvent(EvInsStart, ins, l, t)
	if ev.Path.Instance == nil || (*ev.Path.Instance != strconv.FormatInt(ins.Id, 10)) {
		t.Error("event.Path doesn't contain expected data")
	}

	instance := ev.Source.(*Instance)

	if instance.Ip != ip || instance.Host != host || instance.Port != port {
		t.Fatal("instance fields don't match")
	}

	ins, err = ins.Failed(ip, errors.New("no reason."))
	if err != nil {
		t.Error(err)
	}
	expectEvent(EvInsFail, ins, l, t)

	ins, err = ins.Exited(ip)
	if err != nil {
		t.Error(err)
	}
	expectEvent(EvInsExit, ins, l, t)
}
