package visor

import (
	"testing"
)

func revSetup() (c *Client, app *App) {
	c, err := Dial(DEFAULT_ADDR, DEFAULT_ROOT, new(ByteCodec))
	if err != nil {
		panic(err)
	}
	app, err = NewApp("rev-test", "git://rev.git", "references", c.Snapshot)
	if err != nil {
		panic(err)
	}

	c.Del("/apps")
	c, _ = c.FastForward(-1)

	return
}

func TestRevisionRegister(t *testing.T) {
	c, app := revSetup()
	rev, err := NewRevision(app, "stable", app.Snapshot)
	if err != nil {
		t.Error(err)
	}

	check, _, err := c.conn.Exists(c.prefixPath(rev.Path()), nil)
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

	check, _, err = c.conn.Exists(c.prefixPath(rev.Path()), nil)
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
	c, app := revSetup()
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

	check, err := c.Exists(rev.Path())
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("Revision still registered")
	}
}
