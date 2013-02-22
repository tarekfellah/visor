// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"github.com/soundcloud/visor/net"
	"testing"
	"time"
)

func runnerSetup() (s Snapshot) {
	s, err := Dial(DefaultAddr, "/runner-test")
	if err != nil {
		panic(err)
	}

	r, _ := s.conn.Rev()
	s.conn.Del("/", r)
	s = s.FastForward(-1)

	return
}

func TestRunnerRegisterAndGet(t *testing.T) {
	var insId int64 = 787878

	s := runnerSetup()
	addr := "127.0.0.1:9999"

	r := NewRunner(addr, insId, new(net.Net), s)
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

	r2, err := GetRunner(r1.Dir.Snapshot, addr)
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

	_, err = GetRunner(r2.Dir.Snapshot.FastForward(-1), addr)
	if !IsErrNoEnt(err) {
		t.Fatal("expected runner to be unregistered")
	}
}

func TestRunnersByHost(t *testing.T) {
	s := runnerSetup()

	_, err := NewRunner("10.0.1.1:7777", 9, new(net.Net), s).Register()
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewRunner("10.0.1.2:7777", 7, new(net.Net), s).Register()
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRunner("10.0.1.2:7778", 8, new(net.Net), s).Register()
	if err != nil {
		t.Fatal(err)
	}

	rs, err := RunnersByHost(r.Dir.Snapshot, "10.0.1.2")
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

	go WatchRunnerStart("127.0.0.1", s, ch, errch)

	r := NewRunner(addr, insId, new(net.Net), s)
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

	go WatchRunnerStop("127.0.0.1", s, ch, errch)

	r := NewRunner(addr, insId, new(net.Net), s)
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
