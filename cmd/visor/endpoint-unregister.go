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

var cmdEndpointUnregister = &Command{
	Name:      "endpoint-unregister",
	Short:     "remove endpoint for service",
	UsageLine: "endpoint-register <service> <addr> <port>",
	Long: `
Endpoint-unregister removes an endpoint for a service in the global registry.
`,
}

func init() {
	cmdEndpointUnregister.Run = runEndpointUnregister
}

func runEndpointUnregister(cmd *Command, args []string) {
	if len(args) < 3 {
		cmd.Flag.Usage()
	}

	srv, err := visor.GetService(cmdEndpointUnregister.Snapshot, args[0])
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

	id := visor.EndpointId(args[1], int(port))
	ep, err := visor.GetEndpoint(cmdEndpointUnregister.Snapshot, srv, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err.Error())
		os.Exit(1)
	}

	err = ep.Unregister()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unregistering endpoint %s\n", err.Error())
		os.Exit(1)
	}
}
