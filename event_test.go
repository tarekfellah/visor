// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
	"time"
)

func eventSetup(name string) (s Snapshot, app *App, l chan *Event) {
	s, err := Dial(DEFAULT_ADDR, "/event-test")
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

	app = NewApp(name, "git://"+name, Stack(name+"stack"), s)
	l = make(chan *Event)

	return
}

func TestEventAppRegistered(t *testing.T) {
	s, app, l := eventSetup("regcat")

	go WatchEvent(s, l)

	_, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvAppReg, "regcat", l, t)

}
func TestEventAppUnregistered(t *testing.T) {
	s, app, l := eventSetup("unregcat")
	app, err := app.Register()
	if err != nil {
		t.Error(err)
		return
	}

	s = s.FastForward(app.Rev)

	go WatchEvent(s, l)

	err = app.Unregister()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvAppUnreg, "unregcat", l, t)
}
func TestEventRevRegistered(t *testing.T) {
	s, app, l := eventSetup("regdog")

	app, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(app.Rev)

	rev := NewRevision(app, "stable", s)
	rev = rev.FastForward(s.Rev)

	go WatchEvent(s, l)

	_, err = rev.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvRevReg, "regdog", l, t)
}
func TestEventRevUnregistered(t *testing.T) {
	s, app, l := eventSetup("unregdog")

	app, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(app.Rev)

	rev := NewRevision(app, "stable", s)
	rev, err = rev.FastForward(s.Rev).Register()
	if err != nil {
		t.Error(err)
		return
	}
	s = s.FastForward(rev.Rev)

	go WatchEvent(s, l)

	err = rev.Unregister()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvRevUnreg, "unregdog", l, t)
}
func TestEventProcTypeRegistered(t *testing.T) {
	s, app, l := eventSetup("regstar")

	app, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(app.Rev)

	rev := NewRevision(app, "bang", s)
	rev, err = rev.FastForward(s.Rev).Register()
	if err != nil {
		t.Error(err)
		return
	}
	s = s.FastForward(rev.Rev)

	pty := NewProcType(app, "all", s)

	go WatchEvent(s, l)

	_, err = pty.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvProcReg, "regstar", l, t)
}
func TestEventProcTypeUnregistered(t *testing.T) {
	s, app, l := eventSetup("unregstar")
	pty := NewProcType(app, "all", s)
	pty, err := pty.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(pty.Rev)

	go WatchEvent(s, l)

	err = pty.Unregister()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvProcUnreg, "unregstar", l, t)
}
func TestEventInstanceRegistered(t *testing.T) {
	s, app, l := eventSetup("regmouse")
	ins, _ := NewInstance("web", "stable", app.Name, "127.0.0.1:8080", s)

	go WatchEvent(s, l)

	_, err := ins.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvInsReg, "regmouse", l, t)
}
func TestEventInstanceUnregistered(t *testing.T) {
	s, app, l := eventSetup("unregmouse")
	ins, _ := NewInstance("web", "stable", app.Name, "127.0.0.1:8080", s)
	ins, err := ins.Register()
	if err != nil {
		t.Error(err)
	}
	s = s.FastForward(ins.Rev)

	go WatchEvent(s, l)

	err = ins.Unregister()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvInsUnreg, "unregmouse", l, t)
}

func TestEventInstanceStateChange(t *testing.T) {
	s, _, l := eventSetup("statemouse")
	ins, _ := NewInstance("web-state", "stable-state", "statemouse", "127.0.0.1:8081", s)

	ins, err := ins.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(ins.Rev)

	go WatchEvent(s, l)

	_, err = ins.UpdateState(InsStateStarted)
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvInsStateChange, "", l, t)
}

func expectEvent(etype EventType, appname string, l chan *Event, t *testing.T) {
	for {
		select {
		case event := <-l:
			if event.Type == etype {
				if event.Path["app"] != appname {
					t.Errorf("received incorrect app name: expected %s got %s", appname, event.Path["app"])
				}
				return
			} else if event.Type >= 0 {
				t.Errorf("received incorrect event type: expected %d got %d", etype, event.Type)
				return
			}
		case <-time.After(time.Second):
			t.Errorf("expected event type %d got timeout", etype)
			return
		}
	}
}
