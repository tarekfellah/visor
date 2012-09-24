// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	"fmt"
	"github.com/soundcloud/visor"
	"os"
	"strconv"
)

var cmdEndpointRegister = &Command{
	Name:      "endpoint-register",
	Short:     "create endpoint for service",
	UsageLine: "endpoint-register <service> <addr> <port> [priority] [weight]",
	Long: `
Endpoint-register adds an endpoint to a service in the global registry.
`,
}

func init() {
	cmdEndpointRegister.Run = runEndpointRegister
}

func runEndpointRegister(cmd *Command, args []string) {
	if len(args) < 3 {
		cmd.Flag.Usage()
	}

	srv, err := visor.GetService(cmdEndpointRegister.Snapshot, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching service %s\n", err.Error())
		os.Exit(1)
	}

	if srv == nil {
		fmt.Fprint(os.Stderr, "service could not be found\n")
		os.Exit(1)
	}

	port, err := strconv.ParseInt(args[2], 10, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "port needs to be a positive integer %s\n", err.Error())
		os.Exit(1)
	}

	ep, err := visor.NewEndpoint(srv, args[1], int(port), cmdEndpointRegister.Snapshot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err.Error())
		os.Exit(1)
	}

	if len(args) > 3 {
		prio, err := strconv.ParseInt(args[3], 10, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "priority needs to be a positive integer %s\n", err.Error())
			os.Exit(1)
		}

		ep.Priority = int(prio)
	}

	if len(args) > 4 {
		weight, err := strconv.ParseInt(args[4], 10, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "weight needs to be a positive integer %s\n", err.Error())
			os.Exit(1)
		}

		ep.Weight = int(weight)
	}

	_, err = ep.Register()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error registering endpoint %s\n", err.Error())
		os.Exit(1)
	}
}
