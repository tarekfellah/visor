// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

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
	s.conn.Del("/", s.Rev)

	s = s.FastForward(-1)

	r, err := Init(s)
	if err != nil {
		panic(err)
	}

	s = s.FastForward(r)

	app, err := NewApp("ins-test", "git://ins.git", "insane", s)
	if err != nil {
		panic(err)
	}
	rev, err := NewRevision(app, "7abcde6", s)
	rev.ArchiveUrl = "archive"
	if err != nil {
		panic(err)
	}
	pty := NewProcType(rev, pType, s)
	ins, err = NewInstance(pty, addr, InsStateInitial, s)
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

	check, _, err := ins.conn.Exists(ins.Path())
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

	check, _, err = ins.conn.Exists(ins.Path())
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Error("Instance registration failed")
	}

	check, _, err = ins.conn.Exists(ins.ProcType.InstancePath(ins.Id()))
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
		return
	}

	err = ins.Unregister()
	if err != nil {
		t.Error(err)
	}

	check, _, err := ins.conn.Exists(ins.Path())
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("Instance still registered")
	}
}

func TestInstanceUpdateState(t *testing.T) {
	ins := instanceSetup("localhost:54321", "stateChangeWorker")

	ins, err := ins.Register()
	if err != nil {
		t.Error(err)
	}

	newIns, err := ins.UpdateState(InsStateStarted)
	if err != nil {
		t.Error(err)
	}

	if newIns.State != InsStateStarted {
		t.Error("Instance state wasn't updated")
	}

	if newIns.Rev <= ins.Rev {
		t.Error("Instance wasn't fast forwarded")
	}

	val, _, err := newIns.conn.Get(newIns.Path()+"/state", &newIns.Rev)
	if err != nil {
		t.Error(err)
	}

	if State(val) != InsStateStarted {
		t.Error("Instance state wasn't persisted in the coordinator")
	}
}

func TestInstances(t *testing.T) {
	ins := instanceSetup("127.0.0.1:1337", "clock")
	host := "127.0.0.1:"
	port := 1000

	for i := 0; i < 3; i++ {
		ins, err := NewInstance(ins.ProcType, host+strconv.Itoa(port+i), InsStateInitial, ins.Snapshot)
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
