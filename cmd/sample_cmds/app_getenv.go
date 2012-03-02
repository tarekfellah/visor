package main

import (
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"os"
)

func main() {
	optionDefinition := getopt.Options{
		{"name", "app's name", getopt.IsArg | getopt.Required, ""},
		{"key", "environment variable's name", getopt.IsArg | getopt.Required, ""},
	}

	_, _, _, e := optionDefinition.ParseCommandLine()

	os.Args[0] = "visor [opts] app getenv"

	if e != nil {
		exit_code := 0
		description := "Gets an app environment variable"

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
