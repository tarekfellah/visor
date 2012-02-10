package visor

import (
	"os"
	"strconv"
	"testing"
)

func ticketSetup() (s Snapshot, hostname string) {
	s, err := DialConn(DEFAULT_ADDR, DEFAULT_ROOT)
	if err != nil {
		panic(err)
	}
	hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}

	s.conn.Del("tickets", s.Rev)
	s = s.FastForward(-1)

	return
}

func TestNewTicket(t *testing.T) {
	s, _ := ticketSetup()
	body := "lol cat app start"

	ticket, err := NewTicket("lol", "cat", "app", 0, s)
	if err != nil {
		t.Error(err)
	}
	s = s.FastForward(-1)

	b, _, err := s.conn.Get(ticket.path()+"/op", &s.Rev)
	if err != nil {
		t.Error(err)
	}
	if string(b) != body {
		t.Errorf("expected %s got %s", body, string(b))
	}
}

func TestClaim(t *testing.T) {
	s, host := ticketSetup()
	id := s.Rev
	op := "claim abcd123 test start"
	ticket := &Ticket{Id: id, AppName: "claim", RevisionName: "abcd123", ProcessType: "test", Op: 0, Snapshot: s}

	rev, err := s.conn.Set("tickets/"+strconv.FormatInt(id, 10)+"/op", s.Rev, []byte(op))
	if err != nil {
		t.Error(err)
	}
	s = s.FastForward(rev)

	err = ticket.Claim(s, host)
	if err != nil {
		t.Error(err)
	}
	s = s.FastForward(s.Rev + 1)

	body, _, err := s.conn.Get("tickets/"+strconv.FormatInt(id, 10)+"/claimed", &s.Rev)
	if err != nil {
		t.Error(err)
	}
	if string(body) != host {
		t.Error("Ticket not claimed")
	}

	err = ticket.Claim(s, host)
	if err != ErrTicketClaimed {
		t.Error("Ticket claimed twice")
	}
}

func TestUnclaim(t *testing.T) {
	s, host := ticketSetup()
	id := s.Rev
	ticket := &Ticket{Id: id, AppName: "unclaim", RevisionName: "abcd123", ProcessType: "test", Op: 0, Snapshot: s}

	rev, err := s.conn.Set("tickets/"+strconv.FormatInt(id, 10)+"/claimed", s.Rev, []byte(host))
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(rev)
	err = ticket.Unclaim(s, host)
	if err != nil {
		t.Error(err)
	}

	exists, _, err := s.conn.Exists("tickets/"+strconv.FormatInt(id, 10)+"/claimed", nil)
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Error("ticket still claimed")
	}
}

func TestUnclaimWithWrongLock(t *testing.T) {
	s, host := ticketSetup()
	p := "tickets/" + strconv.FormatInt(s.Rev, 10) + "/claimed"
	ticket := &Ticket{Id: s.Rev, AppName: "unclaim", RevisionName: "abcd123", ProcessType: "test", Op: 0, Snapshot: s}

	rev, err := s.conn.Set(p, s.Rev, []byte(host))
	if err != nil {
		t.Error(err)
	}

	s = s.FastForward(rev)
	err = ticket.Unclaim(s, "foo.bar.local")
	if err != ErrUnauthorized {
		t.Error("ticket unclaimed with wrong lock")
	}
}

func TestDone(t *testing.T) {
	s, host := ticketSetup()
	p := "tickets/" + strconv.FormatInt(s.Rev, 10)
	ticket := &Ticket{Id: s.Rev, AppName: "done", RevisionName: "abcd123", ProcessType: "test", Op: 0, Snapshot: s}

	rev, err := s.conn.Set(p+"/claimed", s.Rev, []byte(host))
	if err != nil {
		t.Error(err)
	}
	s = s.FastForward(rev)

	err = ticket.Done(s, host)
	if err != nil {
		t.Error(err)
	}

	exists, _, err := s.conn.Exists(p, nil)
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Error("ticket not resolved")
	}
}

func TestDoneWithWrongLock(t *testing.T) {
	s, host := ticketSetup()
	p := "tickets/" + strconv.FormatInt(s.Rev, 10)
	ticket := &Ticket{Id: s.Rev, AppName: "done", RevisionName: "abcd123", ProcessType: "test", Op: 0, Snapshot: s}

	_, err := s.conn.Set(p+"/claimed", s.Rev, []byte(host))
	if err != nil {
		t.Error(err)
	}
	s = s.FastForward(-1)

	err = ticket.Done(s, "foo.bar.local")
	if err != ErrUnauthorized {
		t.Error("ticket resolved with wrong lock")
	}
}
