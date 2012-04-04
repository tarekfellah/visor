package main

import (
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"os"
)

func main() {
	optionDefinition := getopt.Options{
		{"instanceid", "id of the instance (<hostname>:<port>)", getopt.IsArg | getopt.Required, ""},
		{"signalname|s", "signal to send to instance (according to the normal unix kill command)", getopt.Optional | getopt.ExampleIsDefault, "SIGTERM"},
	}

	_, _, _, e := optionDefinition.ParseCommandLine()

	os.Args[0] = "visor [opts] instance tail"

	if e != nil {
		exit_code := 0
		description := "send a signal to an instance"

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
