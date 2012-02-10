package visor

import (
	"testing"
	"time"
)

func eventSetup(name string) (s Snapshot, app *App, l chan *Event) {
	s, err := DialConn(DEFAULT_ADDR, "/event-test")
	if err != nil {
		panic(err)
	}
	r, _ := s.conn.Rev()
	err = s.conn.Del("/", r)
	s = s.FastForward(-1)

	app = &App{Name: name, RepoUrl: "git://" + name, Stack: Stack(name + "stack"), Snapshot: s}
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
	}

	s = s.FastForward(app.Rev)

	go WatchEvent(s, l)

	_, err = app.Unregister()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvAppUnreg, "unregcat", l, t)
}
func TestEventRevRegistered(t *testing.T) {
	s, app, l := eventSetup("regdog")
	rev, _ := NewRevision(app, "stable", s)
	rev = rev.FastForward(s.Rev)

	go WatchEvent(s, l)

	_, err := rev.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvRevReg, "regdog", l, t)
}
func TestEventRevUnregistered(t *testing.T) {
	s, app, l := eventSetup("unregdog")
	rev, _ := NewRevision(app, "stable", s)
	rev, err := rev.FastForward(s.Rev).Register()
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
func TestEventInstanceRegistered(t *testing.T) {
	s, app, l := eventSetup("regmouse")
	rev, _ := NewRevision(app, "stable", s)
	ins, _ := NewInstance(rev, "127.0.0.1:8080", "web", 0, s)

	go WatchEvent(s, l)

	_, err := ins.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvInsReg, "regmouse", l, t)
}
func TestEventInstanceUnregistered(t *testing.T) {
	s, app, l := eventSetup("unregmouse")
	rev, _ := NewRevision(app, "stable", s)
	ins, _ := NewInstance(rev, "127.0.0.1:8080", "web", 0, s)
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
	// TODO: Waiting on instance state API
}

func expectEvent(etype EventType, appname string, l chan *Event, t *testing.T) {
	for {
		select {
		case event := <-l:
			if event.Path["app"] == appname {
				if event.Type == etype {
					return
				} else if event.Type >= 0 {
					t.Errorf("received incorrect event type: expected %d got %d", etype, event.Type)
					return
				}
			}
		case <-time.After(time.Second):
			t.Errorf("expected event type %d got timeout", etype)
			return
		}
	}
}
