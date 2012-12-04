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
		os.Exit(1)
	}

	ptys, err := app.GetProcTypes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching proctypes %s\n", err.Error())
		os.Exit(1)
	}

	total := len(ptys)
	insch := make(chan []*visor.Instance, total)

	for _, pty := range ptys {
		if len(args) >= 2 && string(pty.Name) != args[1] {
			total--
			continue
		}
		go func(pty *visor.ProcType) {
			ins, err := pty.GetInstances()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching instances for %s %s\n", pty.Name, err)
				os.Exit(1)
			} else {
				insch <- ins
			}
		}(pty)
	}

	for i := 0; i < total; i++ {
		for _, i := range <-insch {
			fmt.Fprintf(os.Stdout, "%s %s %s %s %s %s %d %s\n", i.Dir.Name, i.ServiceName(), i.AppName, i.ProcessName, i.RevisionName, i.Host, i.Port, i.Status)
		}
	}
}
