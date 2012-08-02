// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"regexp"
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
)

type eventPath int

const (
	pathApp eventPath = iota
	pathRev
	pathProc
	pathIns
	pathInsState
)

var eventPatterns = map[*regexp.Regexp]eventPath{
	regexp.MustCompile("^/apps/([a-zA-Z0-9-]+)/registered$"):                                pathApp,
	regexp.MustCompile("^/apps/([a-zA-Z0-9-]+)/revs/([a-zA-Z0-9-]+)/registered$"):           pathRev,
	regexp.MustCompile("^/apps/([a-zA-Z0-9-]+)/procs/([a-zA-Z0-9-]+)/registered$"):          pathProc,
	regexp.MustCompile("^/apps/([a-zA-Z0-9-]+)/procs/([a-zA-Z0-9-]+)/instances/([0-9-]+)$"): pathIns,
	regexp.MustCompile("^/instances/([0-9-]+)/state$"):                                      pathInsState,
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
		event.Info, err = GetEventInfo(s.FastForward(rev), event)
		if err != nil {
			continue
		}

		listener <- event
	}
	return nil
}

func GetEventInfo(s Snapshot, ev *Event) (info interface{}, err error) {
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

		info, err = GetProcType(s, app, ProcessName(e["proctype"]))
		if err != nil {
			fmt.Printf("error getting proctype: %s\n", err)
		}
	case EvInsReg, EvInsStart, EvInsExit, EvInsFail, EvInsDead:
		var i *Instance

		e := ev.Emitter
		i, err = GetInstance(s, e["instance"])
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

				switch State(src.Body) {
				case InsStateStarted:
					etype = EvInsStart
				case InsStateExited:
					etype = EvInsExit
				case InsStateFailed:
					etype = EvInsFail
				case InsStateDead:
					etype = EvInsDead
				}
			}
			break
		}
	}
	return &Event{etype, emitter, string(src.Body), nil, src, src.Rev}
}
