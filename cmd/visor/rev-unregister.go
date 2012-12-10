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

var cmdRevUnregister = &Command{
	Name:      "rev-unregister",
	Short:     "removes revision",
	UsageLine: "rev-unregister <app> <name>",
	Long: `
Rev-unregister removes a revsion from an application.
  `,
}

func init() {
	cmdRevUnregister.Run = runRevUnregister
}

func runRevUnregister(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.Flag.Usage()
	}

	name := args[1]
	s := cmdRevUnregister.Snapshot

	app, err := visor.GetApp(s, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(1)
	}

	rev, err := visor.GetRevision(s, app, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching rev %s\n", err.Error())
		os.Exit(1)
	}

	err = rev.Purge()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unregistering rev %s\n", err.Error())
		os.Exit(1)
	}
}
