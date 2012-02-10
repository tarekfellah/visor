package visor

import (
	"testing"
)

func revSetup() (s Snapshot, app *App) {
	s, err := DialConn(DEFAULT_ADDR, DEFAULT_ROOT)
	if err != nil {
		panic(err)
	}
	app, err = NewApp("rev-test", "git://rev.git", "references", s)
	if err != nil {
		panic(err)
	}

	s.conn.Del("/apps", -1)
	s = s.FastForward(-1)

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
