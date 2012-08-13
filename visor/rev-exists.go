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

var cmdRevExists = &Command{
	Name:      "rev-exists",
	Short:     "checks if rev exists",
	UsageLine: "rev-exists <app> <name>",
	Long: `
Rev-exists tells of a revision exists for the given application.
  `,
}

func init() {
	cmdRevExists.Run = runRevExists
}

func runRevExists(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.Flag.Usage()
	}

	name := args[1]
	s := cmdRevExists.Snapshot

	app, err := visor.GetApp(s, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	_, err = visor.GetRevision(s, app, name)
	if err != nil {
		os.Exit(-1)
	}
}
