// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"testing"
)

func instanceSetup() (s Snapshot) {
	s, err := Dial(DefaultAddr, "/instance-test")
	if err != nil {
		panic(err)
	}

	r, _ := s.conn.Rev()
	s.conn.Del("/", r)
	s = s.FastForward(-1)

	return
}

func instanceSetupClaimed(name, host string) (i *Instance) {
	s := instanceSetup()

	i, err := RegisterInstance(name, "128af9", "web", s)
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

	ins, err := RegisterInstance("cat", "128af9", "web", s)
	if err != nil {
		t.Fatal(err)
	}

	if ins.Status != InsStatusInitial {
		t.Error("instance status wasn't set correctly")
	}
	if ins.Id <= 0 {
		t.Error("instance id wasn't set correctly")
	}

	ins1, err := GetInstance(ins.Snapshot, ins.Id)
	if err != nil {
		t.Fatal(err)
	}

	if ins1.Id != ins.Id {
		t.Error("ids don't match")
	}
	if ins1.Status != ins.Status {
		t.Error("statuses don't match")
	}
}

func TestInstanceClaiming(t *testing.T) {
	host := "10.0.0.1"
	s := instanceSetup()

	ins, err := RegisterInstance("bat", "128af9", "web", s)
	if err != nil {
		t.Fatal(err)
	}

	ins1, err := ins.Claim(host)
	if err != nil {
		t.Fatal(err)
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

	_, err = ins1.Unclaim("9.9.9.9") // Wrong host
	if err != ErrUnauthorized {
		t.Error("expected unclaim to fail")
	}

	_, err = ins2.Unclaim(host) // Already unclaimed
	if err != ErrUnauthorized {
		t.Error("expected unclaim to fail")
	}
}

func TestInstanceStarted(t *testing.T) {
	appid := "fat"
	ptyid := "web"
	ip := "10.0.0.1"
	port := 25790
	host := "fat.the-pink-rabbit.co"
	s := instanceSetup()

	ins, err := RegisterInstance(appid, "128af9", ptyid, s)
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

	if ins2.Status != InsStatusStarted {
		t.Errorf("unexpected status '%s'", ins2.Status)
	}

	if ins2.Port != port || ins2.Host != host || ins2.Ip != ip {
		t.Errorf("instance attributes not set correctly for %#v", ins2)
	}

	ins3, err := GetInstance(ins2.Snapshot, ins2.Id)
	if err != nil {
		t.Fatal(err)
	}
	if ins3.Port != port || ins3.Host != host || ins3.Ip != ip {
		t.Errorf("instance attributes not stored correctly for %#v", ins3)
	}

	ids, err := getInstanceIds(ins2.Snapshot, appid, ptyid)
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
		t.Errorf("instance wasn't found under proc '%s'", ptyid)
	}
}

func TestInstanceStop(t *testing.T) {
	ip := "10.0.0.1"
	s := instanceSetup()

	ins, err := RegisterInstance("rat", "128af9", "web", s)
	if err != nil {
		t.Fatal(err)
	}
	ins1, err := ins.Claim(ip)
	if err != nil {
		t.Fatal(err)
	}

	_, err = StopInstance(ins.Id, ins1.Snapshot)
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

	ins, err := RegisterInstance("rat-cat", "128af9", "web", s)
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

	_, err = ins2.Exited()
	if err != ErrUnauthorized {
		t.Error("expected command to fail")
	}

	s1, err := StopInstance(ins2.Id, ins2.Snapshot)
	if err != nil {
		t.Fatal(err)
	}

	ins3, err := ins2.FastForward(s1.Rev).Exited()
	if err != nil {
		t.Fatal(err)
	}
	testInstanceStatus(t, ins.Id, InsStatusExited, ins3.Snapshot)
}

func TestInstanceUnclaimable(t *testing.T) {
	ip := "10.0.0.1"
	ins := instanceSetupClaimed("bat-cat", ip)

	ins1, err := ins.Unclaimable(ip, errors.New("because."))
	if err != nil {
		t.Fatal(err)
	}
	testInstanceStatus(t, ins.Id, InsStatusUnclaimable, ins1.Snapshot)

	_, err = ins1.Unclaimable("9.9.9.9", errors.New("no reason."))
	if err != ErrUnauthorized {
		t.Error("expected command to fail")
	}
}

func TestInstanceDead(t *testing.T) {
	ip := "10.0.0.1"
	ins := instanceSetupClaimed("fat-cat", ip)

	ins, err := ins.Started(ip, 9999, "fat-cat.com")
	if err != nil {
		t.Fatal(err)
	}

	ins1, err := ins.Dead(ip, errors.New("because."))
	if err != nil {
		t.Fatal(err)
	}
	testInstanceStatus(t, ins.Id, InsStatusDead, ins1.Snapshot)

	_, err = ins.Dead("9.9.9.9", errors.New("no reason."))
	if err != ErrUnauthorized {
		t.Error("expected command to fail")
	}

	// Note: we do not test whether or not dead instances can be retrieved
	// here. See the proctype tests & (*Proctype).GetDeadInstances()
}

func TestInstanceFailed(t *testing.T) {
	ip := "10.0.0.1"
	ins := instanceSetupClaimed("fat-bat", ip)

	ins, err := ins.Started(ip, 9999, "fat-bat.com")
	if err != nil {
		t.Fatal(err)
	}

	ins1, err := ins.Failed(ip, errors.New("because."))
	if err != nil {
		t.Fatal(err)
	}
	testInstanceStatus(t, ins.Id, InsStatusFailed, ins1.Snapshot)

	_, err = ins.Failed("9.9.9.9", errors.New("no reason."))
	if err != ErrUnauthorized {
		t.Error("expected command to fail")
	}
}

func testInstanceStatus(t *testing.T, id int64, status InsStatus, s Snapshot) {
	ins, err := GetInstance(s, id)
	if err != nil {
		t.Fatal(err)
	}
	if ins.Status != status {
		t.Errorf("expected instance status to be '%s' got '%s'", status, ins.Status)
	}
}
