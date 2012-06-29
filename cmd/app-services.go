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

var cmdAppServices = &Command{
	Name:      "app-services",
	Short:     "app-services for app",
	UsageLine: "app-services <name>",
	Long: `
App-services returns services and meta information for an application.
  `,
}

func init() {
	cmdAppServices.Run = runAppServices
}

func runAppServices(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Flag.Usage()
	}

	s := cmdAppServices.Snapshot

	app, err := visor.GetApp(s, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching app %s\n", err.Error())
		os.Exit(2)
	}

	proxies, err := app.Conn().Getdir("/proxies", app.Snapshot.Rev)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching proxy addresses %s\n", err.Error())
		os.Exit(2)
	}

	proxy := ""
	if len(proxies) > 0 {
		proxy = proxies[0]
	}

	ptys, err := app.GetProcTypes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching proctypes %s\n", err.Error())
		os.Exit(2)
	}

	for _, pty := range ptys {
		ins, err := pty.GetInstanceNames()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching instances for %s %s\n", pty.Name, err.Error())
			os.Exit(2)
		}

		fmt.Fprintf(os.Stdout, "%s-%s %s %d %d\n", pty.App.Name, pty.Name, proxy, pty.Port, len(ins))
	}
}
