// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	"fmt"
	"github.com/soundcloud/visor"
	"log"
	"os"
	"strconv"
)

var cmdScale = &Command{
	Name:      "scale",
	Short:     "scale proctypes",
	UsageLine: "scale [app] [proctype] [rev] [factor]",
	Long: `
Scale scales a proctype at a specific revision to the set factor.
  `,
}

func init() {
	cmdScale.Run = runScale
}

func runScale(cmd *Command, args []string) {
	if len(args) < 4 {
		cmd.Flag.Usage()
	}

	f, err := strconv.Atoi(string(args[3]))
	if err != nil {
		fmt.Fprint(os.Stderr, "Error 'factor' needs to an integer\n")
		os.Exit(2)
	}

	err = visor.Scale(args[0], args[1], args[2], f, cmdScale.Snapshot)
	if err != nil {
		log.Fatal(err)
	}
}
