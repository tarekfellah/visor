// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
)

func endpointSetup(srvName string) (s Snapshot, srv *Service) {
	s, err := Dial(DEFAULT_ADDR, "/endpoint-test")
	if err != nil {
		panic(err)
	}

	r, _ := s.conn.Rev()
	s.conn.Del("/", r)
	s = s.FastForward(-1)

	r, err = Init(s)
	if err != nil {
		panic(err)
	}
	s = s.FastForward(r)

	srv = NewService(srvName, s)

	s = s.FastForward(srv.Rev)

	return
}

func TestEndpointRegister(t *testing.T) {
	s, srv := endpointSetup("dahoopz")
	ep, err := NewEndpoint(srv, "1.2.3.4", 1000, s)
	if err != nil {
		t.Error(err)
	}

	ep, err = ep.Register()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.conn.Exists(ep.Path.String())
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Errorf("endpoint %s isn't registered", ep)
	}
}

func TestEndpointUnregister(t *testing.T) {
	s, srv := endpointSetup("megahoopz")
	ep, err := NewEndpoint(srv, "4.3.2.1", 2000, s)
	if err != nil {
		t.Error(err)
	}

	ep, err = ep.Register()
	if err != nil {
		t.Error(err)
	}

	err = ep.Unregister()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.Exists(ep.Path.String())
	if check {
		t.Errorf("endpoint %s is still registered", ep)
	}
}

func TestEndpointGet(t *testing.T) {
	s, srv := endpointSetup("gethoopz")
	ep, err := NewEndpoint(srv, "5.6.7.8", 8000, s)
	if err != nil {
		t.Error(err)
	}

	ep, err = ep.Register()
	if err != nil {
		t.Error(err)
	}

	ep2, err := GetEndpoint(ep.Snapshot, srv, ep.Id())
	if err != nil {
		t.Error(err)
		return
	}

	if ep.Inspect() != ep2.Inspect() {
		t.Errorf("endpoint missmatch")
	}
}
