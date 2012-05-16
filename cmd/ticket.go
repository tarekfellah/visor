// Copyright (c) 2012, SoundCloud Ltd., Alexis Sellier, Alexander Simmerl, Daniel Bornkessel, Tomás Senart
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	"errors"
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"

//"net"
//"strconv"
)

func Ticket(subCommand string, options map[string]getopt.OptionValue, arguments []string, passThrough []string) (err error) {
	switch subCommand {
	case "create":
		err = TicketCreate(arguments[0], arguments[1], arguments[2], arguments[3])
	}
	return
}

func TicketCreate(app string, revision string, proc string, op string) (err error) {
	if operation := visor.NewOperationType(op); operation != visor.OpInvalid {
		visor.CreateTicket(app, revision, visor.ProcessName(proc), operation, snapshot())
	} else {
		err = errors.New("invalid op type: " + op)
	}

	return
}
