// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"regexp"
	"strconv"
)

// An Event represents a change to a file in the registry.
type Event struct {
	Type    EventType         // Type of event
	Emitter map[string]string // The parsed file path
	Body    string            // Body of the changed file
	Info    interface{}       // Extra information, such as InstanceInfo
	source  *doozer.Event     // Original event returned by doozer
	Rev     int64
}

type EventType int

var EventTypes = map[EventType]string{
	EvAppReg:    "app-register",
	EvAppUnreg:  "app-unregister",
	EvRevReg:    "rev-register",
	EvRevUnreg:  "rev-unregister",
	EvProcReg:   "proc-register",
	EvProcUnreg: "proc-unregister",
	EvInsReg:    "instance-register",
	EvInsUnreg:  "instance-unregister",
	EvInsStart:  "instance-start",
	EvInsFail:   "instance-fail",
	EvInsExit:   "instance-exit",
	EvInsDead:   "instance-dead",
	EvSrvReg:    "service-register",
	EvSrvUnreg:  "service-unregister",
	EvEpReg:     "endpoint-register",
	EvEpUnreg:   "endpoint-unregister",
}

func (e EventType) String() string {
	str, ok := EventTypes[e]
	if !ok {
		return "<unknown event>"
	}
	return str
}

// Event types
const (
	EvAppReg    EventType = iota // App register
	EvAppUnreg                   // App unregister
	EvRevReg                     // Revision register
	EvRevUnreg                   // Revision unregister
	EvProcReg                    // ProcType register
	EvProcUnreg                  // ProcType unregister
	EvInsReg                     // Instance register
	EvInsUnreg                   // Instance unregister
	EvInsStart                   // Instance state changed to 'started'
	EvInsFail                    // Instance state changed to 'failed'
	EvInsDead                    // Instance state changed to 'dead'
	EvInsExit                    // Instance state changed to 'exited'
	EvSrvReg                     // Service register
	EvSrvUnreg                   // Service unregister
	EvEpReg                      // Endpoint register
	EvEpUnreg                    // Endpoint unregister
)

type eventPath int

const (
	pathApp eventPath = iota
	pathRev
	pathProc
	pathIns
	pathInsState
	pathSrv
	pathEp
)

var eventPatterns = map[*regexp.Regexp]eventPath{
	regexp.MustCompile("^/apps/([a-zA-Z0-9-]+)/registered$"):                                pathApp,
	regexp.MustCompile("^/apps/([a-zA-Z0-9-]+)/revs/([a-zA-Z0-9-]+)/registered$"):           pathRev,
	regexp.MustCompile("^/apps/([a-zA-Z0-9-]+)/procs/([a-zA-Z0-9-]+)/registered$"):          pathProc,
	regexp.MustCompile("^/apps/([a-zA-Z0-9-]+)/procs/([a-zA-Z0-9-]+)/instances/([0-9-]+)$"): pathIns,
	regexp.MustCompile("^/instances/([0-9-]+)/state$"):                                      pathInsState,
	regexp.MustCompile("^/services/([a-zA-Z0-9-]+)/registered$"):                            pathSrv,
	regexp.MustCompile("^/services/([a-zA-Z0-9-]+)/endpoints/([0-9-]+)$"):                   pathEp,
}

func (ev *Event) String() string {
	return fmt.Sprintf("%#v", ev)
}

// WatchEventRaw watches for changes to the registry and sends
// them as *Event objects to the provided channel.
func WatchEventRaw(s Snapshot, listener chan *Event) error {
	rev := s.Rev
	for {
		ev, err := s.conn.Wait("**", rev+1)
		if err != nil {
			return err
		}
		rev = ev.Rev
		event := parseEvent(&ev)

		listener <- event
	}
	return nil
}

// WatchEvent wraps WatchEventRaw with additional information.
func WatchEvent(s Snapshot, listener chan *Event) error {
	rev := s.Rev
	for {
		ev, err := s.conn.Wait("**", rev+1)
		if err != nil {
			return err
		}
		rev = ev.Rev
		event := parseEvent(&ev)
		if event.Type == -1 {
			continue
		}
		event.Info, err = getEventInfo(s.FastForward(rev), event)
		if err != nil {
			continue
		}

		listener <- event
	}
	return nil
}

