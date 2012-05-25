// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
	"net"
)

func Instance(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "exists":
		err = InstanceExists(options, arguments, passThrough)
	case "describe":
		err = InstanceDescribe(options, arguments, passThrough)
	case "tail":
		err = InstanceTail(options, arguments, passThrough)
	case "kill":
		err = InstanceKill(options, arguments, passThrough)
	case "create":
		err = InstanceCreate(arguments[0], arguments[1], arguments[2], arguments[3])
	}
	return
}

func InstanceCreate(appName string, revision string, procType string, addrstr string) (err error) {
	s := snapshot()

	addr, err := net.ResolveTCPAddr("tcp", addrstr)
	if err != nil {
		return
	}

	app, err := visor.GetApp(s, appName)
	if err != nil {
		return
	}

	rev, err := visor.GetRevision(s, app, revision)
	if err != nil {
		return
	}

	proc, err := visor.GetProcType(s, rev, visor.ProcessName(procType))
	if err != nil {
		return
	}

	_, err = (&visor.Instance{Snapshot: s, ProcType: proc, Addr: addr, State: visor.InsStateInitial}).Register()
	return
}

func InstanceExists(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	instanceId := arguments[0]

	print("\ninstance_exists\n")
	print("\n\tinstance id           : " + instanceId)
	print("\n")
	return
}

func InstanceDescribe(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	instanceId := arguments[0]

	print("\ninstance_describe\n")
	print("\n\tinstance id           : " + instanceId)
	print("\n")
	return
}

func InstanceTail(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	instanceId := arguments[0]

	print("\ninstance_tail\n")
	print("\n\tinstance id           : " + instanceId)
	print("\n")
	return
}

func InstanceKill(options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	instanceId := arguments[0]
	signal := options["signal"].String

	print("\ninstance_kill\n")
	print("\n\tinstance id           : " + instanceId)
	print("\n\tsignal                : " + signal)
	print("\n")
	return
}
