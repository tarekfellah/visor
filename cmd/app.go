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
		err = AppList(options, arguments, passThrough)
	case "describe":
		err = AppDescribe(options, arguments, passThrough)
	case "setenv":
		err = AppSetenv(options, arguments, passThrough)
	case "getenv":
		err = AppGetenv(options, arguments, passThrough)
	case "register":
		err = AppRegister(options, arguments, passThrough)
	case "env":
		err = AppEnv(options, arguments, passThrough)
	case "revisions":
		err = AppRevisions(options, arguments, passThrough)
	}

	return
}

func AppList(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
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

func AppDescribe(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	name := arguments[0]

	print("\napp_describe\n")
	print("\n\tname                  : " + name)
	print("\n")
	return
}

func AppSetenv(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	name := arguments[0]
	key := arguments[1]

	print("\napp_setenv\n")
	print("\n\tname                  : " + name)
	print("\n\tkey                   : " + key)

	if len(arguments) > 2 {
		print("\n\tvalue                 : " + arguments[2])
	} else {
	}
	print("\n")

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

func AppRegister(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	deployType := options["type"].String
	repoUrl := options["repourl"].String
	stack := visor.Stack(options["stack"].String)
	//ircChannels  := options["irc"].StrArray
	name := arguments[0]

	app := &visor.App{Name: name, RepoUrl: repoUrl, Stack: stack, Snapshot: snapshot(), DeployType: deployType}
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
