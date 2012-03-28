package main

import (
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
	return
}

func app_describe(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\napp_describe\n")
	return
}

func app_setenv(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\napp_setenv\n")
	return
}

func app_getenv(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\napp_getenv\n")
	return
}

func app_register(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\napp_register\n")
	return
}

func app_env(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\napp_env\n")
	return
}

func app_revisions(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\napp_revisions\n")
	return
}
