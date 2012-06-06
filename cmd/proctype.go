// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

// TODO: use return-often style

import (
	"errors"
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
)

func ProcType(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "register":
		err = ProcTypeRegister(arguments[0], arguments[1])
	}
	return
}

func ProcTypeRegister(appName string, proctype string) (err error) {
	s := snapshot()
	app, err := visor.GetApp(s, appName)

	if err == nil {
		pty := visor.NewProcType(app, visor.ProcessName(proctype), s)
		_, err = pty.Register()
		if err == visor.ErrKeyConflict {
			err = errors.New("Proctype '" + proctype + "' for app '" + appName + "' already registered!")
		}
	}

	return
}
