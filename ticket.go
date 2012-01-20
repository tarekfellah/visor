package visor

import "github.com/ha/doozer"

type Ticket struct {
	Type   TicketType
	Source doozer.Event
}
type TicketType int

const (
	TICKET_ALL = iota
)

func (t *Ticket) Claim() error {
	return nil
}
func (t *Ticket) Unclaim() error {
	return nil
}
func (t *Ticket) Done() error {
	return nil
}
func (t *Ticket) String() string {
	return "<ticket>"
}
