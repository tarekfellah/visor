// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"path"
	"strconv"
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
	val, rev, err := s.GetBytes(path)
	return string(val), rev, err
}

// GetFile returns the value at the specified path as a file, at this snapshot's revision
func (s Snapshot) GetFile(path string, codec Codec) (*File, error) {
	return Get(s, path, codec)
}

// GetBytes returns the value at the specified path, at this snapshot's revision
func (s Snapshot) GetBytes(path string) ([]byte, int64, error) {
	return s.conn.Get(path, &s.Rev)
}

// Getdir returns the list of files in the specified directory, at this snapshot's revision
func (s Snapshot) Getdir(path string) ([]string, error) {
	return s.conn.Getdir(path, s.Rev)
}

// Set sets the specfied path's body to the passed value, at this snapshot's revision
func (s Snapshot) Set(path string, val string) (Snapshot, error) {
	return s.SetBytes(path, []byte(val))
}

// SetBytes sets the specfied path's body to the passed value, at this snapshot's revision
func (s Snapshot) SetBytes(path string, val []byte) (Snapshot, error) {
	rev, err := s.conn.Set(path, s.Rev, val)
	if err != nil {
		return s, err
	}
	return s.FastForward(rev), err
}

// Del deletes the file at the specified path, at this snapshot's revision
func (s Snapshot) Del(path string) error {
	return s.conn.Del(path, s.Rev)
}

// Update checks if the specified path exists, and if so, does a (*Snapshot).Set with the passed value.
func (s Snapshot) Update(path string, val string) (Snapshot, error) {
	exists, _, err := s.Exists(path)
	if err != nil {
		return s, err
	}
	if !exists {
		return s, NewError(ErrNoEnt, fmt.Sprintf("path '%s' does not exist at %d", path, s.Rev))
	}
	return s.Set(path, val)
}

func (s Snapshot) createSnapshot(rev int64) Snapshotable {
	return Snapshot{rev, s.conn}
}

func (s Snapshot) FastForward(rev int64) (ns Snapshot) {
	return s.fastForward(s, rev).(Snapshot)
}

func (s Snapshot) GetScale(app string, revision string, processName string) (scale int, rev int64, err error) {
	path := path.Join(APPS_PATH, app, REVS_PATH, revision, SCALE_PATH, processName)
	f, err := s.GetFile(path, new(IntCodec))

	// File doesn't exist, assume scale = 0
	if IsErrNoEnt(err) {
		err = nil
		scale = 0
		return
	}

	if err != nil {
		scale = -1
		return
	}

	rev = f.FileRev
	scale = f.Value.(int)

	return
}

func (s Snapshot) SetScale(app string, revision string, processName string, factor int) (rev int64, err error) {
	path := path.Join(APPS_PATH, app, REVS_PATH, revision, SCALE_PATH, processName)
	s1, err := s.Set(path, strconv.Itoa(factor))
	rev = s1.Rev
	return
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
	bytes, rev, err := s.GetBytes(path)
	if err != nil {
		return
	}

	value, err := codec.Decode(bytes)
	if err != nil {
		return
	}

	file = &File{Path: path, Value: value, FileRev: rev, Codec: codec, Snapshot: s}

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
