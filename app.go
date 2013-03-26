// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	cp "github.com/soundcloud/cotterpin"
	"path"
	"strings"
)

const appsPath = "apps"
const DeployLXC = "lxc"

type Env map[string]string

type App struct {
	Dir        cp.Dir
	Name       string
	RepoUrl    string
	Stack      string
	Head       string
	Env        Env
	DeployType string
}

// NewApp returns a new App given a name, repository url and stack.
func (s *Store) NewApp(name string, repourl string, stack string) (app *App) {
	app = &App{Name: name, RepoUrl: repourl, Stack: stack, Env: Env{}}
	app.Dir = cp.Dir{s.GetSnapshot(), path.Join(appsPath, app.Name)}

	return
}

func (a *App) GetSnapshot() cp.Snapshot {
	return a.Dir.Snapshot
}

// Join advances the App in time. It returns a new
// instance of Application at the rev of the supplied
// cp.Snapshotable.
func (a *App) Join(s cp.Snapshotable) *App {
	tmp := *a
	tmp.Dir.Snapshot = s.GetSnapshot()
	return &tmp
}

// Register adds the App to the global process state.
func (a *App) Register() (*App, error) {
	// Explicit FastForward to assure existence
	// check against latest state
	s, err := a.Dir.Snapshot.FastForward()
	if err != nil {
		return nil, err
	}
	a = a.Join(s)

	exists, _, err := a.Dir.Snapshot.Exists(a.Dir.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrConflict
	}

	if a.DeployType == "" {
		a.DeployType = DeployLXC
	}

	v := map[string]interface{}{
		"repo-url":    a.RepoUrl,
		"stack":       a.Stack,
		"deploy-type": a.DeployType,
	}
	attrs := cp.NewFile(a.Dir.Prefix("attrs"), v, new(cp.JsonCodec), a.Dir.Snapshot)

	_, err = attrs.Save()
	if err != nil {
		return nil, err
	}

	for k, v := range a.Env {
		_, err = a.SetEnvironmentVar(k, v)
		if err != nil {
			return nil, err
		}
	}

	d, err := a.Dir.Set("registered", timestamp())
	if err != nil {
		return nil, err
	}

	return a.Join(d), err
}

// Unregister removes the App form the global process state.
func (a *App) Unregister() error {
	return a.Dir.Del("/")
}

// SetHead sets the application's latest revision
func (a *App) SetHead(head string) (a1 *App, err error) {
	d, err := a.Dir.Set("head", head)
	if err != nil {
		return
	}
	a1 = a.Join(d)
	a1.Head = head

	return
}

// EnvironmentVars returns all set variables for this app as a map.
func (a *App) EnvironmentVars() (vars Env, err error) {
	names, err := a.Dir.Snapshot.Getdir(a.Dir.Prefix("env"))

	vars = Env{}

	type resp struct {
		key, val string
		err      error
	}
	ch := make(chan resp, len(names))

	if err != nil {
		if cp.IsErrNoEnt(err) {
			return vars, nil
		} else {
			return
		}
	}

	for _, name := range names {
		go func(name string) {
			v, err := a.GetEnvironmentVar(name)
			if err != nil {
				ch <- resp{err: err}
			} else {
				ch <- resp{key: name, val: v}
			}
		}(name)
	}
	for i := 0; i < len(names); i++ {
		r := <-ch
		if r.err != nil {
			return nil, err
		} else {
			vars[strings.Replace(r.key, "-", "_", -1)] = r.val
		}
	}
	return
}

// GetEnvironmentVar returns the value stored for the given key.
func (a *App) GetEnvironmentVar(k string) (value string, err error) {
	k = strings.Replace(k, "_", "-", -1)
	val, _, err := a.Dir.Get("env/" + k)
	if err != nil {
		if cp.IsErrNoEnt(err) {
			err = ErrNotFound
		}
		return
	}
	value = string(val)

	return
}

// SetEnvironmentVar stores the value for the given key.
func (a *App) SetEnvironmentVar(k string, v string) (app *App, err error) {
	d, err := a.Dir.Set("env/"+strings.Replace(k, "_", "-", -1), v)
	if err != nil {
		return
	}
	if _, present := a.Env[k]; !present {
		a.Env[k] = v
	}
	app = a.Join(d)
	return
}

