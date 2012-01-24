package visor

import (
	"github.com/soundcloud/doozer"
	"net"
)

type Client struct {
	Addr *net.TCPAddr
	Conn *doozer.Conn
	Root string
	Rev  int64
}

func (c *Client) Close() error {
	return nil
}

//
// TODO find the appropriate location for this helper
//
func (c *Client) Deldir(dirname string, rev int64) (err error) {
	doozer.Walk(c.Conn, rev, dirname, func(path string, f *doozer.FileInfo, e error) error {
		if e == nil {
			err = c.Conn.Del(path, rev)
		}

		return nil
	})

	return err
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

func (c *Client) Apps() ([]*App, error) {
	appNames, err := c.Conn.Getdir("/apps", c.Rev, 0, -1)
	// FIXME proper error handling
	if err != nil {
		return []*App{}, nil
	}
	apps := make([]*App, len(appNames))

	for i := range appNames {
		apps[i] = &App{Name: appNames[i]}
	}

	return apps, err
}
func (c *Client) RegisterApp(name string, repoUrl string, stack Stack) (app *App, err error) {
	app = &App{Name: name, RepoUrl: repoUrl, Stack: stack}
	err = app.Register(c)

	return
}
func (c *Client) UnregisterApp(app *App) error {
	return nil
}

// EVENTS

func (c *Client) WatchEvent(listener chan *Event) error {
	rev, _ := c.Conn.Rev()

	for {
		ev, _ := c.Conn.Wait(c.Root+"*", rev)
		event := &Event{EV_APP_REG, string(ev.Body), &ev}
		rev = ev.Rev + 1
		listener <- event
	}
	return nil
}
func (c *Client) WatchTicket(listener chan *Ticket) error {
	return nil
}

func (c *Client) Exists(path string) (exists bool, err error) {
	_, rev, err := c.Conn.Stat(path, nil)
	if err != nil {
		return
	}

	switch rev {
	case 0:
		exists = false
	default:
		exists = true
	}

	return exists, nil
}
