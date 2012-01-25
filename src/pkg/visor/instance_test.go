package visor

import (
	"net"
	"testing"
)

func instanceSetup(addr string, pType ProcessType) (c *Client, ins *Instance) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic(err)
	}
	app := &App{Name: "ins-test", RepoUrl: "git://ins.git", Stack: "insane"}
	rev := &Revision{App: app, ref: "7abcde6"}
	ins = &Instance{Rev: rev, Addr: tcpAddr, ProcessType: pType}

	c, err = Dial(DEFAULT_ADDR)
	if err != nil {
		panic(err)
	}

	c.Del("/apps")

	return
}

func TestInstanceRegister(t *testing.T) {
	c, ins := instanceSetup("localhost:12345", "web")

	check, err := c.Exists(ins.Path())
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("Instance already registered")
	}

	err = ins.Register(c)
	if err != nil {
		t.Error(err)
	}

	check, err = c.Exists(ins.Path())
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Error("Instance registration failed")
	}

	err = ins.Register(c)
	if err == nil {
		t.Error("Instance allowed to be registered twice")
	}
}

func TestInstanceUnregister(t *testing.T) {
	c, ins := instanceSetup("localhost:54321", "worker")

	err := ins.Register(c)
	if err != nil {
		t.Error(err)
	}

	err = ins.Unregister(c)
	if err != nil {
		t.Error(err)
	}

	check, err := c.Exists(ins.Path())
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("Instance still registered")
	}
}
