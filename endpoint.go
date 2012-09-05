// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"net"
	"strconv"
)

const ENDPOINTS_PATH = "endpoints"

// Endpoint represents an entry of a Service and supports all fields to be used
// as SRV record.
type Endpoint struct {
	dir
	Service  *Service
	Addr     string
	Priority int
	Port     int
	Target   string
	Weight   int
}

func NewEndpoint(srv *Service, addr string, s Snapshot) (e *Endpoint) {
	e = &Endpoint{Addr: addr, Target: addr}
	e.dir = dir{s, srv.dir.prefix(ENDPOINTS_PATH, addr)}

	return
}

func (e *Endpoint) createSnapshot(rev int64) snapshotable {
	tmp := *e
	tmp.Snapshot = Snapshot{rev, e.conn}
	return &tmp
}

// FastForward advances the endpoint in time. It returns
// a new instance of Endpoint with the supplied revision.
func (e *Endpoint) FastForward(rev int64) *Endpoint {
	return e.Snapshot.fastForward(e, rev).(*Endpoint)
}

// Register the endpoint.
func (e *Endpoint) Register() (ep *Endpoint, err error) {
	if net.ParseIP(e.Addr) == nil {
		return nil, fmt.Errorf("addr %s is not a valide IP", e.Addr)
	}

	exists, _, err := e.conn.Exists(e.dir.String())
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	data := []string{
		strconv.Itoa(e.Priority),
		strconv.Itoa(e.Weight),
		strconv.Itoa(e.Port),
		e.Addr,
	}

	f, err := CreateFile(e.Snapshot, e.dir.String(), data, new(ListCodec))
	if err != nil {
		return
	}

	ep = e.FastForward(f.Rev)

	return
}

// Unregister the endpoint.
func (e *Endpoint) Unregister() error {
	return e.del("/")
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("Endpoint<%s>", e.Addr)
}

func (e *Endpoint) Inspect() string {
	return fmt.Sprintf("%#v", e)
}

// GetEndpoint fetches the endpoint for the given service and addr from the global
// registry.
func GetEndpoint(s Snapshot, srv *Service, addr string) (e *Endpoint, err error) {
	path := srv.dir.prefix(ENDPOINTS_PATH, addr)

	f, err := s.getFile(path, new(ListCodec))
	if err != nil {
		return
	}
	data := f.Value.([]string)

	e = &Endpoint{Addr: addr}
	e.dir = dir{s, srv.dir.prefix(ENDPOINTS_PATH, addr)}

	p, err := strconv.ParseInt(data[0], 10, 0)
	if err != nil {
		return
	}
	e.Priority = int(p)

	w, err := strconv.ParseInt(data[1], 10, 0)
	if err != nil {
		return
	}
	e.Weight = int(w)

	p, err = strconv.ParseInt(data[2], 10, 0)
	if err != nil {
		return
	}
	e.Port = int(p)
	e.Target = data[3]

	e = e.FastForward(f.FileRev)

	return
}
