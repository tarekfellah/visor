package visor

import (
	"fmt"
	"net"
	"strconv"
)

// Ticket carries instructions to start and stop Instances.
type Ticket struct {
	Id           int64
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
func (t *Ticket) Claim(c *Client, host string) (err error) {
	exists, err := c.Exists(t.path() + "/claimed")
	if err != nil {
		return
	}
	if exists {
		return ErrTicketClaimed
	}

	err = c.Set(t.path()+"/claimed", host)
	if err != nil {
		return
	}

	return
}

// Unclaim removes the lock applied by Claim of the Ticket.
func (t *Ticket) Unclaim(c *Client, host string) (err error) {
	err = t.locked(c, host)
	if err != nil {
		return
	}

	err = c.Del(t.path() + "/claimed")

	return
}

// Done marks the Ticket as done/solved in the registry.
func (t *Ticket) Done(c *Client, host string) (err error) {
	err = t.locked(c, host)
	if err != nil {
		return
	}

	err = c.Del(t.path())

	return
}

// String returns the Go-syntax representation of Ticket
func (t *Ticket) String() string {
	return fmt.Sprintf("%#v", t)
}

func (t *Ticket) path() (path string) {
	return "tickets/" + strconv.FormatInt(t.Id, 10)
}

func (t *Ticket) locked(c *Client, host string) (err error) {
	lock, err := c.Get(t.path() + "/claimed")
	if err != nil {
		return
	}
	if lock != host {
		return ErrUnauthorized
	}

	return
}
