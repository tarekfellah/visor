package visor

import (
	"testing"
)

func revSetup(ref string) (c *Client, rev *Revision) {
	app := &App{Name: "rev-test", RepoUrl: "git://rev.git", Stack: "references"}
	rev = &Revision{App: app, ref: ref}
	c, err := Dial(DEFAULT_ADDR)
	if err != nil {
		panic(err)
	}

	c.Del("/apps")

	return
}

func TestRevisionRegister(t *testing.T) {
	c, rev := revSetup("master")

	check, err := c.Exists(rev.Path())
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("Revision already registered")
	}

	err = rev.Register(c)
	if err != nil {
		t.Error(err)
	}

	check, err = c.Exists(rev.Path())
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Error("Revision registration failed")
	}

	err = rev.Register(c)
	if err == nil {
		t.Error("Revision allowed to be registered twice")
	}
}

func TestRevisionUnregister(t *testing.T) {
	c, rev := revSetup("stable")

	err := rev.Register(c)
	if err != nil {
		t.Error(err)
	}

	err = rev.Unregister(c)
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
