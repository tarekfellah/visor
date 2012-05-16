// Copyright (c) 2012, SoundCloud Ltd., Alexis Sellier, Alexander Simmerl, Daniel Bornkessel, Tomás Senart
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
)

func Root(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {

	switch subCommand {
	case "init":
		_, err = visor.Init(snapshot())
	}

	return
}
