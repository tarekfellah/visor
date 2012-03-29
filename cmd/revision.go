package main

import (
	getopt "github.com/kesselborn/go-getopt"
)

func revision(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "describe":
		err = revision_describe(options, arguments, passThrough)
	case "unregister":
		err = revision_unregister(options, arguments, passThrough)
	case "scale":
		err = revision_scale(options, arguments, passThrough)
	case "instances":
		err = revision_instances(options, arguments, passThrough)
	}
	return
}

func revision_describe(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	app := arguments[0]
	revision := arguments[1]

	print("\nrevision_describe\n")

	print("\n\tapp                  : " + app)
	print("\n\trevision             : " + revision)
	print("\n")
	return
}

func revision_unregister(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	app := arguments[0]
	revision := arguments[1]

	print("\nrevision_unregister\n")

	print("\n\tapp                  : " + app)
	print("\n\trevision             : " + revision)
	print("\n")
	return
}

func revision_scale(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
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

func revision_instances(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	app := arguments[0]
	revision := arguments[1]

	print("\nrevision_instances\n")

	print("\n\tapp                  : " + app)
	print("\n\trevision             : " + revision)
	print("\n")
	return
}
