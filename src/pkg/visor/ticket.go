package visor

import (
	"fmt"
	"net"
	"strconv"
)

// Ticket carries instructions to start and stop Instances.
type Ticket struct {
	Snapshot
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

// NewTicket returns a new Ticket as it is represented on the coordinator, given an application name, revision name, process type and operation.
func NewTicket(appName string, revName string, pType ProcessType, op OperationType, s Snapshot) (t *Ticket, err error) {
	var o string

	switch op {
	case OpStart:
		o = "start"
	case OpStop:
		o = "stop"
	}

	t = &Ticket{Id: s.Rev, AppName: appName, RevisionName: revName, ProcessType: pType, Op: op, Snapshot: s}
	_, err = s.conn.Set(t.path()+"/op", s.Rev, []byte(fmt.Sprintf("%s %s %s %s", appName, revName, pType, o)))
	if err != nil {
		return
	}

	return
}

// Claim locks the Ticket to the passed host.
func (t *Ticket) Claim(s Snapshot, host string) (err error) {
	exists, _, err := s.conn.Exists(t.path()+"/claimed", &s.Rev)
	if err != nil {
		return
	}
	if exists {
		return ErrTicketClaimed
	}

	_, err = s.conn.Set(t.path()+"/claimed", s.Rev, []byte(host))

	return
}

// Unclaim removes the lock applied by Claim of the Ticket.
func (t *Ticket) Unclaim(s Snapshot, host string) (err error) {
	claimer, _, err := s.conn.Get(t.path()+"/claimed", &s.Rev)
	if err != nil {
		return
	}
	if string(claimer) != host {
		return ErrUnauthorized
	}

	err = s.conn.Del(t.path()+"/claimed", s.Rev)

	return
}

// Done marks the Ticket as done/solved in the registry.
func (t *Ticket) Done(s Snapshot, host string) (err error) {
	claimer, _, err := s.conn.Get(t.path()+"/claimed", &s.Rev)
	if err != nil {
		return
	}
	if string(claimer) != host {
		return ErrUnauthorized
	}

	err = s.conn.Del(t.path(), s.Rev)

	return
}

// String returns the Go-syntax representation of Ticket.
func (t *Ticket) String() string {
	return fmt.Sprintf("%#v", t)
}

func (t *Ticket) path() (path string) {
	return "tickets/" + strconv.FormatInt(t.Id, 10)
}

func Tickets() ([]Ticket, error) {
	return nil, nil
}
func HostTickets(addr string) ([]Ticket, error) {
	return nil, nil
}
func WatchTicket(listener chan *Ticket) error {
	return nil
}
