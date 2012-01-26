package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"net"
)

// Ticket carries instructions to start and stop Instances.
type Ticket struct {
	Type        TicketType
	App         *App
	Rev         *Revision
	ProcessType ProcessType
	Addr        net.TCPAddr
	Source      *doozer.Event
}

// TicketType identifies different operations.
type TicketType int

const (
	T_START TicketType = iota
	T_STOP
)

// Claim locks the Ticket to the passed host
func (t *Ticket) Claim() error {
	return nil
}

// Unclaim removes the lock applied by Claim of the Ticket.
func (t *Ticket) Unclaim() error {
	return nil
}

// Done marks the Ticket as done/solved in the registry.
func (t *Ticket) Done() error {
	return nil
}

// String returns the Go-syntax representation of Ticket
func (t *Ticket) String() string {
	return fmt.Sprintf("%#v", t)
}
