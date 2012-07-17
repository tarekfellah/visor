// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
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

// Exists checks if the specified path exists at this snapshot's revision
func (s Snapshot) Exists(path string) (bool, int64, error) {
	return s.conn.ExistsRev(path, &s.Rev)
}

// Get returns the value at the specified path, at this snapshot's revision
func (s Snapshot) Get(path string) (string, int64, error) {
	val, rev, err := s.conn.Get(path, &s.Rev)
	return string(val), rev, err
}

// Getdir returns the list of files in the specified directory, at this snapshot's revision
func (s Snapshot) Getdir(path string) ([]string, error) {
	return s.conn.Getdir(path, s.Rev)
}

// Set sets the specfied path's body to the passed value, at this snapshot's revision
func (s Snapshot) Set(path string, val string) (int64, error) {
	return s.conn.Set(path, s.Rev, []byte(val))
}

// Del deletes the file at the specified path, at this snapshot's revision
func (s Snapshot) Del(path string) error {
	return s.conn.Del(path, s.Rev)
}

// Update checks if the specified path exists, and if so, does a (*Snapshot).Set with the passed value.
func (s Snapshot) Update(path string, val string) (rev int64, err error) {
	exists, rev, err := s.Exists(path)
	if err != nil {
		return
	}
	if !exists {
		return 0, fmt.Errorf("path %s doesn't exist", path)
	}
	return s.Set(path, val)
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

// GetLatest returns the latest value for the given path
func GetLatest(s Snapshot, path string, codec Codec) (file *File, err error) {
	evalue, rev, err := s.conn.Get(path, nil)
	if err != nil {
		return
	}

	value, err := codec.Decode(evalue)
	if err != nil {
		return
	}

	file = &File{Path: path, Value: value, Codec: codec, Snapshot: s.FastForward(rev)}

	return
}

// Getuid returns a unique ID from the coordinator
func Getuid(s Snapshot) (rev int64, err error) {
	for {
		rev, err = s.conn.Set("/uid", -1, []byte{})
		if err != doozer.ErrOldRev {
			break
		}
	}
	return
}
