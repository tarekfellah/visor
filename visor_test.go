// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"testing"
)

const SCALE_PATH_FMT = "apps/%s/revs/%s/scale/%s"

func TestDialWithDefaultAddrAndRoot(t *testing.T) {
	_, err := Dial(DEFAULT_ADDR, DEFAULT_ROOT)
	if err != nil {
		t.Error(err)
	}
}

func TestDialWithInvalidAddr(t *testing.T) {
	_, err := Dial("foo.bar:123:876", "wrong")
	if err == nil {
		t.Error("Dialed with invalid addr")
	}
}

func TestScaleUp(t *testing.T) {
	s, err := Dial(DEFAULT_ADDR, "/scale-test")
	if err != nil {
		panic(err)
	}
	s.conn.Del("/", s.Rev)
	s = s.FastForward(-1)

	s.conn.Set("/apps/dog/revs/master/file", -1, []byte{})
	s.conn.Set("/apps/dog/procs/lol", -1, []byte{})

	err = Scale("dog", "master", "lol", 5, s)
	if err != nil {
		t.Error(err)
	}

	factor, _, err := s.conn.Get(fmt.Sprintf(SCALE_PATH_FMT, "dog", "master", "lol"), nil)
	if err != nil {
		t.Error(err)
	}
	if string(factor) != "5" {
		t.Errorf("scaling factor expected %s, got %s", "5", factor)
	}

	tickets, err := s.conn.Getdir(TICKETS_PATH, s.FastForward(-1).Rev)
	if err != nil {
		t.Error(err)
	}
	if len(tickets) != 5 {
		t.Errorf("Expected tickets %s, got %d", "5", len(tickets))
	}
}

func TestScaleDown(t *testing.T) {
	s, err := Dial(DEFAULT_ADDR, "/scale-test")
	if err != nil {
		panic(err)
	}
	s.conn.Del("/", s.Rev)
	s = s.FastForward(-1)

	s.conn.Set("/apps/cat/revs/master/file", -1, []byte{})
	s.conn.Set("/apps/cat/procs/lol", -1, []byte{})

	p := fmt.Sprintf(SCALE_PATH_FMT, "cat", "master", "lol")
	r, err := s.conn.Set(p, s.Rev, []byte("5"))
	s = s.FastForward(r)

	err = Scale("cat", "master", "lol", -1, s)
	if err == nil {
		t.Error("Should return an error on a non-positive scaling factor")
	}

	err = Scale("cat", "master", "lol", 2, s)
	if err != nil {
		t.Error(err)
	}

	factor, _, err := s.conn.Get(p, nil)
	if err != nil {
		t.Error(err)
	}
	if string(factor) != "2" {
		t.Errorf("Scaling factor expected %s, got %s", "2", factor)
	}

	tickets, err := s.conn.Getdir(TICKETS_PATH, s.FastForward(-1).Rev)
	if err != nil {
		t.Error(err)
	}
	if len(tickets) != 3 {
		t.Errorf("Expected tickets %s, got %d", "3", len(tickets))
	}
}
