package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"net"
)

type Ticket struct {
	Type        TicketType
	App         *App
	Rev         *Revision
	ProcessType ProcessType
	Addr        net.TCPAddr
	Source      *doozer.Event
}
type TicketType int

const (
	T_START TicketType = iota
	T_STOP
)

func (t *Ticket) Claim() error {
	return nil
}
func (t *Ticket) Unclaim() error {
	return nil
}
func (t *Ticket) Done() error {
	return nil
}
func (t *Ticket) String() string {
	return fmt.Sprintf("%#v", t)
}
