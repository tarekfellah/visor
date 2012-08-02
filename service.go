// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"path"
	"time"
)

const (
	ADDRS_PATH    = "addrs"
	SERVICES_PATH = "services"
)

type Service struct {
	Path
	Name  string
	Addrs map[string]bool
}

// NewService returns a new Service given a name.
func NewService(name string, snapshot Snapshot) (srv *Service) {
	srv = &Service{Name: name, Addrs: map[string]bool{}}
	srv.Path = Path{snapshot, path.Join(SERVICES_PATH, srv.Name)}

	return
}

func (s *Service) createSnapshot(rev int64) Snapshotable {
	tmp := *s
	tmp.Snapshot = Snapshot{rev, s.conn}
	return &tmp
}

// FastForward advances the service in time. It returns
// a new instance of Service with the supplied revision.
func (s *Service) FastForward(rev int64) *Service {
	return s.Snapshot.fastForward(s, rev).(*Service)
}

// Register adds the Service to the global process state.
func (s *Service) Register() (srv *Service, err error) {
	exists, _, err := s.conn.Exists(s.Path.Dir)
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	rev, err := s.Set("registered", time.Now().UTC().String())
	if err != nil {
		return
	}

	srv = s.FastForward(rev)

	return
}

// Unregister removes the Service form the global process state.
func (s *Service) Unregister() error {
	return s.Del("/")
}

// AddAddr adds the given address string to the Service.
func (s *Service) AddAddr(addr string) (srv *Service, err error) {
	s.Addrs[addr] = true

	rev, err := s.Set(path.Join(ADDRS_PATH, addr), time.Now().UTC().String())
	if err != nil {
		return
	}

	srv = s.FastForward(rev)

	return
}

// RemoveAddr removes the given address string from the Service.
func (s *Service) RemoveAddr(addr string) (srv *Service, err error) {
	_, ok := s.Addrs[addr]
	if ok {
		delete(s.Addrs, addr)
	}

	err = s.Del(path.Join(ADDRS_PATH, addr))
	if err != nil {
		return
	}

	srv = s.FastForward(s.Rev + 1)

	return
}

func (s *Service) String() string {
	return fmt.Sprintf("Service<%s>", s.Name)
}

func (s *Service) Inspect() string {
	return fmt.Sprintf("%#v", s)
}

func (s *Service) getAddrs() (addrs []string, err error) {
	addrs, err = s.Getdir(s.Path.Prefix(ADDRS_PATH))
	if err != nil && IsErrNoEnt(err) {
		return addrs, nil
	}

	return
}

// GetService fetches a service with the given name.
func GetService(s Snapshot, name string) (srv *Service, err error) {
	srv = NewService(name, s)

	exists, _, err := s.conn.Exists(srv.Path.Dir)
	if err != nil && !exists {
		return
	}

	addrs, err := srv.getAddrs()
	if err != nil {
		return
	}

	for _, addr := range addrs {
		srv.Addrs[addr] = true
	}

	return
}

// Services returns the list of all registered Services.
func Services(s Snapshot) (srvs []*Service, err error) {
	exists, _, err := s.conn.Exists(SERVICES_PATH)
	if err != nil || !exists {
		return
	}

	names, err := s.Getdir(SERVICES_PATH)
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
