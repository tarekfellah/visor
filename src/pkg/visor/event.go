package visor

import "github.com/soundcloud/doozer"

type Event struct {
	Type   EventType
	Body   string
	Source *doozer.Event
}
type EventType int

// Event types
const (
	EV_APP_REG          EventType = iota // App register
	EV_APP_UNREG                         // App unregister
	EV_REV_REG                           // Revision register
	EV_REV_UNREG                         // Revision unregister
	EV_INS_REG                           // Instance register
	EV_INS_UNREG                         // Instance unregister
	EV_INS_STATE_CHANGE                  // Instance state change
)

func (ev *Event) String() string {
	return "<event>"
}
