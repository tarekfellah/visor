// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import "testing"

func connSetup() (*conn, int64) {
	s, err := Dial(DefaultAddr, "/conn-test")
	if err != nil {
		panic(err)
	}

	err = s.conn.Del("/", s.Rev)
	r, err := s.conn.Rev()
	if err != nil {
		panic(err)
	}

	return s.conn, r
}

func TestConnDifferentRoot(t *testing.T) {
	body := "test"

	s, _ := Dial(DefaultAddr, "/not-conn-test")

	_, err := s.conn.Set("root", s.Rev, []byte(body))
	if err != nil {
		t.Error(err)
	}

	b, _, err := s.conn.conn.Get("/not-conn-test/root", nil)
	if err != nil {
		t.Error(err)
	}
	if string(b) != body {
		t.Errorf("Expected %s got %s", body, string(b))
	}
}

func TestConnExists(t *testing.T) {
	c, rev := connSetup()

	k := "key"
	v := "value"

	exists, _, _ := c.Exists(k)
	if exists {
		t.Errorf("path '%s' shouldn't exist", k)
	}

	_, err := c.Set(k, rev, []byte(v))
	if err != nil {
		panic(err)
	}
	exists, rev1, err := c.Exists(k)
	if !exists {
		t.Errorf("path %s should exist at latest rev", k)
	}

	exists, _, err = c.ExistsRev(k, &rev1)
	if !exists {
		t.Errorf("path %s should exist at rev %d", k, rev1)
	}

	exists, rev2, err := c.ExistsRev(k, &rev)
	if exists {
		t.Errorf("path %s shouldn't exist yet", k)
	}
	if rev2 != 0 {
		t.Errorf("rev for %s should = 0", k)
	}
}
