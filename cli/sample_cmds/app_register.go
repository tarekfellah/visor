package main

import (
	"fmt"
	"os"
	getopt "github.com/kesselborn/go-getopt"
)

func main() {
	optionDefinition := getopt.Options{
		{"name|n", "app's name", getopt.IsArg | getopt.Required, ""},
		{"deploytype|t", "deploy type (one of mount, bazapta, lxc)", getopt.Optional | getopt.ExampleIsDefault, "lxc"},
    {"repourl|r", "repository url of this app", getopt.Optional | getopt.ExampleIsDefault, "http://github.com/soundcloud/<name>"},
    {"stack|s", "stack version this app should be pinned to -- ommit if you always want the latest stack", getopt.Optional, ""},
    {"irc|i", "comma separated list of irc channels where a deploy should be announced", getopt.Optional | getopt.ExampleIsDefault, "deploys"},
	}

	_, _, _, e := optionDefinition.ParseCommandLine()

  
  os.Args[0] = "visor [opts] app register"

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
