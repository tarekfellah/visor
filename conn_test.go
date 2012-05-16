// Copyright (c) 2012, SoundCloud Ltd., Alexis Sellier, Alexander Simmerl, Daniel Bornkessel, Tomás Senart
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import "testing"

func connSetup(name string) (conn *Conn) {
	s, err := Dial(DEFAULT_ADDR, DEFAULT_ROOT)
	if err != nil {
		panic(err)
	}

	r, _ := s.conn.Rev()
	err = s.conn.Del("/", r)

	return s.conn
}

func TestDifferentRoot(t *testing.T) {
	body := "test"

	s, _ := Dial(DEFAULT_ADDR, "/notvisor")

	_, err := s.conn.Set("root", s.Rev, []byte(body))
	if err != nil {
		t.Error(err)
	}

	b, _, err := s.conn.conn.Get("/notvisor/root", nil)
	if err != nil {
		t.Error(err)
	}
	if string(b) != body {
		t.Errorf("Expected %s got %s", body, string(b))
	}
}

func TestExists(t *testing.T) {
	path := "exists-test"
	conn := connSetup(path)

	exists, _, err := conn.Exists(path, nil)
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Error("path shouldn't exist")
	}

	_, err = conn.Set(path+"/key", 0, []byte{})
	if err != nil {
		t.Error(err)
	}

	exists, _, err = conn.Exists(path+"/key", nil)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Error("path doesn't exist")
	}

	exists, _, err = conn.Exists(path, nil)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Error("path doesn't exist")
	}
}
