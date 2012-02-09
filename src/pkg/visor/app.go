package visor

import (
	"fmt"
	"strings"
	"time"
)

const APPS_PATH = "apps"

type App struct {
	Snapshot
	Name    string
	RepoUrl string
	Stack   Stack
}

var appMetaKeys = []string{"repo-url", "stack"}

// NewApp returns a new App given a name, repository url and stack.
func NewApp(name string, repourl string, stack Stack, snapshot Snapshot) (app *App, err error) {
	app = &App{Name: name, RepoUrl: repourl, Stack: stack, Snapshot: snapshot}
	return
}

func (a *App) CreateSnapshot(rev int64) (app Snapshotable) {
	app = &App{Name: a.Name, RepoUrl: a.RepoUrl, Stack: a.Stack, Snapshot: Snapshot{rev, a.conn}}
	return
}

func (a *App) FastForward(rev int64) (app *App) {
	return a.Snapshot.FastForward(a, rev).(*App)
}

// Register adds the App to the global process state.
func (a *App) Register() (app *App, err error) {
	exists, _, err := a.conn.Exists(a.Path(), &a.rev)
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	_, err = a.setPath("registered", time.Now().UTC().String())
	if err != nil {
		return
	}

	_, err = a.setPath("repo-url", a.RepoUrl)
	if err != nil {
		return
	}

	rev, err := a.setPath("stack", string(a.Stack))
	if err != nil {
		return a, err
	}
	app = a.FastForward(rev)

	return
}

// Unregister removes the App form the global process state.
func (a *App) Unregister() (app *App, err error) {
	err = a.conn.Del(a.Path(), a.rev)
	if err != nil {
		return
	}
	// TODO: do something smarter
	app = a.FastForward(0)

	return
}

// EnvironmentVars returns all set variables for this app as a map.
func (a *App) EnvironmentVars() (vars map[string]string, err error) {
	varNames, err := a.conn.Getdir(a.Path()+"/env", a.rev)
	if err != nil {
		return
	}

	vars = map[string]string{}
	var v string

	for i := range varNames {
		v, err = a.GetEnvironmentVar(varNames[i])
		if err != nil {
			return
		}

		vars[varNames[i]] = v
	}

	return
}

// GetEnvironmentVar returns the value stored for the given key.
func (a *App) GetEnvironmentVar(k string) (value string, err error) {
	val, _, err := a.conn.Get(a.Path()+"/env/"+k, &a.rev)
	if err != nil {
		return
	}
	value = string(val)

	return
}

// SetEnvironmentVar stores the value for the given key.
func (a *App) SetEnvironmentVar(k string, v string) (app *App, err error) {
	rev, err := a.setPath("env/"+k, v)
	if err != nil {
		return
	}
	app = a.FastForward(rev)
	return
}

// DelEnvironmentVar removes the env variable for the given key.
func (a *App) DelEnvironmentVar(k string) (app *App, err error) {
	err = a.delPath("env/" + k)
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

func (a *App) setPath(k string, v string) (rev int64, err error) {
	return a.conn.Set(a.prefixPath(k), a.rev, []byte(v))
}
func (a *App) delPath(k string) error {
	return a.conn.Del(a.prefixPath(k), a.rev)
}

// Apps returns the list of all registered Apps.
func Apps(s Snapshot) (apps []*App, err error) {
	names, err := s.conn.Getdir(APPS_PATH, s.rev)
	if err != nil {
		return
	}
	apps = make([]*App, len(names))

	for i := range names {
		a, e := NewApp(names[i], "", "", s)
		if e != nil {
			return nil, e
		}

		vals, e := s.conn.GetMulti(a.Path(), appMetaKeys, s.rev)
		if e != nil {
			return nil, e
		}

		a.RepoUrl = string(vals["repo-url"])
		a.Stack = Stack(string(vals["stack"]))
		apps[i] = a
	}

	return
}
