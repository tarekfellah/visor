// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func instanceSetup() *Store {
	s, err := DialUri(DefaultUri, "/instance-test")
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

	return s
}

func instanceSetupClaimed(name, host string) (i *Instance) {
	s := instanceSetup()

	i, err := s.RegisterInstance(name, "128af9", "web", "default")
	if err != nil {
		panic(err)
	}

	i, err = i.Claim(host)
	if err != nil {
		panic(err)
	}
	return
}

func TestInstanceRegisterAndGet(t *testing.T) {
	s := instanceSetup()

	ins, err := s.RegisterInstance("cat", "128af9", "web", "default")
	if err != nil {
		t.Fatal(err)
	}

	if ins.Status != InsStatusPending {
		t.Error("instance status wasn't set correctly")
	}
	if ins.Id <= 0 {
		t.Error("instance id wasn't set correctly")
	}

	ins1, err := s.GetInstance(ins.Id)
	if err != nil {
		t.Fatal(err)
	}

	if ins1.Id != ins.Id {
		t.Error("ids don't match")
	}
	if ins1.Status != ins.Status {
		t.Error("statuses don't match")
	}
	if ins1.Restarts.Fail != 0 {
		t.Error("restarts != 0")
	}
}

func TestInstanceClaiming(t *testing.T) {
	host := "10.0.0.1"
	s := instanceSetup()

	ins, err := s.RegisterInstance("bat", "128af9", "web", "default")
	if err != nil {
		t.Fatal(err)
	}

	ins1, err := ins.Claim(host)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ins.Claim(host) // Already claimed
	if !IsErrInsClaimed(err) {
		t.Error("expected re-claim to fail")
	}

	_, err = ins1.Claim(host) // Already claimed
	if !IsErrInsClaimed(err) {
		t.Error("expected re-claim to fail")
	}

	claims, err := ins1.Claims()
	if err != nil {
		t.Fatal(err)
	}

	if len(claims) == 0 {
		t.Error("instance claim was unsuccessful")
	}
	if claims[0] != host {
		t.Error("instance claims doesn't include claimer")
	}

	ins2, err := ins1.Unclaim(host)
	if err != nil {
		t.Fatal(err)
	}

	claimer, err := ins2.getClaimer()
	if err != nil {
		t.Fatal(err)
	}
	if claimer != nil {
		t.Error("ticket wasn't unclaimed properly")
	}

	_, err = ins1.Unclaim("9.9.9.9") // Wrong host
	if !IsErrUnauthorized(err) {
		t.Error("expected unclaim to fail")
	}

	_, err = ins2.Unclaim(host) // Already unclaimed
	if !IsErrUnauthorized(err) {
		t.Error("expected unclaim to fail")
	}
}

func TestInstanceStarted(t *testing.T) {
	app := "fat"
	rev := "128af9"
	proc := "web"
	env := "default"
	ip := "10.0.0.1"
	port := 25790
	host := "fat.the-pink-rabbit.co"
	s := instanceSetup()

	ins, err := s.RegisterInstance(app, rev, proc, env)
	if err != nil {
		t.Fatal(err)
	}
	ins1, err := ins.Claim(ip)
	if err != nil {
		t.Fatal(err)
	}

	ins2, err := ins1.Started(ip, port, host)
	if err != nil {
		t.Fatal(err)
	}

	if ins2.Status != InsStatusRunning {
		t.Errorf("unexpected status '%s'", ins2.Status)
	}

	if ins2.Port != port || ins2.Host != host || ins2.Ip != ip {
		t.Errorf("instance attributes not set correctly for %#v", ins2)
	}

	ins3, err := s.GetInstance(ins2.Id)
	if err != nil {
		t.Fatal(err)
	}
	if ins3.Port != port || ins3.Host != host || ins3.Ip != ip {
		t.Errorf("instance attributes not stored correctly for %#v", ins3)
	}

	ids, err := getInstanceIds(app, rev, proc, ins3)
	if err != nil {
		t.Fatal(err)
	}

	if !func() bool {
		for _, id := range ids {
			if id == ins.Id {
				return true
			}
		}
		return false
	}() {
		t.Errorf("instance wasn't found under proc '%s'", proc)
	}
}

