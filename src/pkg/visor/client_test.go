package visor

import (
	"testing"
)

func init() {
	c := createClient()
	err := c.Deldir("/apps", c.Rev)
	if err != nil {
		panic(err)
	}
}

func TestAppRegistration(t *testing.T) {
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
