package visor

import (
	"fmt"
	"net"
)

// Ticket carries instructions to start and stop Instances.
type Ticket struct {
	AppName      string
	RevisionName string
	ProcessType  ProcessType
	Op           OperationType
	Addr         net.TCPAddr
	source       *Event
}

// OperationType identifies different operations.
type OperationType int

const (
	OpStart OperationType = iota
	OpStop
)

// NewTicket returns a new Ticket given an application name, revision name, process type and operation
func NewTicket(appName string, revName string, pType ProcessType, op OperationType) (t *Ticket) {
	t = &Ticket{AppName: appName, RevisionName: revName, ProcessType: pType, Op: op}

	return
}

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