// DelEnvironmentVar removes the env variable for the given key.
func (a *App) DelEnvironmentVar(k string) (*App, error) {
	err := a.Dir.Del("env/" + strings.Replace(k, "_", "-", -1))
	if err != nil {
		return nil, err
	}
	s, err := a.Dir.Snapshot.FastForward()
	if err != nil {
		return nil, err
	}
	return a.Join(s), nil
}

// GetRevisions returns all registered Revisions for the App
func (a *App) GetRevisions() (revisions []*Revision, err error) {
	s := storeFromSnapshotable(a)
	revs, err := s.GetSnapshot().Getdir(a.Dir.Prefix("revs"))
	if err != nil {
		return
	}

	ch, errch := cp.GetSnapshotables(revs, func(name string) (cp.Snapshotable, error) {
		return s.GetRevision(a, name)
	})
	for i := 0; i < len(revs); i++ {
		select {
		case r := <-ch:
			revisions = append(revisions, r.(*Revision))
		case err := <-errch:
			return nil, err
		}
	}
	return
}

// GetProcTypes returns all registered ProcTypes for the App
func (a *App) GetProcTypes() (ptys []*ProcType, err error) {
	s := storeFromSnapshotable(a)
	p := a.Dir.Prefix(procsPath)

	names, err := a.Dir.Snapshot.Getdir(p)
	if err != nil || len(names) == 0 {
		if cp.IsErrNoEnt(err) {
			err = nil
		}
		return
	}
	ch, errch := cp.GetSnapshotables(names, func(name string) (cp.Snapshotable, error) {
		return s.GetProcType(a, name)
	})
	for i := 0; i < len(names); i++ {
		select {
		case r := <-ch:
			ptys = append(ptys, r.(*ProcType))
		case err := <-errch:
			return nil, err
		}
	}
	return
}

// WatchEvent watches for events related to the app
func (a *App) WatchEvent(listener chan *Event) {
	s := storeFromSnapshotable(a)
	ch := make(chan *Event)
	go s.WatchEvent(ch)

	for e := range ch {
		if e.Path.App != nil && *e.Path.App == a.Name {
			listener <- e
		}
		if i, ok := e.Source.(*Instance); ok && i.AppName == a.Name {
			listener <- e
		}
	}
}

func (a *App) String() string {
	return fmt.Sprintf("App<%s>{stack: %s, type: %s}", a.Name, a.Stack, a.DeployType)
}

func (a *App) Inspect() string {
	return fmt.Sprintf("%#v", a)
}

// GetApp fetches an app with the given name.
func (s *Store) GetApp(name string) (app *App, err error) {
	app = s.NewApp(name, "", "")

	f, err := s.GetSnapshot().GetFile(app.Dir.Prefix("attrs"), new(cp.JsonCodec))
	if err != nil {
		return nil, ErrNotFound
	}

	value := f.Value.(map[string]interface{})

	app.RepoUrl = value["repo-url"].(string)
	app.Stack = value["stack"].(string)
	app.DeployType = value["deploy-type"].(string)

	f, err = s.GetSnapshot().GetFile(app.Dir.Prefix("head"), new(cp.StringCodec))
	if err == nil {
		app.Head = f.Value.(string)
	} else if cp.IsErrNoEnt(err) {
		err = nil
	}
	return
}

// Apps returns the list of all registered Apps.
func (s *Store) Apps() (apps []*App, err error) {
	exists, _, err := s.GetSnapshot().Exists(appsPath)
	if err != nil || !exists {
		return
	}

	names, err := s.GetSnapshot().Getdir(appsPath)
	if err != nil {
		return
	}

	ch, errch := cp.GetSnapshotables(names, func(name string) (cp.Snapshotable, error) {
		return s.GetApp(name)
	})

	for i := 0; i < len(names); i++ {
		select {
		case r := <-ch:
			apps = append(apps, r.(*App))
		case err := <-errch:
			return nil, err
		}
	}
	return
}
