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

var cmdAppDescribe = &Command{
	Name:      "app-describe",
	Short:     "app-describe in global registry",
	UsageLine: "app-describe <name>",
	Long: `
App-describe returns meta information for the appliation given.
  `,
}

func init() {
	cmdAppDescribe.Run = runAppDescribe
}

func runAppDescribe(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Flag.Usage()
	}

	s := cmdAppDescribe.Snapshot

	app, err := visor.GetApp(s, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	fmt.Fprintf(os.Stdout, "name: %s\n", app.Name)
	fmt.Fprintf(os.Stdout, "type: %s\n", app.DeployType)
	fmt.Fprintf(os.Stdout, "repo: %s\n", app.RepoUrl)
	fmt.Fprintf(os.Stdout, "stack: %s\n", app.Stack)
}
