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

var cmdAppInstances = &Command{
	Name:      "app-instances",
	Short:     "list instances",
	UsageLine: "app-instances <name> [proc]",
	Long: `
App-instances returns instances and there state for an application.
  `,
}

func init() {
	cmdAppInstances.Run = runAppInstances
}

func runAppInstances(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Flag.Usage()
	}

	s := cmdAppInstances.Snapshot

	app, err := visor.GetApp(s, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	ptys, err := app.GetProcTypes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching proctypes %s\n", err.Error())
		os.Exit(2)
	}

	for _, pty := range ptys {
		if len(args) > 1 && string(pty.Name) != args[0] {
			continue
		}

		ins, err := pty.GetInstances()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching instances for %s %s\n", pty.Name, err.Error())
			os.Exit(2)
		}

		for _, i := range ins {
			fmt.Fprintf(os.Stdout, "%s %s %s %s %s %s %d %s\n", i.Name, i.ServiceName, i.AppName, i.ProcessName, i.RevisionName, i.Host, i.Port, i.State)
		}
	}
}
