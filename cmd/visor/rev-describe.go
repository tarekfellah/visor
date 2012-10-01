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

var cmdRevDescribe = &Command{
	Name:      "rev-describe",
	Short:     "shows info for rev",
	UsageLine: "rev-describe <app> <name>",
	Long: `
Rev-describe returns meta information for the revision.
  `,
}

func init() {
	cmdRevDescribe.Run = runRevDescribe
}

func runRevDescribe(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.Flag.Usage()
	}

	name := args[1]
	s := cmdRevDescribe.Snapshot

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

	fmt.Fprintf(os.Stdout, "name: %s\n", rev.Ref)
	fmt.Fprintf(os.Stdout, "archive-url: %s\n", rev.ArchiveUrl)
}
