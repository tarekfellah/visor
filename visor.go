// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

// Visor is a library which provides an abstract interface
// over a global process state.
//
// This process state is referred to as the registry.
//
// Example usage:
//
//     package main
//
//     import "soundcloud/visor"
//
//     func main() {
//         client, err := visor.Dial("coordinator:8046", "/", new(visor.StringCodec))
//         if err != nil {
//           panic(err)
//         }
//
//         l := make(chan *visor.Event)
//
//         // Watch for changes in the global process state
//         go visor.WatchEvent(client.Snapshot, l)
//
//         for {
//             fmt.Println(<-l)
//         }
//     }
//
package visor

import (
	"errors"
	"path"
	"strconv"
)

const DEFAULT_URI string = "doozer:?ca=localhost:8046"
const DEFAULT_ADDR string = "localhost:8046"
const DEFAULT_ROOT string = "/visor"
const SCALE_PATH string = "scale"
const START_PORT int = 8000
const START_PORT_PATH string = "/next-port"

type ProcessName string
type Stack string
type State string

func Init(s Snapshot) (rev int64, err error) {
	exists, _, err := s.Conn().Exists(START_PORT_PATH)
	if err != nil {
		return
	}

	if !exists {
		rev, err = s.Conn().Set(START_PORT_PATH, s.Rev, []byte(strconv.Itoa(START_PORT)))
		if err != nil {
			return
		}

		return rev, err
	}
	return s.conn.Rev()
}

func Scale(app string, revision string, processName string, factor int, s Snapshot) (err error) {
	if factor < 0 {
		return errors.New("scaling factor needs to be a positive integer")
	}

	p := path.Join(APPS_PATH, app, REVS_PATH, revision, SCALE_PATH, processName)
	op := OpStart
	tickets := factor

	res, frev, err := s.conn.Get(p, nil)
	if err != nil && frev != 0 {
		return
	}

	// Check response isn't empty
	if len(res) > 0 {
		var current int

		current, err = strconv.Atoi(string(res))
		if err != nil {
			return
		}

		tickets = factor - current

		if tickets < 0 {
			op = OpStop
			tickets = -tickets
		}
	}

	rev, err := s.conn.Set(p, s.Rev, []byte(strconv.Itoa(factor)))
	if err != nil {
		return
	}

	s = s.FastForward(rev)

	for i := 0; i < tickets; i++ {
		var ticket *Ticket

		ticket, err = CreateTicket(app, revision, ProcessName(processName), op, s)
		if err != nil {
			return
		}

		s = s.FastForward(ticket.Rev)
	}

	return
}
