package main

import (
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
	"strconv"
)

func App(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {

	switch subCommand {
	case "list":
		err = AppList()
	case "describe":
		err = AppDescribe(arguments[0])
	case "setenv":
		value := ""
		if len(arguments) > 2 {
			value = arguments[2]
		}
		err = AppSetenv(arguments[0], arguments[1], value)
	case "getenv":
		err = AppGetenv(options, arguments, passThrough)
	case "register":
		err = AppRegister(arguments[0], options["type"].String, options["repourl"].String, options["stack"].String)
	case "env":
		err = AppEnv(options, arguments, passThrough)
	case "revisions":
		err = AppRevisions(options, arguments, passThrough)
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

func AppDescribe(name string) (err error) {
	fmtStr := "%15.15s: %s\n"

	app, err := visor.GetApp(snapshot(), name)

	if err == nil {
		fmt.Println()
		fmt.Printf(fmtStr, "Name", app.Name)
		fmt.Printf(fmtStr, "Repo-Url", app.RepoUrl)
		fmt.Printf(fmtStr, "Stack", app.Stack)
		fmt.Printf(fmtStr, "Deploy-Type", app.DeployType)
		fmt.Println()
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

func AppGetenv(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	name := arguments[0]
	key := arguments[1]

	print("\napp_getenv\n")
	print("\n\tname                  : " + name)
	print("\n\tkey                   : " + key)
	print("\n")

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

func AppEnv(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	name := arguments[0]

	print("\napp_env\n")
	print("\n\tname                  : " + name)
	print("\n")

	return
}

func AppRevisions(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	name := arguments[0]

	print("\napp_revisions\n")
	print("\n\tname                  : " + name)
	print("\n")

	return
}