func getEventInfo(s Snapshot, ev *Event) (info interface{}, err error) {
	switch ev.Type {
	case EvAppReg:
		e := ev.Emitter
		info, err = GetApp(s, e["app"])

		if err != nil {
			fmt.Printf("error getting app: %s\n", err)
			return
		}
	case EvRevReg:
		var app *App

		e := ev.Emitter
		app, err = GetApp(s, e["app"])
		if err != nil {
			fmt.Printf("error getting app for revision: %s\n", err)
			return
		}

		info, err = GetRevision(s, app, e["rev"])
		if err != nil {
			fmt.Printf("error getting revision: %s\n", err)
			return
		}
	case EvProcReg:
		var app *App

		e := ev.Emitter
		app, err = GetApp(s, e["app"])
		if err != nil {
			fmt.Printf("error getting app for proctype: %s\n", err)
			return
		}

		info, err = GetProcType(s, app, e["proctype"])
		if err != nil {
			fmt.Printf("error getting proctype: %s\n", err)
		}
	case EvInsReg, EvInsStart, EvInsExit, EvInsFail, EvInsDead:
		var i *Instance
		var id int64

		e := ev.Emitter

		id, err = strconv.ParseInt(e["instance"], 10, 64)
		if err != nil {
			return
		}

		i, err = GetInstance(s, id)
		if err != nil {
			fmt.Printf("error getting instance: %s\n", err)
			return
		}
		// XXX Find better place for this
		e["app"] = i.AppName
		e["rev"] = i.RevisionName
		e["proctype"] = string(i.ProcessName)

		info = i

		if err != nil {
			fmt.Printf("error getting instance info: %s\n", err)
			return
		}
	case EvSrvReg:
		info, err = GetService(s, ev.Emitter["service"])

		if err != nil {
			fmt.Printf("error getting service: %s\n", err)
			return
		}
	case EvEpReg:
		var srv *Service

		e := ev.Emitter
		srv, err = GetService(s, e["service"])
		if err != nil {
			fmt.Printf("error getting service for endpoint: %s\n", err)
			return
		}

		info, err = GetEndpoint(s, srv, e["endpoint"])
		if err != nil {
			fmt.Printf("error getting endpoint: %s\n", err)
		}
	}
	return
}

func parseEvent(src *doozer.Event) *Event {
	path := src.Path

	etype := EventType(-1)
	emitter := map[string]string{}

	for re, ev := range eventPatterns {
		if match := re.FindStringSubmatch(path); match != nil {
			switch ev {
			case pathApp:
				emitter["app"] = match[1]

				if src.IsSet() {
					etype = EvAppReg
				} else if src.IsDel() {
					etype = EvAppUnreg
				}
			case pathRev:
				emitter["app"] = match[1]
				emitter["rev"] = match[2]

				if src.IsSet() {
					etype = EvRevReg
				} else if src.IsDel() {
					etype = EvRevUnreg
				}
			case pathProc:
				emitter["app"] = match[1]
				emitter["proctype"] = match[2]

				if src.IsSet() {
					etype = EvProcReg
				} else if src.IsDel() {
					etype = EvProcUnreg
				}
			case pathIns:
				emitter["app"] = match[1]
				emitter["proctype"] = match[2]
				emitter["instance"] = match[3]

				if src.IsSet() {
					etype = EvInsReg
				} else if src.IsDel() {
					etype = EvInsUnreg
				}
			case pathInsState:
				emitter["instance"] = match[1]

				if !src.IsSet() {
					break
				}

				switch InsStatus(src.Body) {
				case InsStatusStarted:
					etype = EvInsStart
				case InsStatusExited:
					etype = EvInsExit
				case InsStatusFailed:
					etype = EvInsFail
				case InsStatusDead:
					etype = EvInsDead
				}
			case pathSrv:
				emitter["service"] = match[1]

				if src.IsSet() {
					etype = EvSrvReg
				} else if src.IsDel() {
					etype = EvSrvUnreg
				}
			case pathEp:
				emitter["service"] = match[1]
				emitter["endpoint"] = match[2]

				if src.IsSet() {
					etype = EvEpReg
				} else if src.IsDel() {
					etype = EvEpUnreg
				}
			}
			break
		}
	}
	return &Event{etype, emitter, string(src.Body), nil, src, src.Rev}
}
