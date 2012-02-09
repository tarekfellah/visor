package visor

import (
	"os"
	"strconv"
	"testing"
)

func ticketSetup() (c *Client, hostname string) {
	c, err := Dial(DEFAULT_ADDR, DEFAULT_ROOT, new(StringCodec))
	if err != nil {
		panic(err)
	}
	hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}

	c.Del("tickets")
	c, _ = c.FastForward(-1)

	return
}

func TestNewTicket(t *testing.T) {
	c, _ := ticketSetup()
	body := "lol cat app start"

	ticket, err := NewTicket(c, "lol", "cat", "app", 0)
	if err != nil {
		t.Error(err)
	}
	c, _ = c.FastForward(-1) // TODO: this shouldn't be needed?

	b, err := c.Get(ticket.path() + "/op")
	if err != nil {
		t.Error(err)
	}
	if b.Value.String() != body {
		t.Errorf("expected %s got %s", body, b.Value.String())
	}
}

func TestClaim(t *testing.T) {
	c, host := ticketSetup()
	id := c.rev
	op := "claim abcd123 test start"
	ticket := &Ticket{Id: id, AppName: "claim", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	_, err := c.Set("tickets/"+strconv.FormatInt(id, 10)+"/op", []byte(op))
	if err != nil {
		t.Error(err)
	}
	c, _ = c.FastForward(-1)

	err = ticket.Claim(c, host)
	if err != nil {
		t.Error(err)
	}
	c, _ = c.FastForward(-1)

	body, err := c.Get("tickets/" + strconv.FormatInt(id, 10) + "/claimed")
	if err != nil {
		t.Error(err)
	}
	if body.Value.String() != host {
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

	_, err := c.Set("tickets/"+strconv.FormatInt(id, 10)+"/claimed", []byte(host))
	if err != nil {
		t.Error(err)
	}

	c, _ = c.FastForward(-1)
	err = ticket.Unclaim(c, host)
	if err != nil {
		t.Error(err)
	}
	c, _ = c.FastForward(-1)

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

	_, err := c.Set(p, []byte(host))
	if err != nil {
		t.Error(err)
	}

	c, _ = c.FastForward(-1)
	err = ticket.Unclaim(c, "foo.bar.local")
	if err != ErrUnauthorized {
		t.Error("ticket unclaimed with wrong lock")
	}
}

func TestDone(t *testing.T) {
	c, host := ticketSetup()
	p := "tickets/" + strconv.FormatInt(c.rev, 10)
	ticket := &Ticket{Id: c.rev, AppName: "done", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	_, err := c.Set(p+"/claimed", []byte(host))
	if err != nil {
		t.Error(err)
	}
	c, _ = c.FastForward(-1)

	err = ticket.Done(c, host)
	if err != nil {
		t.Error(err)
	}
	c, _ = c.FastForward(-1)

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

	_, err := c.Set(p+"/claimed", []byte(host))
	if err != nil {
		t.Error(err)
	}
	c, _ = c.FastForward(-1)

	err = ticket.Done(c, "foo.bar.local")
	if err != ErrUnauthorized {
		t.Error("ticket resolved with wrong lock")
	}
}
