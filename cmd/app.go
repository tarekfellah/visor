package main

import (
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
)

func app(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {

	switch subCommand {
	case "list":
		return_code = app_list(options, arguments, passThrough)
	case "describe":
		return_code = app_describe(options, arguments, passThrough)
	case "setenv":
		return_code = app_setenv(options, arguments, passThrough)
	case "getenv":
		return_code = app_getenv(options, arguments, passThrough)
	case "register":
		return_code = app_register(options, arguments, passThrough)
	case "env":
		return_code = app_env(options, arguments, passThrough)
	case "revisions":
		return_code = app_revisions(options, arguments, passThrough)
	}

	return
}

func app_list(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\napp_list\n")
	print("\n")

	return
}

func app_describe(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	name := arguments[0]

	print("\napp_describe\n")
	print("\n\tname                  : " + name)
	print("\n")
	return
}

func app_setenv(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
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

func app_getenv(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	name := arguments[0]
	key := arguments[1]

	print("\napp_getenv\n")
	print("\n\tname                  : " + name)
	print("\n\tkey                   : " + key)
	print("\n")

	return
}

func app_register(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	appType := options["type"].String
	repourl := options["repourl"].String
	stack := options["stack"].String
	irc_announce_channels := options["irc"].StrArray
	name := arguments[0]

	print("\napp_register\n")

	print("\n\ttype                  : " + appType)
	print("\n\trepourl               : " + repourl)
	print("\n\tstack                 : " + stack)
	print("\n\tirc_announce_channels : " + fmt.Sprintf("%#v", irc_announce_channels))
	print("\n\tname                  : " + name)
	print("\n")
	return
}

func app_env(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	name := arguments[0]

	print("\napp_env\n")
	print("\n\tname                  : " + name)
	print("\n")

	return
}

func app_revisions(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	name := arguments[0]

	print("\napp_revisions\n")
	print("\n\tname                  : " + name)
	print("\n")

	return
}
