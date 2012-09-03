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

var cmdInit = &Command{
	Name:      "init",
	Short:     "initialize coordinator state",
	UsageLine: "scale",
	Long: `
Init takes care of basic setup of the coordinator tree structure.
  `,
}

func init() {
	cmdInit.Run = runInit
}

func runInit(cmd *Command, args []string) {
	_, err := visor.Init(cmdInit.Snapshot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing %s\n", err.Error())
		os.Exit(2)
	}
}
