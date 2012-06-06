// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
	"strconv"
)

func Root(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {

	switch subCommand {
	case "init":
		_, err = visor.Init(snapshot())
	case "scale":
		var f int

		s := snapshot()
		f, err = strconv.Atoi(string(arguments[3]))
		if err != nil {
			return
		}

		err = visor.Scale(arguments[0], arguments[1], arguments[2], f, s)
	}

	return
}
