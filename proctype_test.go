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

func TestProcTypeRegister(t *testing.T) {
	s, app := proctypeSetup("reg123")
	pty := s.NewProcType(app, "whoop")

	pty, err := pty.Register()
	if err != nil {
		t.Error(err)
	}

	check, _, err := pty.GetSnapshot().Exists(pty.dir.Name)
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Errorf("proctype %s isn't registered", pty)
	}
}

func TestProcTypeRegisterWithInvalidName1(t *testing.T) {
	s, app := proctypeSetup("reg1232")
	pty := s.NewProcType(app, "who-op")

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
	pty := s.NewProcType(app, "who_op")

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
	pty := s.NewProcType(app, "whoop")

	pty, err := pty.Register()
	if err != nil {
		t.Error(err)
	}

	err = pty.Unregister()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.GetSnapshot().Exists(pty.dir.Name)
	if check {
		t.Errorf("proctype %s is still registered", pty)
	}
}

func TestProcTypeGetInstances(t *testing.T) {
	appid := "get-instances-app"
	s, app := proctypeSetup(appid)

	pty := s.NewProcType(app, "web")
	pty, err := pty.Register()
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		ins, err := s.RegisterInstance(appid, "128af90", "web", "default")
		if err != nil {
			t.Fatal(err)
		}
		ins, err = ins.Claim("10.0.0.1")
		if err != nil {
			t.Fatal(err)
		}
		ins, err = ins.Started("10.0.0.1", appid+".org", 9999, 10000)
		if err != nil {
			t.Fatal(err)
		}
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

	pty := s.NewProcType(app, "web")
	pty, err := pty.Register()
	if err != nil {
		t.Fatal(err)
	}

	instances := []*Instance{}

	for i := 0; i < 7; i++ {
		ins, err := s.RegisterInstance(appid, "128af9", "web", "default")
		if err != nil {
			t.Fatal(err)
		}
		ins, err = ins.Claim("10.0.0.1")
		if err != nil {
			t.Fatal(err)
		}
		ins, err = ins.Started("10.0.0.1", appid+".org", 9999, 10000)
		if err != nil {
			t.Fatal(err)
		}
		instances = append(instances, ins)
	}
	for i := 0; i < 4; i++ {
		_, err := instances[i].Failed("10.0.0.1", errors.New("no reason."))
		if err != nil {
			t.Fatal(err)
		}
	}

	failed, err := pty.GetFailedInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(failed) != 4 {
		t.Errorf("list is missing instances: %s", len(failed))
	}

	is, err := pty.GetInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(is) != 3 {
		t.Errorf("remaining instances list wrong: %d", len(is))
	}
}

func TestProcTypeGetLostInstances(t *testing.T) {
	appid := "get-lost-instances-app"
	s, app := proctypeSetup(appid)

	pty, err := s.NewProcType(app, "worker").Register()
	if err != nil {
		t.Fatal(err)
	}

	instances := []*Instance{}

	for i := 0; i < 9; i++ {
		ins, err := s.RegisterInstance(appid, "83jad2f", "worker", "mem-leak")
		if err != nil {
			t.Fatal(err)
		}
		ins, err = ins.Claim("10.3.2.1")
		if err != nil {
			t.Fatal(err)
		}
		ins, err = ins.Started("10.3.2.1", "box00.vm", 9898, 9899)
		if err != nil {
			t.Fatal(err)
		}
		instances = append(instances, ins)
	}

	for i := 0; i < 3; i++ {
		_, err := instances[i].Lost("watchman", errors.New("it's gone!!!"))
		if err != nil {
			t.Fatal(err)
		}
	}
	lost, err := pty.GetLostInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(lost) != 3 {
		t.Errorf("lost list is missing instances: %d", len(lost))
	}

	is, err := pty.GetInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(is) != 6 {
		t.Errorf("remaining instances list wrong: %d", len(is))
	}
}

func TestProcTypeAttributes(t *testing.T) {
	appid := "app-with-attributes"
	var memoryLimitMb = 100
	s, app := proctypeSetup(appid)

	pty := s.NewProcType(app, "web")
	pty, err := pty.Register()
	if err != nil {
		t.Fatal(err)
	}

	pty, err = app.GetProcType("web")
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

	pty, err = app.GetProcType("web")
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
