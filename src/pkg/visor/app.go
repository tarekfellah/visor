package visor

import (
	"strings"
)

type App struct {
	Name    string
	RepoUrl string
	Stack   Stack
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
	path := strings.Join([]string{"/apps", a.Name, "repo-url"}, "/")

	return c.Deldir(path, c.Rev)
}
func (a *App) EnvironmentVariables() (*map[string]string, error) {
	return nil, nil
}
func (a *App) GetEnvironmentVariable(k string) (string, error) {
	return "", nil
}
func (a *App) SetEnvironmentVariable(k string, v string) error {
	return nil
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
