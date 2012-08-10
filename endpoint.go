// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"strconv"
)

const ENDPOINTS_PATH = "endpoints"

// Endpoint represents an entry of a Service and supports all fields to be used
// as SRV record.
type Endpoint struct {
	Path
	Service  *Service
	Addr     string
	Priority int
	Port     int
	Target   string
	Weight   int
}

func NewEndpoint(srv *Service, addr string, s Snapshot) (e *Endpoint) {
	e = &Endpoint{Addr: addr, Target: addr}
	e.Path = Path{s, srv.Path.Prefix(ENDPOINTS_PATH, addr)}

	return
}

func (e *Endpoint) createSnapshot(rev int64) Snapshotable {
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
	exists, _, err := e.conn.Exists(e.Path.String())
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

	f, err := CreateFile(e.Snapshot, e.Path.String(), data, new(ListCodec))
	if err != nil {
		return
	}

	ep = e.FastForward(f.Rev)

	return
}

// Unregister the endpoint.
func (e *Endpoint) Unregister() error {
	return e.Del("/")
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
	path := srv.Path.Prefix(ENDPOINTS_PATH, addr)

	f, err := s.GetFile(path, new(ListCodec))
	if err != nil {
		return
	}
	data := f.Value.([]string)

	e = &Endpoint{Addr: addr}
	e.Path = Path{s, srv.Path.Prefix(ENDPOINTS_PATH, addr)}

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
