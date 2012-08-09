// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
)

func serviceSetup(name string) (srv *Service) {
	s, err := Dial(DEFAULT_ADDR, "/service-test")
	if err != nil {
		panic(err)
	}

	r, _ := s.conn.Rev()
	err = s.conn.Del(SERVICES_PATH, r)
	rev, err := Init(s)
	if err != nil {
		panic(err)
	}

	srv = NewService(name, s)
	srv = srv.FastForward(rev)

	return
}

func TestServiceRegistration(t *testing.T) {
	srv := serviceSetup("fancydb")

	check, _, err := srv.conn.Exists(srv.Path.Dir)
	if err != nil {
		t.Error(err)
		return
	}
	if check {
		t.Error("Service already registered")
		return
	}

	srv2, err := srv.Register()
	if err != nil {
		t.Error(err)
		return
	}
	check, _, err = srv2.conn.Exists(srv2.Path.Dir)
	if err != nil {
		t.Error(err)
		return
	}
	if !check {
		t.Error("Service registration failed")
		return
	}
	_, err = srv.Register()
	if err == nil {
		t.Error("Service allowed to be registered twice")
	}
	_, err = srv2.Register()
	if err == nil {
		t.Error("Service allowed to be registered twice")
	}
}

func TestServiceUnregistration(t *testing.T) {
	srv := serviceSetup("broker")

	srv, err := srv.Register()
	if err != nil {
		t.Error(err)
		return
	}

	err = srv.Unregister()
	if err != nil {
		t.Error(err)
		return
	}

	check, _, err := srv.conn.Exists(srv.Path.Dir)
	if err != nil {
		t.Error(err)
	}
	if check {
		t.Error("srv still registered")
	}
}

func TestServices(t *testing.T) {
	var err error

	srv := serviceSetup("memstore")
	names := []string{"boombroker", "comastorage", "lulzdb"}

	for _, name := range names {
		srv = NewService(name, srv.Snapshot)
		srv, err = srv.Register()
		if err != nil {
			t.Error(err)
		}
	}

	srvs, err := Services(srv.Snapshot)
	if err != nil {
		t.Error(err)
	}

	if len(srvs) != len(names) {
		t.Errorf("expected length %d returned %d", len(names), len(srvs))
	} else {
		for i, name := range names {
			if srvs[i].Name != name {
				t.Errorf("expected %s got %s", name, srvs[i].Name)
			}
		}
	}
}
