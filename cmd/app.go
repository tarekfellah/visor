package main

import (
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
)

func app(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {

	switch subCommand {
	case "list":
		err = app_list(options, arguments, passThrough)
	case "describe":
		err = app_describe(options, arguments, passThrough)
	case "setenv":
		err = app_setenv(options, arguments, passThrough)
	case "getenv":
		err = app_getenv(options, arguments, passThrough)
	case "register":
		err = app_register(options, arguments, passThrough)
	case "env":
		err = app_env(options, arguments, passThrough)
	case "revisions":
		err = app_revisions(options, arguments, passThrough)
	}

	return
}

func app_list(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	print("\napp_list\n")
	print("\n")

	return
}

func app_describe(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	name := arguments[0]

	print("\napp_describe\n")
	print("\n\tname                  : " + name)
	print("\n")
	return
}

func app_setenv(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
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

func app_getenv(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	name := arguments[0]
	key := arguments[1]

	print("\napp_getenv\n")
	print("\n\tname                  : " + name)
	print("\n\tkey                   : " + key)
	print("\n")

	return
}

func app_register(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
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

func app_env(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	name := arguments[0]

	print("\napp_env\n")
	print("\n\tname                  : " + name)
	print("\n")

	return
}

func app_revisions(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	name := arguments[0]

	print("\napp_revisions\n")
	print("\n\tname                  : " + name)
	print("\n")

	return
}
