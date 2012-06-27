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

var cmdProcUnregister = &Command{
	Name:      "proc-unregister",
	Short:     "proc-unregister from app",
	UsageLine: "proc-unregister <app> <name>",
	Long: `
Proc-unregister removes a proctype from an application.
  `,
}

func init() {
	cmdProcUnregister.Run = runProcUnregister
}

func runProcUnregister(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Flag.Usage()
	}

	name := visor.ProcessName(args[1])
	s := cmdProcUnregister.Snapshot

	app, err := visor.GetApp(s, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	proc, err := visor.GetProcType(s, app, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching proc %s\n", err.Error())
		os.Exit(2)
	}

	err = proc.Unregister()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unregistering proc %s\n", err.Error())
		os.Exit(2)
	}
}
