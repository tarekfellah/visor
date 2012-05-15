package visor

import (
	"testing"
)

func revSetup() (s Snapshot, app *App) {
	s, err := Dial(DEFAULT_ADDR, "/revision-test")
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

	app, err = NewApp("rev-test", "git://rev.git", "references", s)
	if err != nil {
		panic(err)
	}

	return
}

func TestRevisionRegister(t *testing.T) {
	s, app := revSetup()
	rev, err := NewRevision(app, "stable", app.Snapshot)
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.conn.Exists(rev.Path(), nil)
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

	check, _, err = s.conn.Exists(rev.Path(), nil)
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
	rev, err := NewRevision(app, "master", app.Snapshot)
	if err != nil {
		t.Error(err)
	}

	rev, err = rev.Register()
	if err != nil {
		t.Error(err)
	}

	err = rev.Unregister()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.conn.Exists(rev.Path(), &s.Rev)
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("Revision still registered")
	}
}

func TestRevisionScaleUp(t *testing.T) {
	s, app := revSetup()
	rev, err := NewRevision(app, "12345", app.Snapshot)
	if err != nil {
		t.Error(err)
	}
	rev, err = rev.Register()
	if err != nil {
		t.Error(err)
	}

	rev, err = rev.Scale("web", 5)
	if err != nil {
		t.Error(err)
	}
	factor, _, err := s.conn.Get(rev.Path()+"/procs/web/scale", &rev.Rev)
	if err != nil {
		t.Error(err)
	}
	if string(factor) != "5" {
		t.Errorf("Scaling factor expected %s, got %s", "5", factor)
	}

	tickets, err := s.conn.Getdir(TICKETS_PATH, rev.Rev)
	if err != nil {
		t.Error(err)
	}
	if len(tickets) != 5 {
		t.Errorf("Expected tickets %s, got %d", "5", len(tickets))
	}
}
