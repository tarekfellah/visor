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
		{"artifacturl|u", "url of the build artifact for this application-revision", getopt.Required, ""},
		{"proctypes|t", "comma seperated list of proc types available in this application-revision", getopt.Optional, ""},
	}

	_, _, _, e := optionDefinition.ParseCommandLine()

	os.Args[0] = "visor [opts] revision register"

	if e != nil {
		exit_code := 0
		description := ""

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
