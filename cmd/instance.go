package main

import (
	getopt "github.com/kesselborn/go-getopt"
)

func instance(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "describe":
		err = instance_describe(options, arguments, passThrough)
	case "tail":
		err = instance_tail(options, arguments, passThrough)
	case "kill":
		err = instance_kill(options, arguments, passThrough)
	}
	return
}
func instance_describe(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	instanceId := arguments[0]

	print("\ninstance_describe\n")
	print("\n\tinstance id           : " + instanceId)
	print("\n")
	return
}

func instance_tail(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	instanceId := arguments[0]

	print("\ninstance_tail\n")
	print("\n\tinstance id           : " + instanceId)
	print("\n")
	return
}

func instance_kill(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	instanceId := arguments[0]
	signal := options["signal"].String

	print("\ninstance_kill\n")
	print("\n\tinstance id           : " + instanceId)
	print("\n\tsignal                : " + signal)
	print("\n")
	return
}
