// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
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
	return NewApp(name, "git://"+name, Stack(name+"stack"), s)
}

func TestEventAppRegistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("regcat", s)

	go WatchEvent(s, l)

	_, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvAppReg, map[string]string{"app": "regcat"}, l, t)
}

func TestEventAppUnregistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("unregcat", s)

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

	expectEvent(EvAppUnreg, map[string]string{"app": "unregcat"}, l, t)
}

func TestEventRevRegistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("regdog", s)
	emitter := map[string]string{"app": "regdog", "rev": "stable"}

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

	expectEvent(EvRevReg, emitter, l, t)
}

func TestEventRevUnregistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("unregdog", s)
	emitter := map[string]string{"app": "unregdog", "rev": "stable"}

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

	expectEvent(EvRevUnreg, emitter, l, t)
}

func TestEventProcTypeRegistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("regstar", s)
	emitter := map[string]string{"app": "regstar", "proctype": "all"}

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

	expectEvent(EvProcReg, emitter, l, t)
}

func TestEventProcTypeUnregistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("unregstar", s)
	pty := NewProcType(app, "all", s)
	emitter := map[string]string{"app": "unregstar", "proctype": "all"}

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

	expectEvent(EvProcUnreg, emitter, l, t)
}

func TestEventInstanceRegistered(t *testing.T) {
	s, l := eventSetup()
	app := eventAppSetup("regmouse", s)
	ins, _ := NewInstance("web", "stable", app.Name, "127.0.0.1:8080", s)
	emitter := map[string]string{"app": "regmouse", "proctype": "web", "rev": "stable"}

	go WatchEvent(s, l)

	_, err := ins.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvInsReg, emitter, l, t)
}

func TestEventInstanceUnregistered(t *testing.T) {
	s, l := eventSetup()
	ins, _ := NewInstance("web", "stable", "unregmouse", "127.0.0.1:8080", s)
	emitter := map[string]string{"app": "unregmouse", "proctype": "web"}

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

	expectEvent(EvInsUnreg, emitter, l, t)
}

func TestEventInstanceStateChange(t *testing.T) {
	s, l := eventSetup()
	ins, _ := NewInstance("web-state", "stable-state", "statemouse", "127.0.0.1:8081", s)
	emitter := map[string]string{"app": "statemouse", "proctype": "web-state", "rev": "stable-state"}

	ins, err := ins.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(ins.Rev)

	go WatchEvent(s, l)

	ins, err = ins.UpdateState(InsStateStarted)
	if err != nil {
		t.Error(err)
	}
	expectEvent(EvInsStart, emitter, l, t)

	ins, err = ins.UpdateState(InsStateFailed)
	if err != nil {
		t.Error(err)
	}
	expectEvent(EvInsFail, emitter, l, t)

	ins, err = ins.UpdateState(InsStateExited)
	if err != nil {
		t.Error(err)
	}
	expectEvent(EvInsExit, emitter, l, t)

	_, err = ins.UpdateState(InsStateDead)
	if err != nil {
		t.Error(err)
	}
	expectEvent(EvInsDead, emitter, l, t)
}

func TestEventSrvRegistered(t *testing.T) {
	s, l := eventSetup()
	srv := NewService("eventsrv", s)

	go WatchEvent(s, l)

	srv, err := srv.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvSrvReg, map[string]string{"service": "eventsrv"}, l, t)
}

func TestEventSrvUnregistered(t *testing.T) {
	s, l := eventSetup()
	srv := NewService("eventunsrv", s)

	srv, err := srv.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(srv.Rev)

	go WatchEvent(s, l)

	err = srv.Unregister()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvSrvUnreg, map[string]string{"service": "eventunsrv"}, l, t)
}

func TestEventEpRegistered(t *testing.T) {
	s, l := eventSetup()
	srv := NewService("eventep", s)
	ep, err := NewEndpoint(srv, "1.2.3.4", 1000, s)
	if err != nil {
		t.Error(err)
	}

	go WatchEvent(s, l)

	ep, err = ep.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvEpReg, map[string]string{"service": "eventep", "endpoint": "1-2-3-4-1000"}, l, t)
}

func TestEventEpUnregistered(t *testing.T) {
	s, l := eventSetup()
	srv := NewService("eventunep", s)
	ep, err := NewEndpoint(srv, "4.3.2.1", 2000, s)
	if err != nil {
		t.Error(err)
	}

	ep, err = ep.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(ep.Rev)

	go WatchEvent(s, l)

	err = ep.Unregister()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvEpUnreg, map[string]string{"service": "eventunep", "endpoint": "4-3-2-1-2000"}, l, t)
}

func expectEvent(etype EventType, emitterMap map[string]string, l chan *Event, t *testing.T) {
	for {
		select {
		case event := <-l:
			if event.Type == etype {
				for key, value := range emitterMap {
					if event.Emitter[key] != value {
						t.Errorf("received incorrect emitter field %s: expected %s got %s", key, value, event.Emitter[key])
					}
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
