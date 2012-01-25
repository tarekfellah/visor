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
		return ErrAppConflict
	}

	rev, err := a.setPath(c, "repo-url", a.RepoUrl)
	if err != nil {
		return
	}
	c.Rev = rev

	rev, err = a.setPath(c, "stack", string(a.Stack))
	if err != nil {
		return
	}
	c.Rev = rev

	return
}
func (a *App) Unregister(c *Client) error {
	return c.Del(a.Path())
}
func (a *App) Revisions() []Revision {
	return nil
}
func (a *App) RegisterRevision(rev string) (*Revision, error) {
	return nil, nil
}
func (a *App) UnregisterRevision(r *Revision) error {
	return nil
}
func (a *App) EnvironmentVars(c *Client) (vars map[string]string, err error) {
	varNames, err := c.Conn.Getdir(a.Path()+"/env", c.Rev, 0, -1)
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
	body, rev, err := c.Conn.Get(a.Path()+"/env/"+k, &c.Rev)
	if err != nil {
		return
	}

	if rev == 0 {
		return value, ErrKeyNotFound
	}

	c.Rev = rev

	value = string(body)

	return
}
func (a *App) SetEnvironmentVar(c *Client, k string, v string) (err error) {
	rev, err := a.setPath(c, "env/"+k, v)

	c.Rev = rev

	return
}
func (a *App) DelEnvironmentVar(c *Client, k string) (err error) {
	err = c.Conn.Del(a.Path()+"/env/"+k, c.Rev)
	if err != nil {
		return
	}

	rev, err := c.Conn.Rev()
	if err != nil {
		return
	}
	c.Rev = rev

	return
}
func (a *App) String() string {
	return "App{\"" + a.Name + "\"}"
}

func (a *App) Path() (p string) {
	return "/apps/" + a.Name
}
func (a *App) setPath(c *Client, k string, v string) (int64, error) {
	path := strings.Join([]string{a.Path(), k}, "/")

	return c.Conn.Set(path, c.Rev, []byte(v))
}

func Apps(c *Client) (apps []*App, err error) {
	rev, err := c.Conn.Rev()
	if err != nil {
		return
	}

	appNames, err := c.Conn.Getdir("/apps", rev, 0, -1)
	// FIXME proper error handling
	if err != nil {
		return []*App{}, nil
	}
	apps = make([]*App, len(appNames))

	for i := range appNames {
		apps[i] = &App{Name: appNames[i]}
	}

	return
}
