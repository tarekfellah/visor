package main

import (
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"os"
)

func main() {
	optionDefinition := getopt.Options{
		{"app", "app's name", getopt.IsArg | getopt.Required, ""},
		{"rev", "revision", getopt.IsArg | getopt.Required, ""},
		{"proc", "proc type that is to be scaled", getopt.IsArg |  getopt.Required, ""},
		{"num", "number of instances that should be running of this app-rev-proc-type (N for absolute values, -N for scaling down by N, +N for scaling up by N)", getopt.IsArg | getopt.Required, ""},
	}

	_, _, _, e := optionDefinition.ParseCommandLine()

	os.Args[0] = "visor [opts] revision scale"

	if e != nil {
		exit_code := 0
		description := "Scale application-revision-proc-type to a number of instances"

		switch {
		case e.ErrorCode == getopt.WantsUsage:
			fmt.Print(optionDefinition.Usage())
		case e.ErrorCode == getopt.WantsHelp:
			fmt.Print(optionDefinition.Help(description))
		default:
			fmt.Println("**** Error: ", e.Message, "\n", optionDefinition.Help(description))
			exit_code = e.ErrorCode
		}
		os.Exit(exit_code)
	}
}
