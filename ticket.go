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
	"time"
)

// Ticket carries instructions to start and stop Instances.
type Ticket struct {
	Path
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
	TicketStatusDead      TicketStatus = "dead"
	TicketStatusDone      TicketStatus = "done"
)

//                                                      procType        
func CreateTicket(appName string, revName string, pName ProcessName, op OperationType, s Snapshot) (t *Ticket, err error) {
	t = &Ticket{
		Id:           -1,
		AppName:      appName,
		RevisionName: revName,
		ProcessName:  pName,
		Op:           op,
		source:       nil,
		Status:       TicketStatusUnClaimed,
		Path:         Path{s, "<invalid-path>"},
	}
	return t.Create()
}

// FastForward advances the ticket in time. It returns
// a new instance of Ticket with the supplied revision.
func (t *Ticket) FastForward(rev int64) *Ticket {
	return t.Snapshot.fastForward(t, rev).(*Ticket)
}

func (t *Ticket) createSnapshot(rev int64) Snapshotable {
	tmp := *t
	tmp.Snapshot = Snapshot{rev, t.conn}
	return &tmp
}

func (t *Ticket) Create() (tt *Ticket, err error) {
	tt = t

	id, err := Getuid(t.Snapshot)
	if err != nil {
		return
	}
	t.Id = id
	t.Path.Dir = path.Join(TICKETS_PATH, strconv.FormatInt(t.Id, 10))

	f, err := CreateFile(t.Snapshot, t.Path.Prefix("op"), t.Array(), new(ListCodec))
	if err != nil {
		return
	}
	f, err = CreateFile(t.Snapshot, t.Path.Prefix("status"), string(t.Status), new(StringCodec))
	if err == nil {
		t.Snapshot = t.Snapshot.FastForward(f.FileRev)
	}
	return
}

// Claims returns the list of claimers
func (t *Ticket) Claims() (claims []string, err error) {
	rev, err := t.conn.Rev()
	if err != nil {
		return
	}
	claims, err = t.conn.Getdir(t.Path.Prefix("claims"), rev)
	if err, ok := err.(*doozer.Error); ok && err.Err == doozer.ErrNoEnt {
		claims = []string{}
		err = nil
	}
	return
}

// Claim locks the Ticket to the specified host.
func (t *Ticket) Claim(host string) (*Ticket, error) {
	status, rev, err := t.conn.Get(t.Path.Prefix("status"), nil)
	if err != nil {
		return t, err
	}
	if TicketStatus(status) != TicketStatusUnClaimed {
		return t, fmt.Errorf("ticket status is '%s'", string(status))
	}

	_, err = t.conn.Set(t.Path.Prefix("status"), rev, []byte(TicketStatusClaimed))
	if err != nil {
		return t, err
	}
	t.Status = TicketStatusClaimed

	rev, err = t.conn.Set(t.claimPath(host), rev, []byte(time.Now().UTC().String()))
	if err != nil {
		return t, err
	}

	return t.FastForward(rev), err
}

// Unclaim removes the lock applied by Claim of the Ticket.
func (t *Ticket) Unclaim(host string) (t1 *Ticket, err error) {
	exists, _, err := t.conn.Exists(t.claimPath(host))
	if !exists {
		return t, ErrUnauthorized
	}
	status, rev, err := t.conn.Get(t.Path.Prefix("status"), nil)
	if err != nil {
		return t, err
	}
	if TicketStatus(status) != TicketStatusClaimed {
		return t, fmt.Errorf("can't unclaim ticket, status is '%s'", status)
	}

	rev, err = t.conn.Set(t.Path.Prefix("status"), rev, []byte(TicketStatusUnClaimed))
	if err != nil {
		return t, err
	}
	t.Status = TicketStatusUnClaimed
	t1 = t.FastForward(rev)

	return
}

// Dead marks the ticket as "dead"
func (t *Ticket) Dead(host string) (t1 *Ticket, err error) {
	exists, _, err := t.conn.Exists(t.claimPath(host))
	if !exists {
		return t, ErrUnauthorized
	}

	rev, err := t.conn.Set(t.Path.Prefix("status"), -1, []byte(TicketStatusDead))
	if err != nil {
		return t, err
	}
	t.Status = TicketStatusDead
	t1 = t.FastForward(rev)

	return
}

// Done marks the Ticket as done/solved in the registry.
func (t *Ticket) Done(host string) (err error) {
	exists, rev, err := t.conn.Exists(t.claimPath(host))
	if !exists {
		return ErrUnauthorized
	}

	rev, err = t.conn.Rev()
	if err != nil {
		rev = t.Rev
	}

	t.conn.Set(t.Path.Prefix("status"), rev, []byte(TicketStatusDone))
	if err == nil {
		t.Status = TicketStatusDone
	}
	return
}

func WatchTicket(s Snapshot, listener chan *Ticket, errors chan error) {
	rev := s.Rev

	for {
		ev, err := s.conn.Wait(path.Join(TICKETS_PATH, "*", "status"), rev+1)
		if err != nil {
			errors <- err
			return
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
}

func WaitTicketProcessed(s Snapshot, id int64) (status TicketStatus, s1 Snapshot, err error) {
	var ev doozer.Event

	rev := s.Rev

	for {
		ev, err = s.conn.Wait(fmt.Sprintf("/%s/%d/status", TICKETS_PATH, id), rev+1)
		if err != nil {
			return
		}
		rev = ev.Rev

		if ev.IsSet() && TicketStatus(ev.Body) == TicketStatusDone {
			status = TicketStatusDone
			break
		}
		if ev.IsSet() && TicketStatus(ev.Body) == TicketStatusDead {
			status = TicketStatusDead
			break
		}
	}
	s1 = s.FastForward(rev)

	return
}

func parseTicket(snapshot Snapshot, ev *doozer.Event, body []byte) (t *Ticket, err error) {
	idStr := strings.Split(ev.Path, "/")[2]
	id, err := strconv.ParseInt(idStr, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("ticket id %s can't be parsed as an int64", idStr)
	}

	p := path.Join(TICKETS_PATH, idStr)

	f, err := Get(snapshot, path.Join(p, "op"), new(ListCodec))
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
		Path:         Path{snapshot, p},
		source:       ev}
	return t, err
}

func (t *Ticket) claimPath(host string) string {
	return t.Path.Prefix("claims", host)
}

func (t *Ticket) Fields() string {
	return fmt.Sprintf("%d %s %s %s %s", t.Id, t.AppName, t.RevisionName, string(t.ProcessName), t.Op.String())
}

func (t *Ticket) Array() []string {
	return []string{t.AppName, t.RevisionName, string(t.ProcessName), t.Op.String()}
}

// String returns the Go-syntax representation of Ticket.
func (t *Ticket) String() string {
	return fmt.Sprintf("Ticket{id: %d, op: %s, app: %s, rev: %s, proc: %s}", t.Id, t.Op.String(), t.AppName, t.RevisionName, t.ProcessName)
}

// IdString returns a string of the format "TICKET[$ticket-id]"
func (t *Ticket) IdString() string {
	return fmt.Sprintf("TICKET[%d]", t.Id)
}
