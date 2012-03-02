package main

import (
	"fmt"
	"os"
	getopt "github.com/kesselborn/go-getopt"
)

func main() {
	optionDefinition := getopt.Options{
		{"name|n", "app's name", getopt.IsArg | getopt.Required, ""},
	}

	_, _, _, e := optionDefinition.ParseCommandLine()

  
  os.Args[0] = "visor [opts] app describe"

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
