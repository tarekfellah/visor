package main

import (
	"fmt"
	"os"
	getopt "github.com/kesselborn/go-getopt"
)

func main() {
  //  visor [-c config] -s <server> [-p <port>] [-r <root>]
	optionDefinition := getopt.Options{
		{"config|c", "config file", getopt.IsConfigFile | getopt.ExampleIsDefault, "/etc/visor.conf"},
		{"server|s", "doozer server", getopt.Required, ""},
		{"port|p", "port the doozer server is listening on", getopt.Optional | getopt.ExampleIsDefault, "8046"},
    {"root|r", "visor namespace within doozer: all entries will be prepended with this path", getopt.Optional | getopt.ExampleIsDefault, "/visor"},
    {"component", "component scope a command should be executed on; call 'visor help' for an overview", getopt.IsArg | getopt.Required, ""},
    {"command", "command to execute", getopt.IsArg | getopt.Required, ""},
    {"...", "command's arguments and options", getopt.IsArg | getopt.Optional, ""},
	}

	_, _, _, e := optionDefinition.ParseCommandLine()

  os.Args[0] = "visor"

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
