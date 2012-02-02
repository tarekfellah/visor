package visor

import (
	"fmt"
	"strings"
	"time"
)

const APPS_PATH = "apps"

type App struct {
	Name    string
	RepoUrl string
	Stack   Stack
}

var appMetaKeys = []string{"repo-url", "stack"}

// NewApp returns a new App given a name, repository url and stack.
func NewApp(name string, repourl string, stack Stack) (app *App, err error) {
	app = &App{Name: name, RepoUrl: repourl, Stack: stack}
	return
}

// Register adds the App to the global process state.
func (a *App) Register(c *Client) (err error) {
	exists, err := c.Exists(a.Path())
	if err != nil {
		return
	}
	if exists {
		return ErrKeyConflict
	}

	err = a.setPath(c, "registered", time.Now().UTC().String())
	if err != nil {
		return
	}

	err = a.setPath(c, "repo-url", a.RepoUrl)
	if err != nil {
		return
	}

	err = a.setPath(c, "stack", string(a.Stack))
	if err != nil {
		return
	}

	return
}

// Unregister removes the App form the global process state.
func (a *App) Unregister(c *Client) error {
	return c.Del(a.Path())
}

// EnvironmentVars returns all set variables for this app as a map.
func (a *App) EnvironmentVars(c *Client) (vars map[string]string, err error) {
	varNames, err := c.Keys(a.Path() + "/env")
	if err != nil {
		return
	}

	vars = map[string]string{}
	var v string

	for i := range varNames {
		v, err = a.GetEnvironmentVar(c, varNames[i])
		if err != nil {
			return
		}

		vars[varNames[i]] = v
	}

	return
}

// GetEnvironmentVar returns the value stored for the given key.
func (a *App) GetEnvironmentVar(c *Client, k string) (value string, err error) {
	bytes, err := c.Get(a.Path() + "/env/" + k)
	if err != nil {
		return
	}
	value = string(bytes)

	return
}

// SetEnvironmentVar stores the value for the given key.
func (a *App) SetEnvironmentVar(c *Client, k string, v string) (err error) {
	return a.setPath(c, "env/"+k, v)
}

// DelEnvironmentVar removes the env variable for the given key.
func (a *App) DelEnvironmentVar(c *Client, k string) (err error) {
	err = c.Del(a.Path() + "/env/" + k)
	if err != nil {
		return
	}

	return
}

func (a *App) String() string {
	return fmt.Sprintf("%#v", a)
}

// Path returns the path for the App in the global process state.
func (a *App) Path() (p string) {
	return strings.Join([]string{APPS_PATH, a.Name}, "/")
}

func (a *App) setPath(c *Client, k string, v string) error {
	path := strings.Join([]string{a.Path(), k}, "/")

	return c.Set(path, []byte(v))
}

// Apps returns the list of all registered Apps.
func Apps(c *Client) (apps []*App, err error) {
	names, err := c.Keys(APPS_PATH)
	if err != nil {
		return
	}
	apps = make([]*App, len(names))

	for i := range names {
		a, e := NewApp(names[i], "", "")
		if e != nil {
			return nil, e
		}

		vals, e := c.GetMulti(a.Path(), appMetaKeys)
		if e != nil {
			return nil, e
		}

		a.RepoUrl = string(vals["repo-url"])
		a.Stack = Stack(vals["stack"])
		apps[i] = a
	}

	return
}
