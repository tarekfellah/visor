package main

// TODO: use return-often style

import (
	"errors"
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
	"strconv"
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
		if scalingFactor, err := strconv.Atoi(arguments[3]); err == nil {
			err = RevisionScale(arguments[0], arguments[1], arguments[2], scalingFactor)
		}
	case "instances":
		err = RevisionInstances(options, arguments, passThrough)
	}
	return
}

func RevisionRegister(appName string, revision string, artifactUrl string, procTypes []string) (err error) {
	snapshot := snapshot()
	var app *visor.App

	if app, err = visor.GetApp(snapshot, appName); err == nil {
		var rev *visor.Revision
		if rev, err = (&visor.Revision{App: app, Ref: revision, Snapshot: snapshot, ArchiveUrl: artifactUrl}).Register(); err == nil {
			for _, pt := range procTypes {
				_, err = (&visor.ProcType{Name: visor.ProcessName(pt), Scale: 0, Revision: rev, Snapshot: snapshot}).Register()
			}
		} else {
			if err == visor.ErrKeyConflict {
				err = errors.New("Revision '" + revision + "' for app '" + appName + "' already registered!")
			}
		}
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
			fmt.Printf(fmtStr, "Proctypes", procTypeList(snapshot, rev))
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

func RevisionScale(appName string, revision string, procTypeName string, scalingFactor int) (err error) {
	snapshot := snapshot()
	var app *visor.App

	if app, err = visor.GetApp(snapshot, appName); err == nil {
		var rev *visor.Revision
		if rev, err = visor.GetRevision(snapshot, app, revision); err == nil {
			var procType *visor.ProcType
			if procType, err = visor.GetProcType(snapshot, rev, visor.ProcessName(procTypeName)); err == nil {
				procType.Scale = scalingFactor
				// TODO: PERSIST!
			}
		}
	}

	return
}

func RevisionInstances(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	// TODO
	app := arguments[0]
	revision := arguments[1]

	print("\nrevision_instances\n")

	print("\n\tapp                  : " + app)
	print("\n\trevision             : " + revision)
	print("\n")
	return
}
