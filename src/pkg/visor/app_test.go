package visor

import (
	"testing"
)

func appSetup(name string) (c *Client, app *App) {
	app, err := NewApp(name, "git://cat.git", "whiskers")
	if err != nil {
		panic(err)
	}
	c, err = Dial(DEFAULT_ADDR, DEFAULT_ROOT)
	if err != nil {
		panic(err)
	}

	r, _ := c.conn.Rev()
	err = c.conn.Del(c.prefixPath("apps"), r)
	c, _ = c.FastForward(-1)

	return
}

func TestAppRegistration(t *testing.T) {
	c, app := appSetup("lolcatapp")

	check, _, err := c.conn.Exists(c.prefixPath(app.Path()), nil)
	if err != nil {
		t.Error(err)
		return
	}
	if check {
		t.Error("App already registered")
		return
	}

	app, err = app.Register(c)
	if err != nil {
		t.Error(err)
		return
	}
	check, _, err = c.conn.Exists(c.prefixPath(app.Path()), &app.rev)
	if err != nil {
		t.Error(err)
		return
	}
	if !check {
		t.Error("App registration failed")
		return
	}

	_, err = app.Register(c)
	if err == nil {
		t.Error("App allowed to be registered twice")
	}
}

func TestAppUnregistration(t *testing.T) {
	c, app := appSetup("dog")

	app, err := app.Register(c)
	if err != nil {
		t.Error(err)
		return
	}

	app, err = app.Unregister(c)
	if err != nil {
		t.Error(err)
		return
	}

	check, _, err := c.conn.Exists(c.prefixPath(app.Path()), nil)
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("App still registered")
	}
}

func TestSetAndGetEnvironmentVar(t *testing.T) {
	c, app := appSetup("lolcatapp")

	app, err := app.SetEnvironmentVar(c, "meow", "w00t")
	if err != nil {
		t.Error(err)
		return
	}

	value, err := app.GetEnvironmentVar(c, "meow")
	if err != nil {
		t.Error(err)
		return
	}

	if value != "w00t" {
		t.Errorf("EnvironmentVar 'meow' expected %s got %s", "w00t", value)
	}
}

func TestSetAndDelEnvironmentVar(t *testing.T) {
	c, app := appSetup("catalolna")

	app, err := app.SetEnvironmentVar(c, "wuff", "lulz")
	if err != nil {
		t.Error(err)
	}

	app, err = app.DelEnvironmentVar(c, "wuff")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = app.GetEnvironmentVar(c, "wuff")
	if err == nil {
		t.Error(err)
		t.Error("EnvironmentVar wasn't deleted")
		return
	}

	if err != ErrKeyNotFound {
		t.Error(err)
	}
}

func TestEnvironmentVars(t *testing.T) {
	c, app := appSetup("cat-A-log")

	_, err := app.SetEnvironmentVar(c, "whiskers", "purr")
	if err != nil {
		t.Error(err)
	}
	app, err = app.SetEnvironmentVar(c, "lasers", "pew pew")
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

func TestApps(t *testing.T) {
	c, _ := appSetup("apps-test")
	names := []string{"cat", "dog", "lol"}

	for i := range names {
		a, err := NewApp(names[i], "zebra", "joke")
		if err != nil {
			t.Error(err)
		}
		_, err = a.Register(c)
		if err != nil {
			t.Error(err)
		}
	}
	c, _ = c.FastForward(-1)

	apps, err := Apps(c)
	if err != nil {
		t.Error(err)
	}
	if len(apps) != len(names) {
		t.Errorf("expected length %d returned length %d", len(names), len(apps))
	} else {
		for i := range apps {
			if apps[i].Name != names[i] {
				t.Error("expected %s got %s", names[i], apps[i].Name)
			}
		}
	}
}
