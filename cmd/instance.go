package main

import (
	getopt "github.com/kesselborn/go-getopt"
)

func Instance(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "describe":
		err = InstanceDescribe(options, arguments, passThrough)
	case "tail":
		err = InstanceTail(options, arguments, passThrough)
	case "kill":
		err = InstanceKill(options, arguments, passThrough)
	}
	return
}
func InstanceDescribe(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	instanceId := arguments[0]

	print("\ninstance_describe\n")
	print("\n\tinstance id           : " + instanceId)
	print("\n")
	return
}

func InstanceTail(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	instanceId := arguments[0]

	print("\ninstance_tail\n")
	print("\n\tinstance id           : " + instanceId)
	print("\n")
	return
}

func InstanceKill(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	instanceId := arguments[0]
	signal := options["signal"].String

	print("\ninstance_kill\n")
	print("\n\tinstance id           : " + instanceId)
	print("\n\tsignal                : " + signal)
	print("\n")
	return
}
