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

var cmdAppEnvSet = &Command{
	Name:      "app-env-set",
	Short:     "app-env-set for app",
	UsageLine: "app-env-set <app> <key> <value>",
	Long: `
App-env-set stores a value for the given key in the application environment.
  `,
}

func init() {
	cmdAppEnvSet.Run = runAppEnvSet
}

func runAppEnvSet(cmd *Command, args []string) {
	if len(args) < 3 {
		cmd.Flag.Usage()
	}

	s := cmdAppEnvSet.Snapshot
	name := args[0]
	key := args[1]
	val := args[2]

	app, err := visor.GetApp(s, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	_, err = app.SetEnvironmentVar(key, val)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting env var %s\n", err.Error())
		os.Exit(2)
	}
}
