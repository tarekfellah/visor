// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"github.com/soundcloud/doozer"
)

// Snapshot represents a specific point in time
// within the coordinator state. It is used by
// all time-aware interfaces to the coordinator.
type Snapshot struct {
	Rev  int64
	conn *Conn
}

// Snapshotable is implemented by any type which
// is time-aware, and can be moved forward in time
// by calling createSnapshot with a new revision.
type Snapshotable interface {
	createSnapshot(rev int64) Snapshotable
}

// Dial calls doozer.Dial and returns a Snapshot of the coordinator
// at the latest revision.
func Dial(addr string, root string) (s Snapshot, err error) {
	dconn, err := doozer.Dial(addr)
	if err != nil {
		return
	}

	rev, err := dconn.Rev()
	if err != nil {
		return
	}

	s = Snapshot{rev, &Conn{addr, root, dconn}}
	return
}

// DialUri calls doozer.DialUri and returns a Snapshot of the coordinator cluster
// at the latest revision.
func DialUri(uri string, root string) (s Snapshot, err error) {
	dconn, err := doozer.DialUri(uri, "")
	if err != nil {
		return
	}

	rev, err := dconn.Rev()
	if err != nil {
		return
	}

	s = Snapshot{rev, &Conn{uri, root, dconn}}
	return
}

func (s Snapshot) Conn() *Conn {
	return s.conn
}

func (s Snapshot) createSnapshot(rev int64) Snapshotable {
	return Snapshot{rev, s.conn}
}

func (s Snapshot) FastForward(rev int64) (ns Snapshot) {
	return s.fastForward(s, rev).(Snapshot)
}

// fastForward either calls *createSnapshot* on *obj* or returns *obj* if it
// can't advance the object in time. Note that fastForward can never fail.
func (s *Snapshot) fastForward(obj Snapshotable, rev int64) Snapshotable {
	var err error

	if rev == -1 {
		rev, err = s.conn.Rev()
		if err != nil {
			return obj
		}
	} else if rev < s.Rev {
		return obj
	}
	return obj.createSnapshot(rev)
}

func Set(s Snapshot, path string, value interface{}, codec Codec) (snapshot Snapshot, err error) {
	evalue, err := codec.Encode(value)

	revision, err := s.conn.Set(path, s.Rev, evalue)

	snapshot = s.FastForward(revision)

	return
}

// Get returns the value for the given path
func Get(s Snapshot, path string, codec Codec) (file *File, err error) {
	evalue, _, err := s.conn.Get(path, &s.Rev)
	if err != nil {
		return
	}

	value, err := codec.Decode(evalue)
	if err != nil {
		return
	}

	file = &File{Path: path, Value: value, Codec: codec, Snapshot: s}

	return
}
