package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"regexp"
	"strconv"
)

// An Event represents a change to a file in the registry.
type Event struct {
	Type   EventType         // Type of event
	Path   map[string]string // The parsed file path
	Body   string            // Body of the changed file
	Info   interface{}       // Extra information, such as InstanceInfo
	source *doozer.Event     // Original event returned by doozer
}
type EventType int

func (e EventType) String() string {
	switch e {
	case EvAppReg:
		return "<app registered>"
	case EvAppUnreg:
		return "<app unregistered>"
	case EvRevReg:
		return "<revision registered>"
	case EvRevUnreg:
		return "<revision unregistered>"
	case EvInsReg:
		return "<instance registered>"
	case EvInsUnreg:
		return "<instance unregistered>"
	case EvInsStateChange:
		return "<instance state changed>"
	}
	return "<<unkown event>>"
}

// Event types
const (
	EvAppReg         EventType = iota // App register
	EvAppUnreg                        // App unregister
	EvRevReg                          // Revision register
	EvRevUnreg                        // Revision unregister
	EvInsReg                          // Instance register
	EvInsUnreg                        // Instance unregister
	EvInsStateChange                  // Instance state change
)

var (
	eventRegexps = map[string]*regexp.Regexp{}
	eventPaths   = map[string]EventType{
		"^/apps/([^/]+)/registered$":                                   EvAppReg,
		"^/apps/([^/]+)/revs/([^/]+)/registered$":                      EvRevReg,
		"^/apps/([^/]+)/revs/([^/]+)/procs/([^/]+)/instances/([^/]+)$": EvInsReg,
		"^/instances([^/]+)/state$":                                    EvInsStateChange,
	}
)

func init() {
	for str, _ := range eventPaths {
		re, err := regexp.Compile(str)
		if err != nil {
			panic(err)
		}
		eventRegexps[str] = re
	}
}

func (ev *Event) String() string {
	return fmt.Sprintf("%#v", ev)
}

// WatchEvent watches for changes to the registry and sends
// them as *Event objects to the provided channel.
func WatchEvent(s Snapshot, listener chan *Event) error {
	rev := s.Rev
	for {
		ev, err := s.conn.Wait("**", rev+1)
		if err != nil {
			return err
		}
		event := parseEvent(&ev)
		listener <- event
		rev = ev.Rev
	}
	return nil
}

func parseEvent(src *doozer.Event) *Event {
	path := src.Path

	etype := EventType(-1)
	emitter := map[string]string{}

	for str, ev := range eventPaths {
		re := eventRegexps[str]

		if match := re.FindStringSubmatch(path); match != nil {
			switch {
			case len(match) >= 5: // Instance
				emitter["instance"] = match[4]
				fallthrough
			case len(match) >= 4: // ProcType
				emitter["proctype"] = match[3]
				fallthrough
			case len(match) >= 3: // Revision
				emitter["rev"] = match[2]
				fallthrough
			case len(match) >= 2: // Application
				emitter["app"] = match[1]
			}

			switch ev {
			case EvAppReg:
				if src.IsSet() {
					etype = ev
				} else if src.IsDel() {
					etype = EvAppUnreg
				}
			case EvRevReg:
				if src.IsSet() {
					etype = ev
				} else if src.IsDel() {
					etype = EvRevUnreg
				}
			case EvInsReg:
				if src.IsSet() {
					etype = ev
				} else if src.IsDel() {
					etype = EvInsUnreg
				}
			case EvInsStateChange:
				if src.IsDel() {
					break
				}

				i, err := strconv.Atoi(string(src.Body))
				if err != nil {
					panic(err)
				}
				if State(i) != InsStateInitial {
					etype = ev
				}
			}
			break
		}
	}
	return &Event{etype, emitter, string(src.Body), nil, src}
}
