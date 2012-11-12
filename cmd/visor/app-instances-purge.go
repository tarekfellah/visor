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

var cmdAppInstancesPurge = &Command{
	Name:      "app-instances-purge",
	Short:     "purge failed instances",
	UsageLine: "app-instances-purge <app> <rev> [proctype]",
	Long: `
App-instances-purge asks the coordinator to clean-up failed instances.
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
		os.Exit(1)
	}

	switch len(args) {
	case 2:
		ptys, err := app.GetProcTypes()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching proctypes: %s\n", err.Error())
			os.Exit(1)
		}
		for _, pty := range ptys {
			purgeProctypeInstances(pty, revname)
		}
	case 3:
		pty, err := visor.GetProcType(s, app, args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching proctype: %s\n", err.Error())
			os.Exit(1)
		}
		purgeProctypeInstances(pty, revname)
		return
	default:
		fmt.Fprintf(os.Stderr, "Wrong number of arguments")
		os.Exit(1)
	}
}

func purgeProctypeInstances(pty *visor.ProcType, revname *string) {
	ins, err := pty.GetFailedInstances()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching instances for %s: %s\n", pty.Name, err.Error())
		os.Exit(1)
	}

	for _, i := range ins {
		if i.RevisionName == *revname {
			err := i.Unregister()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error removing instance %s: %s\n", i.Name, err.Error())
			}
		}
	}
}
