// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	"fmt"
	"github.com/soundcloud/visor"
	"os"
)

var cmdAppUnregister = &Command{
	Name:      "app-unregister",
	Short:     "remove app from the global registry",
	UsageLine: "app-unregister <name>",
	Long: `
App-unregister removes applications from the globa registry.
  `,
}

func init() {
	cmdAppUnregister.Run = runAppUnregister
}

func runAppUnregister(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Flag.Usage()
	}

	name := args[0]
	s := cmdAppUnregister.Snapshot

	app, err := visor.GetApp(s, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	err = app.Unregister()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unregistering app %s\n", err.Error())
		os.Exit(2)
	}
}
