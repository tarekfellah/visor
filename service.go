// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"path"
)

const servicesPath = "services"

type Service struct {
	Dir  dir
	Name string
}

// NewService returns a new Service given a name.
func NewService(name string, snapshot Snapshot) (srv *Service) {
	srv = &Service{Name: name}
	srv.Dir = dir{snapshot, path.Join(servicesPath, srv.Name)}

	return
}

func (s *Service) createSnapshot(rev int64) snapshotable {
	tmp := *s
	tmp.Dir.Snapshot = Snapshot{rev, s.Dir.Snapshot.conn}
	return &tmp
}

// FastForward advances the service in time. It returns
// a new instance of Service with the supplied revision.
func (s *Service) FastForward(rev int64) *Service {
	return s.Dir.Snapshot.fastForward(s, rev).(*Service)
}

// Register adds the Service to the global process state.
func (s *Service) Register() (srv *Service, err error) {
	exists, _, err := s.Dir.Snapshot.conn.Exists(s.Dir.Name)
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	rev, err := s.Dir.set("registered", timestamp())
	if err != nil {
		return
	}

	srv = s.FastForward(rev)

	return
}

// Unregister removes the Service form the global process state.
func (s *Service) Unregister() error {
	return s.Dir.del("/")
}

func (s *Service) String() string {
	return fmt.Sprintf("Service<%s>", s.Name)
}

func (s *Service) Inspect() string {
	return fmt.Sprintf("%#v", s)
}

func (s *Service) GetEndpoints() (endpoints []*Endpoint, err error) {
	p := s.Dir.prefix(endpointsPath)

	exists, _, err := s.Dir.Snapshot.conn.Exists(p)
	if err != nil || !exists {
		return
	}

	ids, err := s.Dir.Snapshot.getdir(p)
	if err != nil {
		return
	}

	for _, id := range ids {
		var e *Endpoint

		e, err = GetEndpoint(s.Dir.Snapshot, s, id)
		if err != nil {
			return
		}

		endpoints = append(endpoints, e)
	}

	return
}

// GetService fetches a service with the given name.
func GetService(s Snapshot, name string) (srv *Service, err error) {
	service := NewService(name, s)

	exists, _, err := s.conn.Exists(service.Dir.Name)
	if err != nil && !exists {
		return
	}

	srv = service

	return
}

// Services returns the list of all registered Services.
func Services(s Snapshot) (srvs []*Service, err error) {
	exists, _, err := s.conn.Exists(servicesPath)
	if err != nil || !exists {
		return
	}

	names, err := s.getdir(servicesPath)
	if err != nil {
		return
	}

	for _, name := range names {
		var srv *Service

		srv, err = GetService(s, name)
		if err != nil {
			return
		}

		srvs = append(srvs, srv)
	}

	return
}
