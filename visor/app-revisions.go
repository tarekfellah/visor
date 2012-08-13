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

var cmdAppRevisions = &Command{
	Name:      "app-revisions",
	Short:     "list revisions for an app",
	UsageLine: "app-revisions <name>",
	Long: `
App-services returns services and meta information for an application.
  `,
}

func init() {
	cmdAppRevisions.Run = runAppRevisions
}

func runAppRevisions(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Flag.Usage()
	}

	s := cmdAppRevisions.Snapshot

	app, err := visor.GetApp(s, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	revs, err := visor.AppRevisions(s, app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching revisions %s\n", err.Error())
		os.Exit(2)
	}

	for _, rev := range revs {
		fmt.Fprintf(os.Stdout, "%s %s %s\n", rev.App.Name, rev.Ref, rev.ArchiveUrl)
	}
}
