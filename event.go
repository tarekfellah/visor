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
	"strings"
)

// An Event represents a change to a file in the registry.
// TODO: turn `Emitter` into own type, instead of map
type Event struct {
	Type    EventType         // Type of event
	Emitter map[string]string // The parsed file path
	Body    string            // Body of the changed file
	Info    interface{}       // Extra information, such as InstanceInfo
	source  *doozer.Event     // Original event returned by doozer
	Rev     int64
}

type EventType string

const (
	EvAppReg    = EventType("app-register")
	EvAppUnreg  = EventType("app-unregister")
	EvRevReg    = EventType("rev-register")
	EvRevUnreg  = EventType("rev-unregister")
	EvProcReg   = EventType("proc-register")
	EvProcUnreg = EventType("proc-unregister")
	EvInsReg    = EventType("instance-register")
	EvInsUnreg  = EventType("instance-unregister")
	EvInsStart  = EventType("instance-start")
	EvInsFail   = EventType("instance-fail")
	EvInsExit   = EventType("instance-exit")
	EvSrvReg    = EventType("service-register")
	EvSrvUnreg  = EventType("service-unregister")
	EvEpReg     = EventType("endpoint-register")
	EvEpUnreg   = EventType("endpoint-unregister")
	EvUnknown   = EventType("UNKNOWN")
)

const (
	doozerGlobPlural = "**"
)

type eventPath int

const (
	pathApp eventPath = iota
	pathRev
	pathProc
	pathIns
	pathInsStatus
	pathInsStart
	pathInsStop
	pathSrv
	pathEp
)

var eventPatterns = map[*regexp.Regexp]eventPath{
	regexp.MustCompile("^/apps/(" + charPat + "+)/registered$"):                          pathApp,
	regexp.MustCompile("^/apps/(" + charPat + "+)/revs/(" + charPat + "+)/registered$"):  pathRev,
	regexp.MustCompile("^/apps/(" + charPat + "+)/procs/(" + charPat + "+)/registered$"): pathProc,
	regexp.MustCompile("^/instances/([-0-9]+)/object$"):                                  pathIns,
	regexp.MustCompile("^/instances/([-0-9]+)/status$"):                                  pathInsStatus,
	regexp.MustCompile("^/instances/([-0-9]+)/start$"):                                   pathInsStart,
	regexp.MustCompile("^/instances/([-0-9]+)/stop$"):                                    pathInsStop,
	regexp.MustCompile("^/services/(" + charPat + "+)/registered$"):                      pathSrv,
	regexp.MustCompile("^/services/(" + charPat + "+)/endpoints/([-0-9]+)$"):             pathEp,
}

func (ev *Event) String() string {
	return fmt.Sprintf("%#v", ev)
}

// WatchEventRaw watches for changes to the registry and sends
// them as *Event objects to the provided channel.
func WatchEventRaw(s Snapshot, listener chan *Event) error {
	rev := s.Rev
	for {
		ev, err := s.conn.Wait(doozerGlobPlural, rev+1)
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
		ev, err := s.conn.Wait(doozerGlobPlural, rev+1)
		if err != nil {
			return err
		}
		rev = ev.Rev
		event := parseEvent(&ev)

		if event.Type == EvUnknown {
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
	emitter := ev.Emitter

	switch ev.Type {
	case EvAppReg:
		info, err = GetApp(s, emitter["app"])

		if err != nil {
			fmt.Printf("error getting app: %s\n", err)
			return
		}
	case EvRevReg:
		var app *App

		app, err = GetApp(s, emitter["app"])
		if err != nil {
			fmt.Printf("error getting app for revision: %s\n", err)
			return
		}

		info, err = GetRevision(s, app, emitter["rev"])
		if err != nil {
			fmt.Printf("error getting revision: %s\n", err)
			return
		}
	case EvProcReg:
		var app *App

		app, err = GetApp(s, emitter["app"])
		if err != nil {
			fmt.Printf("error getting app for proctype: %s\n", err)
			return
		}

		info, err = GetProcType(s, app, emitter["proctype"])
		if err != nil {
			fmt.Printf("error getting proctype: %s\n", err)
		}
	case EvInsReg, EvInsStart, EvInsExit, EvInsFail:
		var i *Instance
		var id int64

		id, err = strconv.ParseInt(emitter["instance"], 10, 64)
		if err != nil {
			return
		}

		i, err = GetInstance(s, id)
		if err != nil {
			fmt.Printf("error getting instance: %s\n", err)
			return
		}
		// XXX Find better place for this
		emitter["app"] = i.AppName
		emitter["rev"] = i.RevisionName
		emitter["proctype"] = string(i.ProcessName)

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

		srv, err = GetService(s, emitter["service"])
		if err != nil {
			fmt.Printf("error getting service for endpoint: %s\n", err)
			return
		}

		info, err = GetEndpoint(s, srv, emitter["endpoint"])
		if err != nil {
			fmt.Printf("error getting endpoint: %s\n", err)
		}
	}
	return
}

func parseEvent(src *doozer.Event) *Event {
	path := src.Path

	etype := EvUnknown
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
				emitter["instance"] = match[1]

				if src.IsSet() {
					fields := strings.Fields(string(src.Body))
					emitter["app"] = fields[0]
					emitter["rev"] = fields[1]
					emitter["proctype"] = fields[2]
					etype = EvInsReg
				} else if src.IsDel() {
					etype = EvInsUnreg
				}
			case pathInsStart:
				emitter["instance"] = match[1]

				if !src.IsSet() {
					break
				}
				body := string(src.Body)
				if body == "" {
					etype = EvInsStart
				} else {
					fields := strings.Fields(body)
					if len(fields) > 1 {
						etype = EvInsStart
					}
				}
			case pathInsStatus:
				emitter["instance"] = match[1]

				if !src.IsSet() {
					break
				}

				switch InsStatus(src.Body) {
				case InsStatusRunning:
					etype = EvInsStart
				case InsStatusExited:
					etype = EvInsExit
				case InsStatusFailed:
					etype = EvInsFail
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

	return &Event{
		Type:    etype,
		Emitter: emitter,
		Body:    string(src.Body),
		Info:    nil,
		source:  src,
		Rev:     src.Rev,
	}
}
