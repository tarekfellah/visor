// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
)

func endpointSetup(srvName string) (s Snapshot, srv *Service) {
	s, err := Dial(DefaultAddr, "/endpoint-test")
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
	ep := NewEndpoint(srv, "1.2.3.4", s)

	ep, err := ep.Register()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.conn.Exists(ep.dir.String())
	if err != nil {
		t.Error(err)
	}
	if !check {
		t.Errorf("endpoint %s isn't registered", ep)
	}
}

func TestEndpointUnregister(t *testing.T) {
	s, srv := endpointSetup("megahoopz")
	ep := NewEndpoint(srv, "4.3.2.1", s)

	ep, err := ep.Register()
	if err != nil {
		t.Error(err)
	}

	err = ep.Unregister()
	if err != nil {
		t.Error(err)
	}

	check, _, err := s.exists(ep.dir.String())
	if check {
		t.Errorf("endpoint %s is still registered", ep)
	}
}

func TestEndpointGet(t *testing.T) {
	s, srv := endpointSetup("gethoopz")
	ep := NewEndpoint(srv, "5.6.7.8", s)

	ep.Port = 8000

	ep, err := ep.Register()
	if err != nil {
		t.Error(err)
	}

	ep2, err := GetEndpoint(ep.Snapshot, srv, ep.Addr)
	if err != nil {
		t.Error(err)
		return
	}

	if ep.Inspect() != ep2.Inspect() {
		t.Errorf("endpoint missmatch")
	}
}
