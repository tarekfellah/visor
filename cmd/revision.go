package main

import (
	"errors"
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
)

func Revision(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "describe":
		err = RevisionDescribe(arguments[0], arguments[1])
	case "unregister":
		err = RevisionUnregister(arguments[0], arguments[1])
	case "register":
		err = RevisionRegister(arguments[0], arguments[1], options["artifacturl"].String, options["proctypes"].StrArray)
	case "scale":
		err = RevisionScale(options, arguments, passThrough)
	case "instances":
		err = RevisionInstances(options, arguments, passThrough)
	}
	return
}

func RevisionRegister(appName string, revision string, artifactUrl string, procTypes []string) (err error) {
	snapshot := snapshot()
	var app *visor.App

	if app, err = visor.GetApp(snapshot, appName); err == nil {
		_, err = (&visor.Revision{App: app, Ref: revision, Snapshot: snapshot, ArchiveUrl: artifactUrl}).Register()
	}

	if err == visor.ErrKeyConflict {
		err = errors.New("Revision '" + revision + "' for app '" + appName + "' already registered!")
	}

	return
}

func RevisionDescribe(appName string, revision string) (err error) {
	snapshot := snapshot()
	var app *visor.App
	fmtStr := "%-15.15s: %s\n"

	if app, err = visor.GetApp(snapshot, appName); err == nil {
		var rev *visor.Revision
		if rev, err = visor.GetRevision(snapshot, app, revision); err == nil {
			fmt.Println()
			fmt.Printf(fmtStr, "App", appName)
			fmt.Printf(fmtStr, "Revision", rev.Ref)
			fmt.Printf(fmtStr, "Artifact-Url", rev.ArchiveUrl)
			fmt.Println()
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

func RevisionScale(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	app := arguments[0]
	revision := arguments[1]
	proc := arguments[2]
	num := arguments[3]

	print("\nrevision_scale\n")

	print("\n\tapp                  : " + app)
	print("\n\trevision             : " + revision)
	print("\n\tproc                 : " + proc)
	print("\n\tnum                  : " + num)
	print("\n")
	return
}

func RevisionInstances(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	app := arguments[0]
	revision := arguments[1]

	print("\nrevision_instances\n")

	print("\n\tapp                  : " + app)
	print("\n\trevision             : " + revision)
	print("\n")
	return
}
