// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"testing"
)

func procSetup(appid string) (s *Store, app *App) {
	s, err := DialUri(DefaultUri, "/proc-test")
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

	app = s.NewApp(appid, "git://proc.git", "master")

	return
}

func TestProcRegister(t *testing.T) {
	s, app := procSetup("reg123")
	proc := s.NewProc(app, "whoop", "./whoop.sh")

	proc, err := proc.Register()
	if err != nil {
		t.Fatal(err)
	}

	check, _, err := proc.GetSnapshot().Exists(proc.dir.Name)
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Errorf("proc %s isn't registered", proc)
	}

	proc1, err := app.GetProc("whoop")
	if err != nil {
		t.Fatal(err)
	}
	if proc.Name != proc1.Name ||
		proc.Cmd != proc1.Cmd ||
		proc.Registered != proc1.Registered {
		t.Errorf("procs differ: (%s,%s,%s) - (%s,%s,%s)", proc.Name, proc.Cmd, proc.Registered, proc1.Name, proc1.Cmd, proc1.Registered)
	}
}

func TestProcRegisterWithInvalidName1(t *testing.T) {
	s, app := procSetup("reg1232")
	proc := s.NewProc(app, "who-op", "./op.sh")

	proc, err := proc.Register()
	if err != ErrBadProcName {
		t.Errorf("invalid proc type name (who-op) did not raise error")
	}
	if err != ErrBadProcName && err != nil {
		t.Fatal("wrong error was raised for invalid proc type name")
	}
}

func TestProcRegisterWithInvalidName2(t *testing.T) {
	s, app := procSetup("reg1233")
	proc := s.NewProc(app, "who_op", "./invalid.sh")

	proc, err := proc.Register()
	if err != ErrBadProcName {
		t.Errorf("invalid proc type name (who_op) did not raise error")
	}
	if err != ErrBadProcName && err != nil {
		t.Fatal("wrong error was raised for invalid proc type name")
	}
}

func TestProcUnregister(t *testing.T) {
	s, app := procSetup("unreg123")
	proc := s.NewProc(app, "whoop", "./whoop.sh")

	proc, err := proc.Register()
	if err != nil {
		t.Error(err)
	}

	err = proc.Unregister()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.GetSnapshot().Exists(proc.dir.Name)
	if check {
		t.Errorf("proc %s is still registered", proc)
	}
}

func TestProcGetInstances(t *testing.T) {
	appid := "get-instances-app"
	s, app := procSetup(appid)

	proc := s.NewProc(app, "web", "./web")
	proc, err := proc.Register()
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

	is, err := proc.GetInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(is) != 3 {
		t.Errorf("list is missing instances: %s", is)
	}
}

func TestProcGetFailedInstances(t *testing.T) {
	appid := "get-failed-instances-app"
	s, app := procSetup(appid)

	proc := s.NewProc(app, "web", "./web")
	proc, err := proc.Register()
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

	failed, err := proc.GetFailedInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(failed) != 4 {
		t.Errorf("list is missing instances: %s", len(failed))
	}

	is, err := proc.GetInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(is) != 3 {
		t.Errorf("remaining instances list wrong: %d", len(is))
	}
}

func TestProcGetLostInstances(t *testing.T) {
	appid := "get-lost-instances-app"
	s, app := procSetup(appid)

	proc, err := s.NewProc(app, "worker", "./worker.rb").Register()
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
	lost, err := proc.GetLostInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(lost) != 3 {
		t.Errorf("lost list is missing instances: %d", len(lost))
	}

	is, err := proc.GetInstances()
	if err != nil {
		t.Fatal(err)
	}
	if len(is) != 6 {
		t.Errorf("remaining instances list wrong: %d", len(is))
	}
}

func TestProcAttributes(t *testing.T) {
	appid := "app-with-attributes"
	var memoryLimitMb = 100
	s, app := procSetup(appid)

	proc := s.NewProc(app, "web", "./server")
	proc, err := proc.Register()
	if err != nil {
		t.Fatal(err)
	}

	proc, err = app.GetProc("web")
	if err != nil {
		t.Fatal(err)
	}
	if proc.Attrs.Limits.MemoryLimitMb != nil {
		t.Fatal("MemoryLimitMb should not be set at this point")
	}

	proc.Attrs.Limits.MemoryLimitMb = &memoryLimitMb
	proc, err = proc.StoreAttrs()
	if err != nil {
		t.Fatal(err)
	}

	proc, err = app.GetProc("web")
	if err != nil {
		t.Fatal(err)
	}
	if proc.Attrs.Limits.MemoryLimitMb == nil {
		t.Fatalf("MemoryLimitMb is nil")
	}
	if *proc.Attrs.Limits.MemoryLimitMb != memoryLimitMb {
		t.Fatalf("MemoryLimitMb does not contain the value that was set")
	}
}
