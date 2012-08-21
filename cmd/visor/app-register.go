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

var cmdAppRegister = &Command{
	Name:      "app-register",
	Short:     "create app in global registry",
	UsageLine: "app-register [options] <name>",
	Long: `
App-register adds applications to the global registry.

Options:
  -repo   CVS repository address (https://github.com/foo/<name>)
  -stack  Runtime stack to use (HEAD)
  -type   Application type (lxc)
  `,
}

var r = cmdAppRegister.Flag.String("repo", "", "")
var s = cmdAppRegister.Flag.String("stack", "HEAD", "")
var t = cmdAppRegister.Flag.String("type", "lxc", "")

func init() {
	cmdAppRegister.Run = runAppRegister
}

func runAppRegister(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Flag.Usage()
	}

	name := args[0]
	repo := *r
	stack := *s

	if len(repo) == 0 {
		repo = "https://github.com/foo/" + name
	}

	app := visor.NewApp(name, repo, visor.Stack(stack), cmdAppRegister.Snapshot)
	app.DeployType = *t

	_, err := app.Register()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error registering app %s\n", err.Error())
		os.Exit(2)
	}
}
