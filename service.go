// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"path"
	"strconv"
	"time"
)

const (
	ADDRS_PATH    = "addrs"
	SERVICES_PATH = "services"
)

// ServiceAddr represents an entry of a Service and supports all fields to be used
// as SRV record.
type ServiceAddr struct {
	Addr     string
	Priority int
	Port     int
	Target   string
	Weight   int
}

// NewServiceAddr returns a new ServiceAddr.
func NewServiceAddr(addr string, port, prio, weight int) (a *ServiceAddr) {
	a = &ServiceAddr{
		Addr:     addr,
		Target:   addr,
		Priority: prio,
		Port:     port,
		Weight:   weight,
	}

	return
}

func (a *ServiceAddr) Create(s Snapshot, prefix string) (f *File, err error) {
	data := []string{
		strconv.Itoa(a.Priority),
		strconv.Itoa(a.Weight),
		strconv.Itoa(a.Port),
		a.Addr,
	}

	f, err = CreateFile(s, path.Join(prefix, a.Addr), data, new(ListCodec))

	return
}

func GetServiceAddr(s Snapshot, path string) (addr *ServiceAddr, err error) {
	f, err := s.GetFile(path, new(ListCodec))
	if err != nil {
		return
	}

	addr = &ServiceAddr{}
	data := f.Value.([]string)

	p, err := strconv.ParseInt(data[0], 10, 0)
	if err != nil {
		return
	}
	addr.Priority = int(p)

	w, err := strconv.ParseInt(data[1], 10, 0)
	if err != nil {
		return
	}
	addr.Weight = int(w)

	p, err = strconv.ParseInt(data[2], 10, 0)
	if err != nil {
		return
	}
	addr.Port = int(p)

	addr.Addr = data[3] // target
	addr.Target = data[3]

	return
}

type Service struct {
	Path
	Name  string
	Addrs map[string]*ServiceAddr
}

// NewService returns a new Service given a name.
func NewService(name string, snapshot Snapshot) (srv *Service) {
	srv = &Service{Name: name, Addrs: map[string]*ServiceAddr{}}
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
func (s *Service) AddAddr(addr *ServiceAddr) (srv *Service, err error) {
	f, err := addr.Create(s.Snapshot, s.Path.Prefix(ADDRS_PATH))
	if err != nil {
		return
	}

	srv = s.FastForward(f.Rev)

	s.Addrs[addr.Addr] = addr

	return
}

// RemoveAddr removes the given address string from the Service.
func (s *Service) RemoveAddr(addr string) (srv *Service, err error) {
	err = s.Del(path.Join(ADDRS_PATH, addr))
	if err != nil {
		return
	}

	srv = s.FastForward(s.Rev + 1)

	_, ok := s.Addrs[addr]
	if ok {
		delete(s.Addrs, addr)
	}

	return
}

func (s *Service) String() string {
	return fmt.Sprintf("Service<%s>", s.Name)
}

func (s *Service) Inspect() string {
	return fmt.Sprintf("%#v", s)
}

func (s *Service) getAddrs() (addrs map[string]*ServiceAddr, err error) {
	names, err := s.Getdir(s.Path.Prefix(ADDRS_PATH))
	if err != nil {
		if IsErrNoEnt(err) {
			return addrs, nil
		} else {
			return
		}
	}

	addrs = map[string]*ServiceAddr{}

	for _, name := range names {
		var addr *ServiceAddr

		addr, err = GetServiceAddr(s.Snapshot, s.Path.Prefix(path.Join(ADDRS_PATH, name)))
		if err != nil {
			return
		}

		addrs[name] = addr
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

	srv.Addrs = addrs

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
