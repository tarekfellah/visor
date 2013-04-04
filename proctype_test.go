// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"testing"
)

func proctypeSetup(appid string) (s *Store, app *App) {
	s, err := DialUri(DefaultUri, "/proctype-test")
	if err != nil {
		panic(err)
	}
	err = s.reset()
	if err != nil {
		panic(err)
	}
	s, err = s.FastForward()
	if err != nil {
		panic(err)
	}
	s, err = s.Init()
	if err != nil {
		panic(err)
	}

	app = s.NewApp(appid, "git://proctype.git", "master")

	return
}

func TestProcTypeRegisterAndGet(t *testing.T) {
	s, app := proctypeSetup("reg123")
	pty := s.NewProcType(app, "whoop", "whoop.sh")

	pty, err := pty.Register()
	if err != nil {
		t.Error(err)
	}

	s = s.Join(pty)

	check, _, err := s.GetSnapshot().Exists(pty.Dir.Name)
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Errorf("proctype %s isn't registered", pty)
	}

	pty, err = s.GetProcType(app, "whoop")
	if err != nil {
		t.Fatal(err)
	}
	if pty.Cmd != "whoop.sh" {
		t.Fatal("command was not set properly")
	}
}

func TestProcTypeRegisterWithInvalidName1(t *testing.T) {
	s, app := proctypeSetup("reg1232")
	pty := s.NewProcType(app, "who-op", "who-op.sh")

	pty, err := pty.Register()
	if err != ErrBadPtyName {
		t.Errorf("invalid proc type name (who-op) did not raise error")
	}
	if err != ErrBadPtyName && err != nil {
		t.Fatal("wrong error was raised for invalid proc type name")
	}
}

func TestProcTypeRegisterWithInvalidName2(t *testing.T) {
	s, app := proctypeSetup("reg1233")
	pty := s.NewProcType(app, "who_op", "who_op.sh")

	pty, err := pty.Register()
	if err != ErrBadPtyName {
		t.Errorf("invalid proc type name (who_op) did not raise error")
	}
	if err != ErrBadPtyName && err != nil {
		t.Fatal("wrong error was raised for invalid proc type name")
	}
}

func TestProcTypeUnregister(t *testing.T) {
	s, app := proctypeSetup("unreg123")
	pty := s.NewProcType(app, "whoop", "whoop")

	pty, err := pty.Register()
	if err != nil {
		t.Error(err)
	}

	err = pty.Unregister()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.GetSnapshot().Exists(pty.Dir.Name)
	if check {
		t.Errorf("proctype %s is still registered", pty)
	}
}

func TestProcTypeGetInstances(t *testing.T) {
	appid := "get-instances-app"
	s, app := proctypeSetup(appid)

	pty := s.NewProcType(app, "web", "web.sh")
	pty, err := pty.Register()
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		ins, err := s.RegisterInstance(appid, "128af90", "web")
		if err != nil {
			t.Fatal(err)
		}
		ins, err = ins.Claim("10.0.0.1")
		if err != nil {
			t.Fatal(err)
		}
		ins, err = ins.Started("10.0.0.1", 9999, appid+".org")
		if err != nil {
			t.Fatal(err)
		}
		pty = pty.Join(ins)
	}

	is, err := pty.GetInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(is) != 3 {
		t.Errorf("list is missing instances: %s", is)
	}
}

func TestProcTypeGetFailedInstances(t *testing.T) {
	appid := "get-failed-instances-app"
	s, app := proctypeSetup(appid)

	pty := s.NewProcType(app, "web", "web.sh")
	pty, err := pty.Register()
	if err != nil {
		t.Fatal(err)
	}

	instances := []*Instance{}

	for i := 0; i < 7; i++ {
		ins, err := s.RegisterInstance(appid, "128af9", "web")
		if err != nil {
			t.Fatal(err)
		}
		ins, err = ins.Claim("10.0.0.1")
		if err != nil {
			t.Fatal(err)
		}
		ins, err = ins.Started("10.0.0.1", 9999, appid+".org")
		if err != nil {
			t.Fatal(err)
		}
		instances = append(instances, ins)
		pty = pty.Join(ins)
	}
	for i := 0; i < 4; i++ {
		ins, err := instances[i].Failed("10.0.0.1", errors.New("no reason."))
		if err != nil {
			t.Fatal(err)
		}
		pty = pty.Join(ins)
	}

	is, err := pty.GetFailedInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(is) != 4 {
		t.Errorf("list is missing instances: %s", is)
	}
}

func TestProcTypeAttributes(t *testing.T) {
	appid := "app-with-attributes"
	var memoryLimitMb = 100
	s, app := proctypeSetup(appid)

	pty := s.NewProcType(app, "web", "web.sh")
	pty, err := pty.Register()
	if err != nil {
		t.Fatal(err)
	}

	pty, err = s.Join(pty.Dir.Snapshot).GetProcType(app, "web")
	if err != nil {
		t.Fatal(err)
	}
	if pty.Attrs.Limits.MemoryLimitMb != nil {
		t.Fatal("MemoryLimitMb should not be set at this point")
	}

	pty.Attrs.Limits.MemoryLimitMb = &memoryLimitMb
	pty, err = pty.StoreAttrs()
	if err != nil {
		t.Fatal(err)
	}

	pty, err = s.Join(pty.Dir.Snapshot).GetProcType(app, "web")
	if err != nil {
		t.Fatal(err)
	}
	if pty.Attrs.Limits.MemoryLimitMb == nil {
		t.Fatalf("MemoryLimitMb is nil")
	}
	if *pty.Attrs.Limits.MemoryLimitMb != memoryLimitMb {
		t.Fatalf("MemoryLimitMb does not contain the value that was set")
	}
}
