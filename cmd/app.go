// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
	"os"
	"strconv"
)

func App(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {

	switch subCommand {
	case "list":
		err = AppList()
	case "exists":
		err = AppExists(arguments[0])
	case "describe":
		err = AppDescribe(arguments[0], options)
	case "setenv":
		value := ""
		if len(arguments) > 2 {
			value = arguments[2]
		}
		err = AppSetenv(arguments[0], arguments[1], value)
	case "getenv":
		err = AppGetenv(arguments[0], arguments[1])
	case "register":
		err = AppRegister(arguments[0], options["type"].String, options["repourl"].String, options["stack"].String)
	case "unregister":
		err = AppUnRegister(arguments[0])
	case "env":
		err = AppEnv(arguments[0])
	case "revisions":
		err = AppRevisions(arguments[0])
	}

	return
}

func AppList() (err error) {
	entryFmtStr := "| %-3.3s | %-30.30s | %-40.40s | %-9.9s | %-15.15s |\n"
	rulerFmtStr := "+-%-3.3s-+-%-30.30s-+-%-40.40s-+-%-9.9s-+-%-15.15s-+\n"
	ruler := "--------------------------------------------------"

	var apps []*visor.App

	if apps, err = visor.Apps(snapshot()); err == nil {
		fmt.Println()
		fmt.Printf(rulerFmtStr, ruler, ruler, ruler, ruler, ruler)
		fmt.Printf(entryFmtStr, "No.", "Name", "Repo-Url", "Stack", "Deploy-Type")
		fmt.Printf(rulerFmtStr, ruler, ruler, ruler, ruler, ruler)
		for i, app := range apps {
			fmt.Printf(entryFmtStr, strconv.Itoa(i), app.Name, app.RepoUrl, app.Stack, app.DeployType)
		}
		fmt.Printf(rulerFmtStr, ruler, ruler, ruler, ruler, ruler)
		fmt.Println()
	}

	return
}

func AppExists(name string) (err error) {
	if _, err = visor.GetApp(snapshot(), name); err != nil {
		os.Exit(-1)
	}

	return
}

func AppDescribe(name string, options map[string]getopt.OptionValue) (err error) {
	fmtStr := "%-15.15s: %s\n"

	app, err := visor.GetApp(snapshot(), name)

	if err == nil {
		if onlyStack, exists := options["stack"]; exists == true && onlyStack.Bool == true {
			fmt.Print(app.Stack)
		} else if onlyType, exists := options["type"]; exists == true && onlyType.Bool == true {
			fmt.Print(app.DeployType)
		} else if onlyUrl, exists := options["repourl"]; exists == true && onlyUrl.Bool == true {
			fmt.Print(app.RepoUrl)
		} else {
			fmt.Println()
			fmt.Printf(fmtStr, "Name", app.Name)
			fmt.Printf(fmtStr, "Repo-Url", app.RepoUrl)
			fmt.Printf(fmtStr, "Stack", app.Stack)
			fmt.Printf(fmtStr, "Deploy-Type", app.DeployType)
			fmt.Println()
		}
	}

	return
}

func AppSetenv(name string, key string, value string) (err error) {
	var app *visor.App
	app, err = visor.GetApp(snapshot(), name)

	if err == nil {
		if value != "" {
			_, err = app.SetEnvironmentVar(key, value)
		} else {
			_, err = app.DelEnvironmentVar(key)
		}
	}

	return
}

func AppGetenv(name string, key string) (err error) {
	var app *visor.App
	app, err = visor.GetApp(snapshot(), name)

	if err == nil {
		var value string
		if value, err = app.GetEnvironmentVar(key); err == nil {
			fmt.Println(value)
		}
	}

	return
}

func AppUnRegister(name string) (err error) {
	var app *visor.App
	app, err = visor.GetApp(snapshot(), name)

	if err == nil {
		answer := ""

		for answer != "y" && answer != "n" {
			fmt.Printf("\nThis will delete the app and all revisions from bazooka and stop all running instances. Really continue? (y/n): ")
			_, err = fmt.Scanf("%s", &answer)
		}

		if answer == "n" {
			fmt.Println("pussy!")
		} else {
			// TODO: delete revisions and stop instances
			err = app.Unregister()
		}
	}

	return

}

func AppEnv(name string) (err error) {
	var app *visor.App
	app, err = visor.GetApp(snapshot(), name)

	if err == nil {
		var envVars map[string]string
		if envVars, err = app.EnvironmentVars(); err == nil {
			for key, value := range envVars {
				fmt.Printf("%s=%s\n", key, value)
			}
		}
	}

	return
}

func AppRegister(name string, deployType string, repoUrl string, stack string) (err error) {

	app := &visor.App{Name: name, RepoUrl: repoUrl, Stack: visor.Stack(stack), Snapshot: snapshot(), DeployType: deployType}
	app, err = app.Register()

	if err != nil {
		print(err.Error())
	}

	return
}

func AppRevisions(appName string) (err error) {
	entryFmtStr := "| %-3.3s | %-10.10s | %-10.10s | %-100.100s |\n"
	rulerFmtStr := "+-%-3.3s-+-%-10.10s-+-%-10.10s-+-%-100.100s-+\n"
	ruler := "-----------------------------------------------------------------------------------------------------------"

	var app *visor.App
	snapshot := snapshot()

	if app, err = visor.GetApp(snapshot, appName); err == nil {
		var revs []*visor.Revision

		if revs, err = visor.AppRevisions(snapshot, app); err == nil {
			fmt.Println()
			fmt.Printf(rulerFmtStr, ruler, ruler, ruler, ruler)
			fmt.Printf(entryFmtStr, "No.", "App", "Revision", "Archive-Url")
			fmt.Printf(rulerFmtStr, ruler, ruler, ruler, ruler)
			for i, rev := range revs {
				fmt.Printf(entryFmtStr, strconv.Itoa(i), appName, rev.Ref, rev.ArchiveUrl)
			}
			fmt.Printf(rulerFmtStr, ruler, ruler, ruler, ruler)
		}
	}

	return
}
