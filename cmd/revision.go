package main

import (
	getopt "github.com/kesselborn/go-getopt"
)

func revision(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	switch subCommand {
	case "describe":
		return_code = revision_describe(options, arguments, passThrough)
	case "unregister":
		return_code = revision_unregister(options, arguments, passThrough)
	case "scale":
		return_code = revision_scale(options, arguments, passThrough)
	case "instances":
		return_code = revision_instances(options, arguments, passThrough)
	}
	return
}

func revision_describe(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\nrevision_describe\n")
	return
}

func revision_unregister(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\nrevision_unregister\n")
	return
}

func revision_scale(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\nrevision_scale\n")
	return
}

func revision_instances(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (return_code int) {
	print("\nrevision_instances\n")
	return
}
