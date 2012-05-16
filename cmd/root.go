package main

import (
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
)

func Root(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {

	switch subCommand {
	case "init":
		_, err = visor.Init(snapshot())
	}

	return
}
