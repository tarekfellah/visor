// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const endpointsPath = "endpoints"

// Endpoint represents an entry of a Service and supports all fields to be used
// as SRV record.
type Endpoint struct {
	dir
	Service  *Service
	Addr     string
	IP       string
	Priority int
	Port     int
	Target   string
	Weight   int
}

func NewEndpoint(srv *Service, addr string, port int, s Snapshot) (e *Endpoint, err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return
	}

	e = &Endpoint{
		Addr: addr,
		IP:   tcpAddr.IP.String(),
		Port: tcpAddr.Port,
	}
	e.dir = dir{s, srv.dir.prefix(endpointsPath, e.Id())}

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
	exists, _, err := e.conn.Exists(e.dir.String())
	if err != nil {
		return
	}
	if exists {
		return nil, ErrKeyConflict
	}

	data := []string{
		e.Addr,
		e.IP,
		strconv.Itoa(e.Port),
		strconv.Itoa(e.Priority),
		strconv.Itoa(e.Weight),
	}

	f, err := createFile(e.Snapshot, e.dir.String(), data, new(listCodec))
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

func (e *Endpoint) Id() string {
	return EndpointId(e.Addr, e.Port)
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("Endpoint<%s>", e.Id())
}

func (e *Endpoint) Inspect() string {
	return fmt.Sprintf("%#v", e)
}

// GetEndpoint fetches the endpoint for the given service and id from the global
// registry.
func GetEndpoint(s Snapshot, srv *Service, id string) (e *Endpoint, err error) {
	path := srv.dir.prefix(endpointsPath, id)

	f, err := s.getFile(path, new(listCodec))
	if err != nil {
		return
	}
	data := f.Value.([]string)

	e = &Endpoint{Addr: data[0], IP: data[1]}
	e.dir = dir{s, srv.dir.prefix(endpointsPath, id)}

	p, err := strconv.ParseInt(data[2], 10, 0)
	if err != nil {
		return
	}
	e.Port = int(p)

	p, err = strconv.ParseInt(data[3], 10, 0)
	if err != nil {
		return
	}
	e.Priority = int(p)

	w, err := strconv.ParseInt(data[4], 10, 0)
	if err != nil {
		return
	}
	e.Weight = int(w)

	e = e.FastForward(f.FileRev)

	return
}

// EndpointId returns a proper Id for the given addr & port
func EndpointId(addr string, port int) string {
	return fmt.Sprintf("%s-%d", strings.Replace(addr, ".", "-", -1), port)
}
