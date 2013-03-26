// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	cp "github.com/soundcloud/cotterpin"
	"github.com/soundcloud/visor/net"
	"testing"
	"time"
)

func runnerSetup() (s *Store) {
	s, err := DialUri(DefaultUri, "/runner-test")
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

func TestRunnerRegisterAndGet(t *testing.T) {
	var insId int64 = 787878

	s := runnerSetup()
	addr := "127.0.0.1:9999"

	r := s.NewRunner(addr, insId, new(net.Net))
	r1, err := r.Register()
	if err != nil {
		t.Fatal(err)
	}

	if r1.Addr != addr {
		t.Error("runner addr wasn't set correctly")
	}
	if r1.InstanceId != insId {
		t.Error("runner instance-id wasn't set correctly")
	}

	r2, err := s.Join(r1.Dir.Snapshot).GetRunner(addr)
	if err != nil {
		t.Fatal(err)
	}

	if r2.Addr != r1.Addr {
		t.Error("addrs don't match")
	}
	if r2.InstanceId != r1.InstanceId {
		t.Error("instance ids don't match")
	}

	err = r2.Unregister()
	if err != nil {
		t.Fatal(err)
	}

	sp, err := r2.Dir.Snapshot.FastForward()
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Join(sp).GetRunner(addr)
	if !cp.IsErrNoEnt(err) {
		t.Fatal("expected runner to be unregistered")
	}
}

func TestRunnersByHost(t *testing.T) {
	s := runnerSetup()

	_, err := s.NewRunner("10.0.1.1:7777", 9, new(net.Net)).Register()
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.NewRunner("10.0.1.2:7777", 7, new(net.Net)).Register()
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.NewRunner("10.0.1.2:7778", 8, new(net.Net)).Register()
	if err != nil {
		t.Fatal(err)
	}

	rs, err := s.Join(r.Dir.Snapshot).RunnersByHost("10.0.1.2")
	if err != nil {
		t.Fatal(err)
	}

	if len(rs) != 2 {
		t.Fatalf("incorrect number of runners for host (%d)", len(rs))
	}

	if !func() bool {
		for _, r := range rs {
			if r.Addr == "10.0.1.2:7777" {
				return true
			}
		}
		return false
	}() {
		t.Errorf("runner wasn't found under host")
	}

	if !func() bool {
		for _, r := range rs {
			if r.Addr == "10.0.1.2:7778" {
				return true
			}
		}
		return false
	}() {
		t.Errorf("runner wasn't found under host")
	}
}

func TestWatchRunnerStart(t *testing.T) {
	var insId int64 = 797979

	addr := "127.0.0.1:9898"
	s := runnerSetup()
	ch := make(chan *Runner)
	errch := make(chan error)

	go s.WatchRunnerStart("127.0.0.1", ch, errch)

	r := s.NewRunner(addr, insId, new(net.Net))
	r1, err := r.Register()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case r2 := <-ch:
		if r2.InstanceId == r1.InstanceId && r2.Addr == r1.Addr {
			break
		}
		t.Errorf("received unexpected runner: %#v", r2)
	case err := <-errch:
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Errorf("expected runner, got timeout")
	}
}

func TestWatchRunnerStop(t *testing.T) {
	var insId int64 = 797979

	addr := "127.0.0.1:9898"
	s := runnerSetup()
	ch := make(chan string)
	errch := make(chan error)

	go s.WatchRunnerStop("127.0.0.1", ch, errch)

	r := s.NewRunner(addr, insId, new(net.Net))
	r1, err := r.Register()
	if err != nil {
		t.Fatal(err)
	}
	err = r1.Unregister()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case addr1 := <-ch:
		if addr == addr {
			break
		}
		t.Errorf("received unexpected addr: %#v", addr1)
	case err := <-errch:
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Errorf("expected runner, got timeout")
	}
}
