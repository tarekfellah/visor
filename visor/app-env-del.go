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

var cmdAppEnvDel = &Command{
	Name:      "app-env-del",
	Short:     "app-env-del for app",
	UsageLine: "app-env-del <app> <key>",
	Long: `
App-env-del removes a value for the given key in the application environment.
  `,
}

func init() {
	cmdAppEnvDel.Run = runAppEnvDel
}

func runAppEnvDel(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.Flag.Usage()
	}

	s := cmdAppEnvDel.Snapshot
	name := args[0]
	key := args[1]

	app, err := visor.GetApp(s, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	_, err = app.DelEnvironmentVar(key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing env var %s\n", err.Error())
		os.Exit(2)
	}
}
