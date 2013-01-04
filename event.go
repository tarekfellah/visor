// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type EventData struct {
	App      *string
	Endpoint *string
	Instance *string
	Proctype *string
	Revision *string
	Service  *string
}

func (d EventData) String() string {
	fields := []string{}
	t := reflect.TypeOf(d)

	for i := 0; i < t.NumField(); i++ {
		v := reflect.ValueOf(d).Field(i)

		if !v.IsNil() {
			fields = append(fields, fmt.Sprintf("%s: %v", t.Field(i).Name, v.Elem().Interface()))
		}
	}

	return fmt.Sprintf("EventData{%s}", strings.Join(fields, ", "))
}

// An Event represents a change to a file in the registry.
type Event struct {
	Type   EventType // Type of event
	Body   string    // Body of the changed file
	Source snapshotable
	Path   EventData
	raw    *doozer.Event // Original event returned by doozer
	Rev    int64
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

		event, err := enrichEvent(s.FastForward(rev), &ev)
		if err != nil {
			return err
		}

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
		event, err := enrichEvent(s.FastForward(rev), &ev)
		if err != nil {
			return err
		}

		if event.Type == EvUnknown {
			continue
		}

		listener <- event
	}
	return nil
}

func canonicalizeMetadata(s Snapshot, etype EventType, uncanonicalized EventData) (source snapshotable, err error) {
	var (
		app *App
		rev *Revision
		pty *ProcType
		ins *Instance
		srv *Service
		edp *Endpoint
	)

	if uncanonicalized.App != nil {
		app, err = GetApp(s, *uncanonicalized.App)

		if err != nil {
			return
		}
	}

	if uncanonicalized.Revision != nil {
		rev, err = GetRevision(s, app, *uncanonicalized.Revision)

		if err != nil {
			return
		}
	}

	if uncanonicalized.Proctype != nil {
		pty, err = GetProcType(s, app, *uncanonicalized.Proctype)
		if err != nil {
			return
		}
	}

	if uncanonicalized.Instance != nil {
		var id int64 = -1
		if id, err = strconv.ParseInt(*uncanonicalized.Instance, 10, 64); err != nil {
			return
		}
		if ins, err = GetInstance(s, id); err != nil {
			return
		}
	}

	if uncanonicalized.Service != nil {
		srv, err = GetService(s, *uncanonicalized.Service)
		if err != nil {
			return
		}

	}

	if uncanonicalized.Endpoint != nil {
		edp, err = GetEndpoint(s, srv, *uncanonicalized.Endpoint)
		if err != nil {
			return
		}
	}

	switch etype {
	case EvAppReg:
		source = app
	case EvRevReg:
		source = rev
	case EvProcReg:
		source = pty
	case EvInsReg, EvInsStart, EvInsFail, EvInsExit:
		source = ins
	case EvSrvReg:
		source = srv
	case EvEpReg:
		source = edp
	}

	return
}

func enrichEvent(s Snapshot, src *doozer.Event) (event *Event, err error) {
	var canonicalized snapshotable

	path := src.Path
	etype := EvUnknown
	uncanonicalized := EventData{}

	for re, ev := range eventPatterns {
		if match := re.FindStringSubmatch(path); match != nil {
			switch ev {
			case pathApp:
				uncanonicalized.App = &match[1]

				if src.IsSet() {
					etype = EvAppReg
				} else if src.IsDel() {
					etype = EvAppUnreg
				}
			case pathRev:
				uncanonicalized.App = &match[1]
				uncanonicalized.Revision = &match[2]

				if src.IsSet() {
					etype = EvRevReg
				} else if src.IsDel() {
					etype = EvRevUnreg
				}
			case pathProc:
				uncanonicalized.App = &match[1]
				uncanonicalized.Proctype = &match[2]

				if src.IsSet() {
					etype = EvProcReg
				} else if src.IsDel() {
					etype = EvProcUnreg
				}
			case pathIns:
				uncanonicalized.Instance = &match[1]

				if src.IsSet() {
					etype = EvInsReg
				} else if src.IsDel() {
					etype = EvInsUnreg
				}
			case pathInsStart:
				uncanonicalized.Instance = &match[1]

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
				uncanonicalized.Instance = &match[1]

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
				uncanonicalized.Service = &match[1]

				if src.IsSet() {
					etype = EvSrvReg
				} else if src.IsDel() {
					etype = EvSrvUnreg
				}
			case pathEp:
				uncanonicalized.Service = &match[1]
				uncanonicalized.Endpoint = &match[2]
				if src.IsSet() {
					etype = EvEpReg
				} else if src.IsDel() {
					etype = EvEpUnreg
				}
			}
			break
		}
	}

	if src.IsSet() {
		canonicalized, err = canonicalizeMetadata(s, etype, uncanonicalized)
		if err != nil {
			fmt.Printf("error canonicalizing inputs: %s\n", err)
			return nil, err
		}
	}

	return &Event{
		Type:   etype,
		Body:   string(src.Body),
		Source: canonicalized,
		Path:   uncanonicalized,
		raw:    src,
		Rev:    src.Rev,
	}, nil
}
