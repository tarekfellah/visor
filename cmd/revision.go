package main

import (
	"errors"
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
)

func Revision(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "describe":
		err = RevisionDescribe(options, arguments, passThrough)
	case "unregister":
		err = RevisionUnregister(options, arguments, passThrough)
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

func RevisionDescribe(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	app := arguments[0]
	revision := arguments[1]

	print("\nrevision_describe\n")

	print("\n\tapp                  : " + app)
	print("\n\trevision             : " + revision)
	print("\n")
	return
}

func RevisionUnregister(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	app := arguments[0]
	revision := arguments[1]

	print("\nrevision_unregister\n")

	print("\n\tapp                  : " + app)
	print("\n\trevision             : " + revision)
	print("\n")
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
