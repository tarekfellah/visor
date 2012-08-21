// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	"../../"
	"fmt"
	"os"
)

var cmdAppInstancesPurge = &Command{
	Name:      "app-instances-purge",
	Short:     "purge dead instances",
	UsageLine: "app-instances-purge <app> <rev> [proctype]",
	Long: `
App-instances-purge asks the coordinator to clean-up dead instances.
  `,
}

func init() {
	cmdAppInstancesPurge.Run = runAppInstancesPurge
}

func runAppInstancesPurge(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.Flag.Usage()
	}
	s := cmdAppInstancesPurge.Snapshot

	appname := &args[0]
	revname := &args[1]

	app, err := visor.GetApp(s, *appname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app: %s\n", err.Error())
		os.Exit(2)
	}

	ptys, err := app.GetProcTypes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching proctypes: %s\n", err.Error())
		os.Exit(2)
	}

	for _, pty := range ptys {
		if len(args) >= 3 && string(pty.Name) != args[2] {
			continue
		}

		ins, err := pty.GetInstances()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching instances for %s: %s\n", pty.Name, err.Error())
			os.Exit(2)
		}

		for _, i := range ins {
			if i.State == visor.InsStateDead && i.RevisionName == *revname {
				err := i.Unregister()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error removing instance %s: %s\n", i.Name, err.Error())
				}
			}
		}
	}
}
