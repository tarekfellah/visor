package visor

import "github.com/ha/doozer"

type Event struct {
	Type   EventType
	Body   string
	Source doozer.Event
}
type EventType int

// Event types
const (
	EV_ALL      = iota // Catch-all
	EV_APPREG   = iota // App register
	EV_APPUNREG = iota // App unregister
	EV_INSREG   = iota // Instance register
	EV_INSUNREG = iota // Instance unregister
)

func (ev *Event) String() string {
	return "<event>"
}
