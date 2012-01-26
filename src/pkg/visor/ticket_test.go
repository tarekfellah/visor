package visor

import (
	"testing"
)

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
