// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"path"
	"strings"
	"time"
)

const appsPath = "apps"
const DeployLXC = "lxc"

type Env map[string]string

type App struct {
	dir
	Name       string
	RepoUrl    string
	Stack      string
	Env        Env
	DeployType string
}

// NewApp returns a new App given a name, repository url and stack.
func NewApp(name string, repourl string, stack string, snapshot Snapshot) (app *App) {
	app = &App{Name: name, RepoUrl: repourl, Stack: stack, Env: Env{}}
	app.dir = dir{snapshot, path.Join(appsPath, app.Name)}

	return
}

func (a *App) createSnapshot(rev int64) snapshotable {
	tmp := *a
	tmp.Snapshot = Snapshot{rev, a.conn}
	return &tmp
}

// FastForward advances the application in time. It returns
// a new instance of Application with the supplied revision.
func (a *App) FastForward(rev int64) (app *App) {
	return a.Snapshot.fastForward(a, rev).(*App)
}

// Register adds the App to the global process state.
func (a *App) Register() (app *App, err error) {
	exists, _, err := a.conn.Exists(a.dir.Name)
	if err != nil {
		return nil, fmt.Errorf("application '%s' is already registered", a.Name)
	}
	if exists {
		return nil, ErrKeyConflict
	}

	if a.DeployType == "" {
		a.DeployType = DeployLXC
	}

	attrs := &file{
		Snapshot: a.Snapshot,
		codec:    new(jsonCodec),
		dir:      a.dir.prefix("attrs"),
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

	rev, err := a.set("registered", time.Now().UTC().String())
	if err != nil {
		return
	}

	app = a.FastForward(rev)

	return
}

// Unregister removes the App form the global process state.
func (a *App) Unregister() error {
	return a.del("/")
}

// EnvironmentVars returns all set variables for this app as a map.
func (a *App) EnvironmentVars() (vars Env, err error) {
	varNames, err := a.getdir(a.dir.prefix("env"))

	vars = Env{}

	if err != nil {
		if IsErrNoEnt(err) {
			return vars, nil
		} else {
			return
		}
	}

	var v string

	for i := range varNames {
		v, err = a.GetEnvironmentVar(varNames[i])
		if err != nil {
			return
		}

		vars[strings.Replace(varNames[i], "-", "_", -1)] = v
	}

	return
}

// GetEnvironmentVar returns the value stored for the given key.
func (a *App) GetEnvironmentVar(k string) (value string, err error) {
	k = strings.Replace(k, "_", "-", -1)
	val, _, err := a.get("env/" + k)
	if err != nil {
		return
	}
	value = string(val)

	return
}

// SetEnvironmentVar stores the value for the given key.
func (a *App) SetEnvironmentVar(k string, v string) (app *App, err error) {
	rev, err := a.set("env/"+strings.Replace(k, "_", "-", -1), v)
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
	err = a.del("env/" + strings.Replace(k, "_", "-", -1))
	if err != nil {
		return
	}
	app = a.FastForward(a.Rev + 1)
	return
}

// GetProcTypes returns all registered ProcTypes for the App
func (a *App) GetProcTypes() (ptys []*ProcType, err error) {
	p := a.dir.prefix(procsPath)

	names, err := a.getdir(p)
	if err != nil {
		if IsErrNoEnt(err) {
			err = nil
		}
		return
	}

	ptych := make(chan *ProcType, len(names))
	errch := make(chan error, 1)

	for _, name := range names {
		go func(name string) {
			pty, err := GetProcType(a.Snapshot, a, name)
			if err != nil {
				errch <- err
			} else {
				ptych <- pty
			}
		}(name)
	}
	for {
		select {
		case pty := <-ptych:
			ptys = append(ptys, pty)

			if len(ptys) == len(names) {
				return
			}
		case err = <-errch:
			return
		}
	}
	return
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

	f, err := s.getFile(app.dir.prefix("attrs"), new(jsonCodec))
	if err != nil {
		return nil, err
	}

	value := f.Value.(map[string]interface{})

	app.RepoUrl = value["repo-url"].(string)
	app.Stack = value["stack"].(string)
	app.DeployType = value["deploy-type"].(string)

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

	for i := range names {
		var app *App

		app, err = GetApp(s, names[i])
		if err != nil {
			return
		}

		apps = append(apps, app)
	}

	return
}
