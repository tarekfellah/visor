package visor

import (
	"github.com/ha/doozer"
	"net"
)

type Client struct {
	Addr *net.TCPAddr
	Conn *doozer.Conn
	Root string
}

func (c *Client) Close() error {
	return nil
}

// INSTANCES

func (c *Client) Instances() ([]Instance, error) {
	return nil, nil
}
func (c *Client) HostInstances(addr string) ([]Instance, error) {
	return nil, nil
}

// TICKETS

func (c *Client) Tickets() ([]Ticket, error) {
	return nil, nil
}
func (c *Client) HostTickets(addr string) ([]Ticket, error) {
	return nil, nil
}

// APPS

func (c *Client) Apps() ([]App, error) {
	return nil, nil
}
func (c *Client) RegisterApp() (*App, error) {
	return nil, nil
}
func (c *Client) UnregisterApp(app *App) error {
	return nil
}

// EVENTS

func (c *Client) WatchEvent(listener chan *Event) error {
	rev, _ := c.Conn.Rev()

	for {
		ev, _ := c.Conn.Wait(c.Root+"*", rev)
		event := &Event{EV_ALL, string(ev.Body), ev}
		rev = ev.Rev + 1
		listener <- event
	}
	return nil
}
func (c *Client) WatchTicket(listener chan *Ticket) error {
	return nil
}
