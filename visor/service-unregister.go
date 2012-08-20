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

var cmdServiceUnregister = &Command{
	Name:      "service-unregister",
	Short:     "remove service from the global registry",
	UsageLine: "service-unregister <name>",
	Long: `
Service-unregister removes the service from the globa registry.
  `,
}

func init() {
	cmdServiceUnregister.Run = runServiceUnregister
}

func runServiceUnregister(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Flag.Usage()
	}

	srv, err := visor.GetService(cmdServiceUnregister.Snapshot, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching service %s\n", err.Error())
		os.Exit(2)
	}

	if srv == nil {
		fmt.Fprint(os.Stderr, "service could not be found\n")
		os.Exit(2)
	}

	err = srv.Unregister()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unregistering service %s\n", err.Error())
		os.Exit(2)
	}
}
