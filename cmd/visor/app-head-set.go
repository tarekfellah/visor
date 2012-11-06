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

var cmdAppHeadSet = &Command{
	Name:      "app-head-set",
	Short:     "set the app head",
	UsageLine: "app-head-set <app> <head>",
	Long: `
App-head-set sets the latest revision of the app to <head>
  `,
}

func init() {
	cmdAppHeadSet.Run = runAppHeadSet
}

func runAppHeadSet(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.Flag.Usage()
	}

	s := cmdAppHeadSet.Snapshot
	name := args[0]
	head := args[1]

	app, err := visor.GetApp(s, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(1)
	}

	_, err = app.SetHead(head)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting app head %s\n", err.Error())
		os.Exit(1)
	}
}
