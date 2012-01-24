package visor

import (
	"testing"
)

func appSetup(name string) (c *Client, app *App) {
	c = clientSetup()
	err := c.Deldir("/apps", c.Rev)

	app, err = c.RegisterApp(name, "git://lolcathub", Stack("lol"))
	if err != nil {
		panic(err)
	}

	return
}

func TestSetAndGetEnvironmentVar(t *testing.T) {
	c, app := appSetup("lolcatapp")

	err := app.SetEnvironmentVar(c, "meow", "w00t")
	if err != nil {
		t.Error(err)
	}

	value, err := app.GetEnvironmentVar(c, "meow")
	if err != nil {
		t.Error(err)
	}

	if value != "w00t" {
		t.Errorf("EnvironmentVar 'meow' expected %s got %s", "w00t", value)
	}
}

func TestSetAndDetEnvironmentVar(t *testing.T) {
	c, app := appSetup("catalolna")

	err := app.SetEnvironmentVar(c, "wuff", "lulz")
	if err != nil {
		t.Error(err)
	}

	err = app.DelEnvironmentVar(c, "wuff")
	if err != nil {
		t.Error(err)
	}

	_, err = app.GetEnvironmentVar(c, "wuff")
	if err == nil {
		t.Error("EnvironmentVar wasn't deleted")
	}

	if err != ErrKeyNotFound {
		t.Error(err)
	}
}

func TestEnvironmentVars(t *testing.T) {
	c, app := appSetup("cat-A-log")

	err := app.SetEnvironmentVar(c, "whiskers", "purr")
	if err != nil {
		t.Error(err)
	}
	err = app.SetEnvironmentVar(c, "lasers", "pew pew")
	if err != nil {
		t.Error(err)
	}

	vars, err := app.EnvironmentVars(c)
	if err != nil {
		t.Error(err)
	}
	if vars["whiskers"] != "purr" {
		t.Error("Var not set")
	}
	if vars["lasers"] != "pew pew" {
		t.Error("Var not set")
	}
}
