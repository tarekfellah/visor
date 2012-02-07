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
	rev     int64
}

var appMetaKeys = []string{"repo-url", "stack"}

// NewApp returns a new App given a name, repository url and stack.
func NewApp(name string, repourl string, stack Stack) (app *App, err error) {
	app = &App{Name: name, RepoUrl: repourl, Stack: stack}
	return
}

// Register adds the App to the global process state.
func (a *App) Register(c *Client) (app *App, err error) {
	c, _ = c.FastForward(a.rev)

	exists, err := c.Exists(a.Path())
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	_, err = a.setPath(c, "registered", time.Now().UTC().String())
	if err != nil {
		return
	}

	_, err = a.setPath(c, "repo-url", a.RepoUrl)
	if err != nil {
		return
	}

	rev, err := a.setPath(c, "stack", string(a.Stack))
	if err != nil {
		return a, err
	}
	app = a.FastForward(rev)

	return
}

// Unregister removes the App form the global process state.
func (a *App) Unregister(c *Client) (app *App, err error) {
	c, _ = c.FastForward(a.rev)

	err = c.Del(a.Path())
	if err != nil {
		return
	}
	// TODO: do something smarter
	app = a.FastForward(0)

	return
}

// EnvironmentVars returns all set variables for this app as a map.
func (a *App) EnvironmentVars(c *Client) (vars map[string]string, err error) {
	c, _ = c.FastForward(a.rev)

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
	c, _ = c.FastForward(a.rev)

	val, err := c.Get(a.Path() + "/env/" + k)
	if err != nil {
		return
	}
	// TODO: verify we are getting a string value
	// before calling String()
	value = string(val.Value.String())

	return
}

func (a *App) FastForward(rev int64) (app *App) {
	app = &App{Name: a.Name, RepoUrl: a.RepoUrl, rev: rev, Stack: a.Stack}
	return
}

// SetEnvironmentVar stores the value for the given key.
func (a *App) SetEnvironmentVar(c *Client, k string, v string) (app *App, err error) {
	c, _ = c.FastForward(a.rev)

	rev, err := a.setPath(c, "env/"+k, v)
	if err != nil {
		return
	}
	app = a.FastForward(rev)
	return
}

// DelEnvironmentVar removes the env variable for the given key.
func (a *App) DelEnvironmentVar(c *Client, k string) (app *App, err error) {
	c, _ = c.FastForward(a.rev)

	err = a.delPath(c, "env/"+k)
	if err != nil {
		return
	}
	app = a.FastForward(a.rev + 1)
	return
}

func (a *App) prefixPath(path string) string {
	return strings.Join([]string{a.Path(), path}, "/")
}

func (a *App) String() string {
	return fmt.Sprintf("%#v", a)
}

// Path returns the path for the App in the global process state.
func (a *App) Path() (p string) {
	return strings.Join([]string{APPS_PATH, a.Name}, "/")
}

func (a *App) setPath(c *Client, k string, v string) (rev int64, err error) {
	file, err := c.Set(a.prefixPath(k), []byte(v))
	if err != nil {
		return
	}
	return file.Rev, err
}
func (a *App) delPath(c *Client, k string) error {
	return c.Del(a.prefixPath(k))
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

		a.RepoUrl = vals["repo-url"].Value.String()
		a.Stack = Stack(vals["stack"].Value.String())
		apps[i] = a
	}

	return
}
