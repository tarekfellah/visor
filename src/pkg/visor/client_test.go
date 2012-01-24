package visor

import (
	"testing"
)

func clientSetup() (c *Client) {
	c = createClient()
	c.Deldir("/apps", c.Rev)

	return
}

func TestAppRegistration(t *testing.T) {
	clientSetup()

	name := "mobile-prod"
	repoUrl := "git://ashdkha"
	stack := Stack("blossom")

	c := createClient()
	check, err := appIsRegistered(c, name)
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("App already registered")
	}

	_, err = c.RegisterApp(name, repoUrl, stack)
	if err != nil {
		t.Error(err)
	}

	check, err = appIsRegistered(c, name)
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Error("App registration failed")
	}

	_, err = c.RegisterApp(name, repoUrl, stack)
	if err == nil {
		t.Error("App allowed to be registered twice")
	}
}

func TestAppUnregistration(t *testing.T) {
	clientSetup()

	name := "mobile-prod"
	repoUrl := "git://ashdkha"
	stack := Stack("blossom")

	c := createClient()
	app, err := c.RegisterApp(name, repoUrl, stack)
	if err != nil {
		t.Error(err)
	}

	err = c.UnregisterApp(app)
	if err != nil {
		t.Error(err)
	}

	check, err := appIsRegistered(c, name)
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("App still registered")
	}
}

// HELPER
func createClient() (c *Client) {
	c, err := Dial(DEFAULT_ADDR)
	if err != nil {
		panic(err)
	}

	return
}

func appIsRegistered(c *Client, name string) (isRegistered bool, err error) {
	apps, err := c.Apps()
	if err != nil {
		return
	}

	isRegistered = false

	for i := range apps {
		if apps[i].Name == name {
			isRegistered = true
		}
	}

	return
}
