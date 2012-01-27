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
	ticket := NewTicket("lol", "cat", "app", 0)

	if ticket.AppName != "lol" {
		t.Errorf("expected %s got %s", "lol", ticket.AppName)
	}

	if ticket.RevisionName != "cat" {
		t.Errorf("expected %s got %s", "cat", ticket.RevisionName)
	}

	if ticket.ProcessType != "app" {
		t.Errorf("expected %s got %s", "app", ticket.ProcessType)
	}

	if ticket.Op != OpStart {
		t.Errorf("expected %d got %d", OpStart, ticket.Op)
	}
}

func TestClaim(t *testing.T) {
	c, host := ticketSetup()
	id := c.rev
	op := "claim abcd123 test start"
	ticket := &Ticket{Id: id, AppName: "claim", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set("tickets/"+strconv.FormatInt(id, 10)+"/op", op)
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
	if body != host {
		t.Error("Ticket not claimed")
	}

	err = ticket.Claim(c, host)
	if err != ErrTicketClaimed {
		t.Error("Ticket claimed twice")
	}
}

func TestUnclaim(t *testing.T) {
	c, host := ticketSetup()
	id := c.rev
	ticket := &Ticket{Id: id, AppName: "unclaim", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set("tickets/"+strconv.FormatInt(id, 10)+"/claimed", host)
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
	p := "tickets/" + strconv.FormatInt(c.rev, 10) + "/claimed"
	ticket := &Ticket{Id: c.rev, AppName: "unclaim", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set(p, host)
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
	p := "tickets/" + strconv.FormatInt(c.rev, 10)
	ticket := &Ticket{Id: c.rev, AppName: "done", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set(p+"/claimed", host)
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
	p := "tickets/" + strconv.FormatInt(c.rev, 10)
	ticket := &Ticket{Id: c.rev, AppName: "done", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set(p+"/claimed", host)
	if err != nil {
		t.Error(err)
	}

	err = ticket.Done(c, "foo.bar.local")
	if err != ErrUnauthorized {
		t.Error("ticket resolved with wrong lock")
	}
}
