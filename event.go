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

type Metadata struct {
	Application *App
	Revision    *Revision
	ProcType    *ProcType
	Instance    *Instance
	Service     *Service
	Endpoint    *Endpoint
}

// An Event represents a change to a file in the registry.
// TODO: turn `Emitter` into own type, instead of map
type Event struct {
	Type     EventType     // Type of event
	Body     string        // Body of the changed file
	Metadata Metadata      // Extra information, such as InstanceInfo
	source   *doozer.Event // Original event returned by doozer
	Rev      int64
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

		event, err := enrichEvent(s, &ev)
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
		event, err := enrichEvent(s, &ev)
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

func canonicalizeMetadata(s Snapshot, uncanonicalized uncanonicalizedMetadata) (metadata Metadata, err error) {
	if uncanonicalized.application != nil {
		metadata.Application, err = GetApp(s, *uncanonicalized.application)

		if err != nil {
			fmt.Printf("error canonicalizing Application: %s\n", err)
			return
		}
	}

	if uncanonicalized.revision != nil {
		metadata.Revision, err = GetRevision(s, metadata.Application, *uncanonicalized.revision)

		if err != nil {
			fmt.Printf("error canonicalizing Revision: %s\n", err)
			return
		}
	}

	if uncanonicalized.proctype != nil {
		metadata.ProcType, err = GetProcType(s, metadata.Application, *uncanonicalized.proctype)
		if err != nil {
			fmt.Printf("error canonicalizing ProcType: %s\n", err)
			return
		}
	}

	if uncanonicalized.instance != nil {
		onError := func(err error) {
			fmt.Printf("error canonicalizing Instance: %s\n", err)
		}

		var id int64 = -1
		if id, err = strconv.ParseInt(*uncanonicalized.instance, 10, 64); err != nil {
			onError(err)
			return
		}
		if metadata.Instance, err = GetInstance(s, id); err != nil {
			onError(err)
			return
		}
	}

	if uncanonicalized.service != nil {
		metadata.Service, err = GetService(s, *uncanonicalized.service)
		if err != nil {
			fmt.Printf("error canonicalizing Service: %s\n", err)
			return
		}

	}

	if uncanonicalized.endpoint != nil {
		metadata.Endpoint, err = GetEndpoint(s, metadata.Service, *uncanonicalized.endpoint)
		if err != nil {
			fmt.Printf("error canonicalizing Endpoint: %s\n", err)
		}
	}
	return
}

type uncanonicalizedMetadata struct {
	application *string
	endpoint    *string
	instance    *string
	proctype    *string
	revision    *string
	service     *string
}

func enrichEvent(s Snapshot, src *doozer.Event) (event *Event, err error) {
	path := src.Path
	etype := EvUnknown
	uncanonicalized := uncanonicalizedMetadata{}

	for re, ev := range eventPatterns {
		if match := re.FindStringSubmatch(path); match != nil {
			switch ev {
			case pathApp:
				uncanonicalized.application = &match[1]

				if src.IsSet() {
					etype = EvAppReg
				} else if src.IsDel() {
					etype = EvAppUnreg
				}
			case pathRev:
				uncanonicalized.application = &match[1]
				uncanonicalized.revision = &match[2]

				if src.IsSet() {
					etype = EvRevReg
				} else if src.IsDel() {
					etype = EvRevUnreg
				}
			case pathProc:
				uncanonicalized.application = &match[1]
				uncanonicalized.proctype = &match[2]

				if src.IsSet() {
					etype = EvProcReg
				} else if src.IsDel() {
					etype = EvProcUnreg
				}
			case pathIns:
				uncanonicalized.instance = &match[1]

				if src.IsSet() {
					fields := strings.Fields(string(src.Body))
					uncanonicalized.application = &fields[0]
					uncanonicalized.revision = &fields[1]
					uncanonicalized.proctype = &fields[2]
					etype = EvInsReg
				} else if src.IsDel() {
					etype = EvInsUnreg
				}
			case pathInsStart:
				uncanonicalized.instance = &match[1]

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
				uncanonicalized.instance = &match[1]

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
				uncanonicalized.service = &match[1]

				if src.IsSet() {
					etype = EvSrvReg
				} else if src.IsDel() {
					etype = EvSrvUnreg
				}
			case pathEp:
				uncanonicalized.service = &match[1]
				uncanonicalized.endpoint = &match[2]
				if src.IsSet() {
					etype = EvEpReg
				} else if src.IsDel() {
					etype = EvEpUnreg
				}
			}
			break
		}
	}

	canonicalized, err := canonicalizeMetadata(s, uncanonicalized)

	if err != nil {
		fmt.Printf("error canonicalizing inputs: %s\n", err)
		return nil, err
	}

	return &Event{
		Type:     etype,
		Body:     string(src.Body),
		Metadata: canonicalized,
		source:   src,
		Rev:      src.Rev,
	}, nil
}
