// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"path"
	"strings"
)

const appsPath = "apps"
const DeployLXC = "lxc"

type Env map[string]string

type App struct {
	Dir        dir
	Name       string
	RepoUrl    string
	Stack      string
	Head       string
	Env        Env
	DeployType string
}

// NewApp returns a new App given a name, repository url and stack.
func NewApp(name string, repourl string, stack string, snapshot Snapshot) (app *App) {
	app = &App{Name: name, RepoUrl: repourl, Stack: stack, Env: Env{}}
	app.Dir = dir{snapshot, path.Join(appsPath, app.Name)}

	return
}

func (a *App) createSnapshot(rev int64) snapshotable {
	tmp := *a
	tmp.Dir.Snapshot = Snapshot{rev, a.Dir.Snapshot.conn}
	return &tmp
}

// FastForward advances the application in time. It returns
// a new instance of Application with the supplied revision.
func (a *App) FastForward(rev int64) (app *App) {
	return a.Dir.Snapshot.fastForward(a, rev).(*App)
}

// Register adds the App to the global process state.
func (a *App) Register() (app *App, err error) {
	exists, _, err := a.Dir.Snapshot.conn.Exists(a.Dir.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrKeyConflict
	}

	if a.DeployType == "" {
		a.DeployType = DeployLXC
	}

	attrs := &file{
		Snapshot: a.Dir.Snapshot,
		codec:    new(jsonCodec),
		dir:      a.Dir.prefix("attrs"),
		Value: map[string]interface{}{
			"repo-url":    a.RepoUrl,
			"stack":       a.Stack,
			"deploy-type": a.DeployType,
		},
	}

	_, err = attrs.Create()
	if err != nil {
		return
	}

	for k, v := range a.Env {
		_, err = a.SetEnvironmentVar(k, v)
		if err != nil {
			return
		}
	}

	rev, err := a.Dir.set("registered", timestamp())
	if err != nil {
		return
	}

	app = a.FastForward(rev)

	return
}

// Unregister removes the App form the global process state.
func (a *App) Unregister() error {
	return a.Dir.del("/")
}

// SetHead sets the application's latest revision
func (a *App) SetHead(head string) (a1 *App, err error) {
	rev, err := a.Dir.set("head", head)
	if err != nil {
		return
	}
	a1 = a.FastForward(rev)
	a1.Head = head

	return
}

// EnvironmentVars returns all set variables for this app as a map.
func (a *App) EnvironmentVars() (vars Env, err error) {
	names, err := a.Dir.Snapshot.getdir(a.Dir.prefix("env"))

	vars = Env{}

	type resp struct {
		key, val string
		err      error
	}
	ch := make(chan resp, len(names))

	if err != nil {
		if IsErrNoEnt(err) {
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
	val, _, err := a.Dir.get("env/" + k)
	if err != nil {
		return
	}
	value = string(val)

	return
}

// SetEnvironmentVar stores the value for the given key.
func (a *App) SetEnvironmentVar(k string, v string) (app *App, err error) {
	rev, err := a.Dir.set("env/"+strings.Replace(k, "_", "-", -1), v)
	if err != nil {
		return
	}
	if _, present := a.Env[k]; !present {
		a.Env[k] = v
	}
	app = a.FastForward(rev)
	return
}

// DelEnvironmentVar removes the env variable for the given key.
func (a *App) DelEnvironmentVar(k string) (app *App, err error) {
	err = a.Dir.del("env/" + strings.Replace(k, "_", "-", -1))
	if err != nil {
		return
	}
	app = a.FastForward(a.Dir.Snapshot.Rev + 1)
	return
}

// GetRevisions returns all registered Revisions for the App
func (a *App) GetRevisions() (revisions []*Revision, err error) {
	s := a.Dir.Snapshot
	revs, err := s.getdir(a.Dir.prefix("revs"))
	if err != nil {
		return
	}

	ch, errch := getSnapshotables(revs, func(name string) (snapshotable, error) {
		return GetRevision(s, a, name)
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
	p := a.Dir.prefix(procsPath)

	names, err := a.Dir.Snapshot.getdir(p)
	if err != nil || len(names) == 0 {
		if IsErrNoEnt(err) {
			err = nil
		}
		return
	}
	ch, errch := getSnapshotables(names, func(name string) (snapshotable, error) {
		return GetProcType(a.Dir.Snapshot, a, name)
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
	ch := make(chan *Event)
	go WatchEvent(a.Dir.Snapshot, ch)

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
func GetApp(s Snapshot, name string) (app *App, err error) {
	app = NewApp(name, "", "", s)

	f, err := s.getFile(app.Dir.prefix("attrs"), new(jsonCodec))
	if err != nil {
		return nil, err
	}

	value := f.Value.(map[string]interface{})

	app.RepoUrl = value["repo-url"].(string)
	app.Stack = value["stack"].(string)
	app.DeployType = value["deploy-type"].(string)

	f, err = s.getFile(app.Dir.prefix("head"), new(stringCodec))
	if err == nil {
		app.Head = f.Value.(string)
	} else if IsErrNoEnt(err) {
		err = nil
	}
	return
}

// Apps returns the list of all registered Apps.
func Apps(s Snapshot) (apps []*App, err error) {
	exists, _, err := s.conn.Exists(appsPath)
	if err != nil || !exists {
		return
	}

	names, err := s.getdir(appsPath)
	if err != nil {
		return
	}

	ch, errch := getSnapshotables(names, func(name string) (snapshotable, error) {
		return GetApp(s, name)
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
