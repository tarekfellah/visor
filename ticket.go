// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
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
	Status       string
	source       *doozer.Event
}

// OperationType identifies different operations.
type OperationType int

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

//                                                      procType        
func CreateTicket(appName string, revName string, pName ProcessName, op OperationType, s Snapshot) (t *Ticket, err error) {
	t = &Ticket{Id: s.Rev, AppName: appName, RevisionName: revName, ProcessName: pName, Op: op, Snapshot: s, source: nil, Status: "unclaimed"}
	f, err := CreateFile(s, t.prefixPath("op"), t.toArray(), new(ListCodec))
	if err == nil {
		t.Snapshot = t.Snapshot.FastForward(f.Rev)
	}
	return t, err
}

// Claim locks the Ticket to the passed host.
func (t *Ticket) Claim(host string) (*Ticket, error) {
	exists, _, err := t.conn.Exists(t.prefixPath("claimed"))
	if err != nil {
		return t, err
	}
	if exists {
		return t, ErrTicketClaimed
	}

	rev, err := t.conn.Set(t.prefixPath("claimed"), t.Rev, []byte(host))
	t.Snapshot = t.Snapshot.FastForward(rev)

	if err == nil {
		t.Status = "claimed"
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

	err = t.conn.Del(t.prefixPath("claimed"), rev)
	if err == nil {
		t.Status = "unclaimed"
	}
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

	go WatchTicketUnclaim(s, listener)

	for {
		ev, err := s.conn.Wait(path.Join(TICKETS_PATH, "*", "op"), rev+1)
		if err != nil {
			return err
		}
		rev = ev.Rev

		if !ev.IsSet() {
			continue
		}
		ticket, err := parseTicket(s.FastForward(rev), &ev, ev.Body)
		if err != nil {
			fmt.Println(err)
			continue
		}
		listener <- ticket
	}
	return err
}

func WatchTicketUnclaim(s Snapshot, listener chan *Ticket) (err error) {
	rev := s.Rev
	for {
		ev, err := s.conn.Wait(path.Join(TICKETS_PATH, "*", "claimed"), rev+1)
		if err != nil {
			return err
		}
		rev = ev.Rev

		if ev.IsSet() {
			continue
		}
		ticket, err := parseTicket(s.FastForward(rev), &ev, nil)
		if err != nil {
			fmt.Println(err)
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
		return nil, errors.New(fmt.Sprintf("ticket id %s can't be parsed as an int64", idStr))
	}

	var decoded interface{}
	var codec = new(ListCodec)

	if body == nil {
		f, err := Get(snapshot, path.Join(TICKETS_PATH, idStr, "op"), codec)
		if err != nil {
			return t, err
		}
		decoded = f.Value
	} else {
		decoded, err = codec.Decode(body)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("invalid ticket body: %s", body))
		}
	}
	data := decoded.([]string)

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

func (t *Ticket) toArray() []string {
	return []string{t.AppName, t.RevisionName, string(t.ProcessName), t.Op.String()}
}