func TestInstanceStop(t *testing.T) {
	ip := "10.0.0.1"
	s := instanceSetup()

	ins, err := s.RegisterInstance("rat", "128af9", "web", "default")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ins.Claim(ip)
	if err != nil {
		t.Fatal(err)
	}

	ins, err = ins.Started(ip, 5555, "localhost")
	if err != nil {
		t.Fatal(err)
	}

	err = ins.Stop()
	if err != nil {
		t.Fatal(err)
	}
	// Note: we aren't checking that the files are created in the coordinator,
	// that is better tested via events in event.go, as we don't want to couple
	// the tests with the schema.
}

func TestInstanceExited(t *testing.T) {
	ip := "10.0.0.1"
	port := 25790
	host := "fat.the-pink-rabbit.co"
	s := instanceSetup()

	ins, err := s.RegisterInstance("rat-cat", "128af9", "web", "default")
	if err != nil {
		t.Fatal(err)
	}
	ins, err = ins.Claim(ip)
	if err != nil {
		t.Fatal(err)
	}

	ins, err = ins.Started(ip, port, host)
	if err != nil {
		t.Fatal(err)
	}

	err = ins.Stop()
	if err != nil {
		t.Fatal(err)
	}

	ins, err = ins.Exited(ip)
	if err != nil {
		t.Fatal(err)
	}
	testInstanceStatus(s, t, ins.Id, InsStatusExited)
}

func TestInstanceRestarted(t *testing.T) {
	ip := "10.0.0.1"
	ins := instanceSetupClaimed("fat-pat", ip)

	ins, err := ins.Started(ip, 9999, "fat-pat.com")
	if err != nil {
		t.Fatal(err)
	}

	if ins.Restarts.Fail != 0 {
		t.Error("expected restart count to be 0")
	}

	ins1, err := ins.Restarted(RestartFail, 1)
	if err != nil {
		t.Fatal(err)
	}

	if ins1.Restarts.Fail != 1 {
		t.Error("expected restart count to be set to 1")
	}

	ins2, err := ins1.Restarted(RestartFail, 1)
	if err != nil {
		t.Fatal(err)
	}

	if ins2.Restarts.Fail != 2 {
		t.Error("expected restart count to be set to 2")
	}
	ins3, err := storeFromSnapshotable(ins).GetInstance(ins.Id)
	if err != nil {
		t.Fatal(err)
	}
	if ins3.Restarts.Fail != 2 {
		t.Error("expected restart count to be set to 2")
	}
}

func TestInstanceFailed(t *testing.T) {
	ip := "10.0.0.1"
	ins := instanceSetupClaimed("fat-cat", ip)

	ins, err := ins.Started(ip, 9999, "fat-cat.com")
	if err != nil {
		t.Fatal(err)
	}

	ins1, err := ins.Failed(ip, errors.New("because."))
	if err != nil {
		t.Fatal(err)
	}
	testInstanceStatus(storeFromSnapshotable(ins1), t, ins.Id, InsStatusFailed)

	_, err = ins.Failed("9.9.9.9", errors.New("no reason."))
	if !IsErrUnauthorized(err) {
		t.Error("expected command to fail")
	}

	// Note: we do not test whether or not failed instances can be retrieved
	// here. See the proctype tests & (*Proctype).GetFailedInstances()
}

func TestInstanceLost(t *testing.T) {
	ip := "10.0.0.2"
	ins := instanceSetupClaimed("slim-cat", ip)

	ins, err := ins.Started(ip, 9898, "box00.vm")
	if err != nil {
		t.Fatal(err)
	}
	ins, err = ins.Lost("watchdog", errors.New("dunno if cat is dead or not"))
	if err != nil {
		t.Fatal(err)
	}
	testInstanceStatus(storeFromSnapshotable(ins), t, ins.Id, InsStatusLost)

	// Note: we do not test whether or not lost instances can be retrieved
	// here. See the proctype tests & (*Proctype).GetLostInstances()
}

