package visor

import (
	"strconv"
	"testing"
)

func instanceSetup(addr string, pType ProcessName) (ins *Instance) {
	s, err := Dial(DEFAULT_ADDR, "/instance-test")
	if err != nil {
		panic(err)
	}
	s.conn.Del("apps", s.Rev)
	s = s.FastForward(-1)

	app, err := NewApp("ins-test", "git://ins.git", "insane", s)
	if err != nil {
		panic(err)
	}
	rev, err := NewRevision(app, "7abcde6", s)
	rev.ArchiveUrl = "archive"
	if err != nil {
		panic(err)
	}
	pty, err := NewProcType(rev, pType, s)
	if err != nil {
		panic(err)
	}
	ins, err = NewInstance(pty, addr, 0, s)
	if err != nil {
		panic(err)
	}

	_, err = app.Register()
	if err != nil {
		panic(err)
	}
	_, err = rev.Register()
	if err != nil {
		panic(err)
	}
	_, err = pty.Register()
	if err != nil {
		panic(err)
	}

	return
}

func TestInstanceRegister(t *testing.T) {
	ins := instanceSetup("localhost:12345", "web")

	check, _, err := ins.conn.Exists(ins.Path(), &ins.Rev)
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("Instance already registered")
	}

	_, err = ins.Register()
	if err != nil {
		t.Error(err)
	}

	check, _, err = ins.conn.Exists(ins.Path(), nil)
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Error("Instance registration failed")
	}

	_, err = ins.Register()
	if err == nil {
		t.Error("Instance allowed to be registered twice")
	}
}

func TestInstanceUnregister(t *testing.T) {
	ins := instanceSetup("localhost:54321", "worker")

	ins, err := ins.Register()
	if err != nil {
		t.Error(err)
	}

	err = ins.Unregister()
	if err != nil {
		t.Error(err)
	}

	check, _, err := ins.conn.Exists(ins.Path(), nil)
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("Instance still registered")
	}
}

func TestInstances(t *testing.T) {
	ins := instanceSetup("127.0.0.1:1337", "clock")
	host := "127.0.0.1:"
	port := 1000

	for i := 0; i < 3; i++ {
		ins, err := NewInstance(ins.ProcType, host+strconv.Itoa(port+i), 0, ins.Snapshot)
		if err != nil {
			t.Error(err)
		}
		_, err = ins.Register()
		if err != nil {
			t.Error(err)
		}
	}

	ins = ins.FastForward(-1)
	instances, err := Instances(ins.Snapshot)
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
