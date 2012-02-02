package visor

import (
	"os"
	"strconv"
	"testing"
)

func ticketSetup() (c *Client, hostname string) {
	c, err := Dial(DEFAULT_ADDR, DEFAULT_ROOT)
	if err != nil {
		panic(err)
	}
	hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}

	c.Del("tickets")

	return
}

func TestNewTicket(t *testing.T) {
	c, _ := ticketSetup()
	body := "lol cat app start"

	ticket, err := NewTicket(c, "lol", "cat", "app", 0)
	if err != nil {
		t.Error(err)
	}

	b, err := c.Get(ticket.path() + "/op")
	if err != nil {
		t.Error(err)
	}
	if string(b) != body {
		t.Errorf("expected %s got %s", body, b)
	}
}

func TestClaim(t *testing.T) {
	c, host := ticketSetup()
	id := c.Rev
	op := "claim abcd123 test start"
	ticket := &Ticket{Id: id, AppName: "claim", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set("tickets/"+strconv.FormatInt(id, 10)+"/op", []byte(op))
	if err != nil {
		t.Error(err)
	}

	err = ticket.Claim(c, host)
	if err != nil {
		t.Error(err)
	}

	body, err := c.Get("tickets/" + strconv.FormatInt(id, 10) + "/claimed")
	if err != nil {
		t.Error(err)
	}
	if string(body) != host {
		t.Error("Ticket not claimed")
	}

	err = ticket.Claim(c, host)
	if err != ErrTicketClaimed {
		t.Error("Ticket claimed twice")
	}
}

func TestUnclaim(t *testing.T) {
	c, host := ticketSetup()
	id := c.Rev
	ticket := &Ticket{Id: id, AppName: "unclaim", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set("tickets/"+strconv.FormatInt(id, 10)+"/claimed", []byte(host))
	if err != nil {
		t.Error(err)
	}

	err = ticket.Unclaim(c, host)
	if err != nil {
		t.Error(err)
	}

	exists, err := c.Exists("tickets/" + strconv.FormatInt(id, 10) + "/claimed")
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Error("ticket still claimed")
	}
}

func TestUnclaimWithWrongLock(t *testing.T) {
	c, host := ticketSetup()
	p := "tickets/" + strconv.FormatInt(c.Rev, 10) + "/claimed"
	ticket := &Ticket{Id: c.Rev, AppName: "unclaim", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set(p, []byte(host))
	if err != nil {
		t.Error(err)
	}

	err = ticket.Unclaim(c, "foo.bar.local")
	if err != ErrUnauthorized {
		t.Error("ticket unclaimed with wrong lock")
	}
}

func TestDone(t *testing.T) {
	c, host := ticketSetup()
	p := "tickets/" + strconv.FormatInt(c.Rev, 10)
	ticket := &Ticket{Id: c.Rev, AppName: "done", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set(p+"/claimed", []byte(host))
	if err != nil {
		t.Error(err)
	}

	err = ticket.Done(c, host)
	if err != nil {
		t.Error(err)
	}

	exists, err := c.Exists(p)
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Error("ticket not resolved")
	}
}

func TestDoneWithWrongLock(t *testing.T) {
	c, host := ticketSetup()
	p := "tickets/" + strconv.FormatInt(c.Rev, 10)
	ticket := &Ticket{Id: c.Rev, AppName: "done", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set(p+"/claimed", []byte(host))
	if err != nil {
		t.Error(err)
	}

	err = ticket.Done(c, "foo.bar.local")
	if err != ErrUnauthorized {
		t.Error("ticket resolved with wrong lock")
	}
}
