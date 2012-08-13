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

var cmdServiceRegister = &Command{
	Name:      "service-register",
	Short:     "create service in global registry",
	UsageLine: "service-register <name>",
	Long: `
Service-register adds a service to the global registry.
`,
}

func init() {
	cmdServiceRegister.Run = runServiceRegister
}

func runServiceRegister(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Flag.Usage()
	}

	srv := visor.NewService(args[0], cmdServiceRegister.Snapshot)

	_, err := srv.Register()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error registering service %s\n", err.Error())
		os.Exit(2)
	}
}
