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

var cmdRevRegister = &Command{
	Name:      "rev-register",
	Short:     "create revision",
	UsageLine: "rev-register <app> <name> <artifact-url>",
	Long: `
Rev-register adds a new named revision to an application.
  `,
}

func init() {
	cmdRevRegister.Run = runRevRegister
}

func runRevRegister(cmd *Command, args []string) {
	if len(args) < 3 {
		cmd.Flag.Usage()
	}

	name := args[1]
	url := args[2]
	s := cmdRevRegister.Snapshot

	app, err := visor.GetApp(s, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	rev := visor.NewRevision(app, name, s)

	rev.ArchiveUrl = url

	_, err = rev.Register()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error registering rev %s\n", err.Error())
		os.Exit(2)
	}
}
