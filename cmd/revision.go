package main

import (
	getopt "github.com/kesselborn/go-getopt"
)

func Revision(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "describe":
		err = RevisionDescribe(options, arguments, passThrough)
	case "unregister":
		err = RevisionUnregister(options, arguments, passThrough)
	case "scale":
		err = RevisionScale(options, arguments, passThrough)
	case "instances":
		err = RevisionInstances(options, arguments, passThrough)
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
