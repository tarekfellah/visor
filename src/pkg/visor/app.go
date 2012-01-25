package visor

import (
	"strings"
)

type App struct {
	Name    string
	RepoUrl string
	Stack   Stack
}

func (a *App) Register(c *Client) (err error) {
	exists, err := c.Exists(a.Path())
	if err != nil {
		return
	}
	if exists {
		return ErrKeyConflict
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
func (a *App) Unregister(c *Client) error {
	return c.Del(a.Path())
}
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
func (a *App) GetEnvironmentVar(c *Client, k string) (value string, err error) {
	value, err = c.Get(a.Path() + "/env/" + k)
	if err != nil {
		return
	}

	return
}
func (a *App) SetEnvironmentVar(c *Client, k string, v string) (err error) {
	return a.setPath(c, "env/"+k, v)
}
func (a *App) DelEnvironmentVar(c *Client, k string) (err error) {
	err = c.Del(a.Path() + "/env/" + k)
	if err != nil {
		return
	}

	return
}
func (a *App) String() string {
	return "App{\"" + a.Name + "\"}"
}

func (a *App) Path() (p string) {
	return "/apps/" + a.Name
}
func (a *App) setPath(c *Client, k string, v string) error {
	path := strings.Join([]string{a.Path(), k}, "/")

	return c.Set(path, v)
}

func Apps(c *Client) (apps []*App, err error) {
	names, err := c.Keys("/apps")
	if err != nil {
		return
	}

	apps = make([]*App, len(names))

	for i := range names {
		apps[i] = &App{Name: names[i]}
	}

	return
}
