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

var cmdAppEnvGet = &Command{
	Name:      "app-env-get",
	Short:     "retrieve environment",
	UsageLine: "app-env-get <app> [key]",
	Long: `
App-env-get returns the whole or filtered environment for an application.
  `,
}

func init() {
	cmdAppEnvGet.Run = runAppEnvGet
}

func runAppEnvGet(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Flag.Usage()
	}

	s := cmdAppEnvGet.Snapshot
	name := args[0]

	app, err := visor.GetApp(s, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(1)
	}

	if len(args) == 2 {
		key := args[1]
		val, err := app.GetEnvironmentVar(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching app env %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "%s=%s\n", key, val)
	} else {
		env, err := app.EnvironmentVars()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching app env %s\n", err.Error())
			os.Exit(1)
		}

		for key, val := range env {
			fmt.Fprintf(os.Stdout, "%s=%s\n", key, val)
		}
	}
}