func TestWatchInstanceStartAndStop(t *testing.T) {
	app := "w-app"
	rev := "w-rev"
	proc := "w-proc"
	env := "w-env"
	s := instanceSetup()
	l := make(chan *Instance)

	go s.WatchInstanceStart(l, make(chan error))

	ins, err := s.RegisterInstance(app, rev, proc, env)
	if err != nil {
		t.Error(err)
	}

	select {
	case ins = <-l:
		// TODO Check other fields
		if ins.AppName == app && ins.RevisionName == rev && ins.ProcessName == proc {
			break
		}
		t.Errorf("received unexpected instance: %s", ins.String())
	case <-time.After(time.Second):
		t.Errorf("expected instance, got timeout")
	}

	ins, err = ins.Claim("10.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	ins, err = ins.Started("10.0.0.1", 5555, "localhost")
	if err != nil {
		t.Fatal(err)
	}

	// Stop test

	ch := make(chan *Instance)

	go func() {
		ins, err := ins.WaitStop()
		if err != nil {
			t.Fatal(err)
		}
		ch <- ins
	}()

	err = ins.Stop()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case ins1 := <-ch:
		if ins1 == nil {
			t.Error("instance is nil")
		}
	case <-time.After(time.Second):
		t.Errorf("expected instance, got timeout")
	}
}

func TestInstanceWait(t *testing.T) {
	s := instanceSetup()
	ins, err := s.RegisterInstance("bob", "985245a", "web", "default")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if _, err := ins.Claim("127.0.0.1"); err != nil {
			panic(err)
		}
	}()
	ins1, err := ins.WaitClaimed()
	if err != nil {
		t.Error(err)
	}
	if ins1.Status != InsStatusClaimed {
		t.Errorf("expected instance status to be %s", InsStatusClaimed)
	}

	go func() {
		if _, err := ins1.Started("127.0.0.1", 9000, "localhost"); err != nil {
			panic(err)
		}
	}()
	ins2, err := ins1.WaitStarted()
	if err != nil {
		t.Error(err)
	}
	if ins2.Status != InsStatusRunning {
		t.Errorf("expected instance status to be %s", InsStatusRunning)
	}
	if ins2.Ip != "127.0.0.1" || ins2.Port != 9000 || ins2.Host != "localhost" {
		t.Errorf("expected ip/port/host to match for %#v", ins2)
	}
}

func TestInstanceWaitStop(t *testing.T) {
	s := instanceSetup()
	ins, err := s.RegisterInstance("bobby", "985245a", "web", "default")
	if err != nil {
		t.Fatal(err)
	}
	ins, err = ins.Claim("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ins.Started("127.0.0.1", 9000, "localhost"); err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := ins.Stop(); err != nil {
			panic(err)
		}
	}()
	_, err = ins.WaitStop()
	if err != nil {
		t.Fatal(err)
	}
	ins, err = s.GetInstance(ins.Id)
	if err != nil {
		t.Fatal(err)
	}
	if ins.Status != InsStatusStopping {
		t.Error("expected instance to be stopped")
	}
	// println(ins.GetSnapshot().Rev)
	// println(ins1.GetSnapshot().Rev)
	// if ins1.GetSnapshot().Rev <= ins.GetSnapshot().Rev {
	// 	t.Error("expected new revision to be greater than previous")
	// }
}

func TestInstanceLocking(t *testing.T) {
	ip := "10.0.10.0"
	ins := instanceSetupClaimed("grumpy-cat", ip)

	ins, err := ins.Started(ip, 7676, "box01.vm")
	if err != nil {
		t.Fatal(err)
	}
	ins, err = ins.Lock("schroedinger", errors.New("to be rescheduled"))
	if err != nil {
		t.Fatal(err)
	}
	locked, err := ins.IsLocked()
	if err != nil {
		t.Fatal(err)
	}
	if !locked {
		t.Fatal("expected instance to be locked")
	}
	_, err = ins.Lock("somebody", errors.New("steal the lock"))
	fmt.Printf("%#v\n", err)
	if !IsErrUnauthorized(err) {
		t.Fatal("expected to not allow aquire lock twice")
	}
	ins, err = ins.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	locked, err = ins.IsLocked()
	if err != nil {
		t.Fatal(err)
	}
	if locked {
		t.Error("expected instance to not be locked")
	}
}

func testInstanceStatus(s *Store, t *testing.T, id int64, status InsStatus) {
	ins, err := s.GetInstance(id)
	if err != nil {
		t.Fatal(err)
	}
	if ins.Status != status {
		t.Errorf("expected instance status to be '%s' got '%s'", status, ins.Status)
	}
}
