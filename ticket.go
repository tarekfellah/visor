// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"net"
	"path"
	"strconv"
	"strings"
)

// Ticket carries instructions to start and stop Instances.
type Ticket struct {
	Snapshot
	Id           int64
	AppName      string
	RevisionName string
	ProcessName  ProcessName
	Op           OperationType
	Addr         net.TCPAddr
	Status       TicketStatus
	source       *doozer.Event
}

// OperationType identifies different operations.
type OperationType int
type TicketStatus string

func NewOperationType(opStr string) OperationType {
	var op OperationType
	switch opStr {
	case "start":
		op = OpStart
	case "stop":
		op = OpStop
	default:
		op = OpInvalid
	}
	return op
}

func (op OperationType) String() string {
	var o string
	switch op {
	case OpStart:
		o = "start"
	case OpStop:
		o = "stop"
	case OpInvalid:
		o = "<invalid>"
	}
	return o
}

const TICKETS_PATH = "tickets"

const (
	OpInvalid               = -1
	OpStart   OperationType = 0
	OpStop                  = 1
)

const (
	TicketStatusClaimed   TicketStatus = "claimed"
	TicketStatusUnClaimed TicketStatus = "unclaimed"
)

//                                                      procType        
func CreateTicket(appName string, revName string, pName ProcessName, op OperationType, s Snapshot) (t *Ticket, err error) {
	t = &Ticket{
		Id:           s.Rev,
		AppName:      appName,
		RevisionName: revName,
		ProcessName:  pName,
		Op:           op,
		Snapshot:     s,
		source:       nil,
		Status:       TicketStatusUnClaimed,
	}
	return t.Create()
}

func (t *Ticket) Create() (*Ticket, error) {
	f, err := CreateFile(t.Snapshot, t.prefixPath("op"), t.toArray(), new(ListCodec))
	if err != nil {
		return t, err
	}
	f, err = CreateFile(t.Snapshot, t.prefixPath("status"), string(t.Status), new(StringCodec))
	if err == nil {
		t.Snapshot = t.Snapshot.FastForward(f.Rev)
	}
	return t, err
}

// Claim locks the Ticket to the passed host.
func (t *Ticket) Claim(host string) (*Ticket, error) {
	status, _, err := t.conn.Get(t.prefixPath("status"), &t.Rev)
	if err != nil {
		return t, err
	}
	if TicketStatus(status) == TicketStatusClaimed {
		return t, ErrTicketClaimed
	}

	rev, err := t.conn.Set(t.prefixPath("status"), t.Rev, []byte(TicketStatusClaimed))
	if err != nil {
		return nil, err
	}

	rev, err = t.conn.Set(t.prefixPath("claimed"), t.Rev, []byte(host))
	t.Snapshot = t.Snapshot.FastForward(rev)
	if err == nil {
		t.Status = TicketStatusClaimed
	}

	return t, err
}

// Unclaim removes the lock applied by Claim of the Ticket.
func (t *Ticket) Unclaim(host string) (err error) {
	claimer, rev, err := t.conn.Get(t.prefixPath("claimed"), nil)
	if err != nil {
		return
	}
	if string(claimer) != host {
		return ErrUnauthorized
	}

	rev, err = t.conn.Set(t.prefixPath("status"), rev, []byte("unclaimed"))
	if err == nil {
		t.Status = "unclaimed"
	}
	// TODO: Return new snapshot
	return
}

// Done marks the Ticket as done/solved in the registry.
func (t *Ticket) Done(host string) (err error) {
	claimer, rev, err := t.conn.Get(t.prefixPath("claimed"), nil)
	if err != nil {
		return
	}
	if string(claimer) != host {
		return ErrUnauthorized
	}

	err = t.conn.Del(t.Path(), rev)
	if err == nil {
		t.Status = "done"
	}
	return
}

// String returns the Go-syntax representation of Ticket.
func (t *Ticket) String() string {
	return fmt.Sprintf("Ticket{id: %d, op: %s, app: %s, rev: %s, proc: %s}", t.Id, t.Op.String(), t.AppName, t.RevisionName, t.ProcessName)
}

// IdString returns a string of the format "TICKET[$ticket-id]"
func (t *Ticket) IdString() string {
	return fmt.Sprintf("TICKET[%d]", t.Id)
}

func (t *Ticket) Path() string {
	return path.Join(TICKETS_PATH, strconv.FormatInt(t.Id, 10))
}

func (t *Ticket) prefixPath(aPath string) string {
	return path.Join(t.Path(), aPath)
}

func Tickets() ([]Ticket, error) {
	return nil, nil
}

func HostTickets(addr string) ([]Ticket, error) {
	return nil, nil
}

func WatchTicket(s Snapshot, listener chan *Ticket) (err error) {
	rev := s.Rev

	for {
		ev, err := s.conn.Wait(path.Join(TICKETS_PATH, "*", "status"), rev+1)
		if err != nil {
			return err
		}
		rev = ev.Rev

		if !ev.IsSet() || string(ev.Body) != "unclaimed" {
			continue
		}

		ticket, err := parseTicket(s.FastForward(rev), &ev, ev.Body)
		if err != nil {
			continue
		}
		listener <- ticket
	}
	return err
}

func parseTicket(snapshot Snapshot, ev *doozer.Event, body []byte) (t *Ticket, err error) {
	idStr := strings.Split(ev.Path, "/")[2]
	id, err := strconv.ParseInt(idStr, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("ticket id %s can't be parsed as an int64", idStr)
	}

	f, err := Get(snapshot, path.Join(TICKETS_PATH, idStr, "op"), new(ListCodec))
	if err != nil {
		return t, err
	}
	data := f.Value.([]string)

	t = &Ticket{
		Id:           id,
		AppName:      data[0],
		RevisionName: data[1],
		ProcessName:  ProcessName(data[2]),
		Op:           NewOperationType(data[3]),
		Snapshot:     snapshot,
		source:       ev}
	return t, err
}

func (t *Ticket) Fields() string {
	return fmt.Sprintf("%d %s %s %s %s", t.Id, t.AppName, t.RevisionName, string(t.ProcessName), t.Op.String())
}

func (t *Ticket) toArray() []string {
	return []string{t.AppName, t.RevisionName, string(t.ProcessName), t.Op.String()}
}
