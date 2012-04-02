package main

import (
	"errors"
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
	"net"
	"strconv"
)

func Instance(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "describe":
		err = InstanceDescribe(options, arguments, passThrough)
	case "tail":
		err = InstanceTail(options, arguments, passThrough)
	case "kill":
		err = InstanceKill(options, arguments, passThrough)
	case "create":
		if len(arguments) > 4 {
			err = InstanceCreate(arguments[0], arguments[1], arguments[2], arguments[3], arguments[4])
		} else {
			err = InstanceCreate(arguments[0], arguments[1], arguments[2], arguments[3], "")
		}

	}
	return
}

func InstanceCreate(appName string, revision string, procType string, ipStr string, portStr string) (err error) {
	snapshot := snapshot()
	var app *visor.App

	var ip net.IP
	if ip = net.ParseIP(ipStr); ip == nil {
		err = errors.New(ipStr + " is not a valid ip addres (hostnames are not allowed")
		return
	}

	var port int
	if port, err = strconv.Atoi(portStr); err != nil {
		return
	}

	if app, err = visor.GetApp(snapshot, appName); err == nil {

		var rev *visor.Revision
		if rev, err = visor.GetRevision(snapshot, app, revision); err == nil {
			proc := &visor.ProcType{Snapshot: snapshot, Revision: rev, Name: visor.ProcessName(procType)}
			_, err = (&visor.Instance{Snapshot: snapshot, ProcType: proc, Addr: &net.TCPAddr{IP: ip, Port: port}, State: visor.State(0)}).Register()
		}
	}

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
