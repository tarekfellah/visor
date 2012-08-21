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

var cmdProcRegister = &Command{
	Name:      "proc-register",
	Short:     "create proctype",
	UsageLine: "proc-register <app> <name>",
	Long: `
Proc-register adds a new named proctype to an application.
  `,
}

func init() {
	cmdProcRegister.Run = runProcRegister
}

func runProcRegister(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.Flag.Usage()
	}

	name := visor.ProcessName(args[1])
	s := cmdProcRegister.Snapshot

	app, err := visor.GetApp(s, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	proc := visor.NewProcType(app, name, s)

	_, err = proc.Register()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error registering proc %s\n", err.Error())
		os.Exit(2)
	}
}
