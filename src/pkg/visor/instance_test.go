package visor

import (
	"strconv"
	"testing"
)

func instanceSetup(addr string, pType ProcessType) (c *Client, ins *Instance) {
	app, err := NewApp("ins-test", "git://ins.git", "insane")
	if err != nil {
		panic(err)
	}
	rev, err := NewRevision(app, "7abcde6")
	if err != nil {
		panic(err)
	}
	ins, err = NewInstance(rev, addr, pType, 0)
	if err != nil {
		panic(err)
	}

	c, err = Dial(DEFAULT_ADDR)
	if err != nil {
		panic(err)
	}

	c.Del("/apps")
	err = app.Register(c)
	if err != nil {
		panic(err)
	}

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

func TestInstances(t *testing.T) {
	c, instance := instanceSetup("127.0.0.1:1337", "clock")
	host := "127.0.0.1:"
	port := 1000

	for i := 0; i < 3; i++ {
		ins, err := NewInstance(instance.Rev, host+strconv.Itoa(port+i), "clock", 0)
		if err != nil {
			t.Error(err)
		}
		err = ins.Register(c)
		if err != nil {
			t.Error(err)
		}
	}

	instances, err := Instances(c)
	if err != nil {
		t.Error(err)
	}
	if len(instances) != 3 {
		t.Errorf("expected length %d returned length %d", 3, len(instances))
	} else {
		for i := range instances {
			compAddr := host + strconv.Itoa(port+i)
			if instances[i].Addr.String() != compAddr {
				t.Errorf("expected %s got %s", compAddr, instances[i].Addr.String())
			}
		}
	}
}
