// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

// TODO: use return-often style

import (
	"errors"
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
	"os"
)

func Revision(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "exists":
		err = RevisionExists(arguments[0], arguments[1])
	case "describe":
		err = RevisionDescribe(arguments[0], arguments[1], options)
	case "unregister":
		err = RevisionUnregister(arguments[0], arguments[1])
	case "register":
		err = RevisionRegister(arguments[0], arguments[1], options["artifacturl"].String)
	case "instances":
		err = RevisionInstances(arguments[0], arguments[1])
	}
	return
}

func RevisionRegister(appName string, revision string, artifactUrl string) (err error) {
	snapshot := snapshot()
	var app *visor.App

	if app, err = visor.GetApp(snapshot, appName); err == nil {
		rev := &visor.Revision{
			App:        app,
			Ref:        revision,
			Snapshot:   snapshot,
			ArchiveUrl: artifactUrl,
		}

		rev, err = rev.Register()
		if err == visor.ErrKeyConflict {
			err = errors.New("Revision '" + revision + "' for app '" + appName + "' already registered!")
		}
	}

	return
}

func RevisionExists(appName string, revision string) (err error) {
	snapshot := snapshot()

	if app, err := visor.GetApp(snapshot, appName); err == nil {
		if _, err = visor.GetRevision(snapshot, app, revision); err != nil {
			os.Exit(-1)
		}
	}
	return
}

func RevisionDescribe(appName string, revision string, options map[string]getopt.OptionValue) (err error) {
	snapshot := snapshot()
	var app *visor.App
	fmtStr := "%-15.15s: %s\n"

	if app, err = visor.GetApp(snapshot, appName); err == nil {
		var rev *visor.Revision
		if rev, err = visor.GetRevision(snapshot, app, revision); err == nil {
			if onlyArtifactUrl, exists := options["artifacturl"]; exists == true && onlyArtifactUrl.Bool == true {
				fmt.Print(rev.ArchiveUrl)
			} else {
				fmt.Print(app.RepoUrl)
				fmt.Println()
				fmt.Printf(fmtStr, "App", appName)
				fmt.Printf(fmtStr, "Revision", rev.Ref)
				fmt.Printf(fmtStr, "Artifact-Url", rev.ArchiveUrl)
				fmt.Println()
			}
		}
	}
	return
}

func RevisionUnregister(appName string, revision string) (err error) {
	snapshot := snapshot()
	var app *visor.App

	if app, err = visor.GetApp(snapshot, appName); err == nil {
		var rev *visor.Revision
		if rev, err = visor.GetRevision(snapshot, app, revision); err == nil {
			err = rev.Unregister()
		}
	}

	return
}

func RevisionInstances(appName string, revision string) (err error) {
	// TODO
	print("\nrevision_instances\n")

	print("\n\tapp                  : " + appName)
	print("\n\trevision             : " + revision)
	print("\n")
	return
}
