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

var cmdAppList = &Command{
	Name:      "app-list",
	Short:     "lists apps",
	UsageLine: "app-list",
	Long: `
App-list returns all registered applications.
  `,
}

func init() {
	cmdAppList.Run = runAppList
}

func runAppList(cmd *Command, args []string) {
	s := cmdAppList.Snapshot

	apps, err := visor.Apps(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching apps %s\n", err.Error())
		os.Exit(2)
	}

	for _, app := range apps {
		fmt.Fprintf(os.Stdout, "%s\n", app.Name)
	}
}
