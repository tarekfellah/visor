package visor

import (
	"strconv"
	"testing"
)

func ticketSetup() (c *Client) {
	c, err := Dial(DEFAULT_ADDR, DEFAULT_ROOT)
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
	c := ticketSetup()
	id := c.rev
	op := "claim abcd123 test start"
	ticket := &Ticket{Id: id, AppName: "claim", RevisionName: "abcd123", ProcessType: "test", Op: 0}

	err := c.Set("tickets/"+strconv.FormatInt(id, 10)+"/op", op)
	if err != nil {
		t.Error(err)
	}

	err = ticket.Claim(c, "localhost")
	if err != nil {
		t.Error(err)
	}

	body, err := c.Get("tickets/" + strconv.FormatInt(id, 10) + "/claimed")
	if err != nil {
		t.Error(err)
	}
	if body != "localhost" {
		t.Error("Ticket not claimed")
	}

	err = ticket.Claim(c, "localhost")
	if err != ErrTicketClaimed {
		t.Error("Ticket claimed twice")
	}
}
