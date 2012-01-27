package visor

import (
	"github.com/soundcloud/doozer"
	"regexp"
	"strings"
)

type Event struct {
	Type    EventType
	Emitter map[string]string
	Body    string
	Path    string
	source  *doozer.Event
}
type EventType int

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

var eventPaths = map[string]EventType{
	"^/apps/([^/]+)/registered$":                      EvAppReg,
	"^/apps/([^/]+)/revs/([^/]+)/registered$":         EvRevReg,
	"^/apps/([^/]+)/revs/([^/]+)/([^/]+)/registered$": EvInsReg,
	"^/apps/([^/]+)/revs/([^/]+)/([^/]+)/state$":      EvInsStateChange,
}

func NewEvent(etype EventType, emitter map[string]string, body string, src *doozer.Event) (ev *Event) {
	return &Event{etype, emitter, body, src.Path, src}
}

func (ev *Event) String() string {
	return "<event>"
}

func (c *Client) WatchEvent(listener chan *Event) error {
	rev, _ := c.conn.Rev()
	path := c.prefixPath("**")

	for {
		ev, _ := c.conn.Wait(path, rev+1)
		event := c.parseEvent(&ev)
		listener <- event
		rev = ev.Rev
	}
	return nil
}

func (c *Client) parseEvent(src *doozer.Event) *Event {
	path := strings.Replace(src.Path, c.Root, "", 1)

	etype := EventType(-1)
	emitter := map[string]string{}

	for re, ev := range eventPaths {
		// TODO: Compile all Regexps beforehand
		re, err := regexp.Compile(re)
		if err != nil {
			panic(err)
		}

		if match := re.FindStringSubmatch(path); match != nil {
			switch {
			case len(match) >= 4: // Instance
				emitter["instance"] = match[3]
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
				if src.IsSet() {
					etype = ev
				} else {
					etype = -1
				}
			}
			break
		}
	}
	return NewEvent(etype, emitter, string(src.Body), src)
}
