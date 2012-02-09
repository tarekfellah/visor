package visor

import (
	"testing"
	"time"
)

func eventSetup(name string) (c *Client, app *App, l chan *Event) {
	c, err := Dial(DEFAULT_ADDR, "/event-test", new(ByteCodec))
	if err != nil {
		panic(err)
	}
	r, _ := c.conn.Rev()
	err = c.conn.Del("/", r)
	c, _ = c.FastForward(-1)

	app = &App{Name: name, RepoUrl: "git://" + name, Stack: Stack(name + "stack"), Snapshot: c.Snapshot}
	l = make(chan *Event)

	return
}

func TestEventAppRegistered(t *testing.T) {
	c, app, l := eventSetup("regcat")

	go WatchEvent(c, l)

	_, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvAppReg, "regcat", l, t)

}
func TestEventAppUnregistered(t *testing.T) {
	c, app, l := eventSetup("unregcat")
	app, err := app.Register()
	if err != nil {
		t.Error(err)
	}

	c, _ = c.FastForward(app.rev)

	go WatchEvent(c, l)

	_, err = app.Unregister()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvAppUnreg, "unregcat", l, t)
}
func TestEventRevRegistered(t *testing.T) {
	c, app, l := eventSetup("regdog")
	rev, _ := NewRevision(app, "stable", c.Snapshot)
	rev = rev.FastForward(c.rev)

	go WatchEvent(c, l)

	_, err := rev.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvRevReg, "regdog", l, t)
}
func TestEventRevUnregistered(t *testing.T) {
	c, app, l := eventSetup("unregdog")
	rev, _ := NewRevision(app, "stable", c.Snapshot)
	rev, err := rev.FastForward(c.rev).Register()
	if err != nil {
		t.Error(err)
		return
	}
	c, _ = c.FastForward(rev.rev)

	go WatchEvent(c, l)

	err = rev.Unregister()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvRevUnreg, "unregdog", l, t)
}
func TestEventInstanceRegistered(t *testing.T) {
	c, app, l := eventSetup("regmouse")
	rev, _ := NewRevision(app, "stable", c.Snapshot)
	ins, _ := NewInstance(rev, "127.0.0.1:8080", "web", 0, c.Snapshot)

	go WatchEvent(c, l)

	_, err := ins.Register()
	if err != nil {
		t.Error(err)
	}

	expectEvent(EvInsReg, "regmouse", l, t)
}
func TestEventInstanceUnregistered(t *testing.T) {
	c, app, l := eventSetup("unregmouse")
	rev, _ := NewRevision(app, "stable", c.Snapshot)
	ins, _ := NewInstance(rev, "127.0.0.1:8080", "web", 0, c.Snapshot)
	ins, err := ins.Register()
	if err != nil {
		t.Error(err)
	}
	c, _ = c.FastForward(ins.rev)

	go WatchEvent(c, l)

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
