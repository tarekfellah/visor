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

// GetScale returns the scale of an app:proc@rev tuple. If the scale isn't found, 0 is returned.
func (s Snapshot) GetScale(app string, revision string, processName string) (scale int, rev int64, err error) {
	// NOTE: This method does not check whether or not the scale target exists.
	// A scale of `0` will be returned if any of the path components are missing.
	// This is to avoid having to set the /apps/<app>/revs/<rev>/scale/<proc> paths to 0
	// when registering revisions.
	path := path.Join(APPS_PATH, app, REVS_PATH, revision, SCALE_PATH, processName)
	f, err := s.getFile(path, new(IntCodec))

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

// SetScale sets the scale of an app:proc@rev tuple to the specified value.
func (s Snapshot) SetScale(app string, revision string, processName string, factor int) (s1 Snapshot, err error) {
	path := path.Join(APPS_PATH, app, REVS_PATH, revision, SCALE_PATH, processName)
	return s.set(path, strconv.Itoa(factor))
}

// Getuid returns a unique ID from the coordinator
func Getuid(s Snapshot) (int64, error) {
	return s.conn.Set(UID_PATH, -1, []byte{})
}

// exists checks if the specified path exists at this snapshot's revision
func (s Snapshot) exists(path string) (bool, int64, error) {
	return s.conn.ExistsRev(path, &s.Rev)
}

// get returns the value at the specified path, at this snapshot's revision
func (s Snapshot) get(path string) (string, int64, error) {
	val, rev, err := s.getBytes(path)
	return string(val), rev, err
}

// getFile returns the value at the specified path as a file, at this snapshot's revision
func (s Snapshot) getFile(path string, codec Codec) (*File, error) {
	bytes, rev, err := s.getBytes(path)
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

// getBytes returns the value at the specified path, at this snapshot's revision
func (s Snapshot) getBytes(path string) ([]byte, int64, error) {
	return s.conn.Get(path, &s.Rev)
}

// getdir returns the list of files in the specified directory, at this snapshot's revision
func (s Snapshot) getdir(path string) ([]string, error) {
	return s.conn.Getdir(path, s.Rev)
}

// set sets the specfied path's body to the passed value, at this snapshot's revision
func (s Snapshot) set(path string, val string) (Snapshot, error) {
	return s.setBytes(path, []byte(val))
}

// setBytes sets the specfied path's body to the passed value, at this snapshot's revision
func (s Snapshot) setBytes(path string, val []byte) (Snapshot, error) {
	rev, err := s.conn.Set(path, s.Rev, val)
	if err != nil {
		return s, err
	}
	return s.FastForward(rev), err
}

// del deletes the file at the specified path, at this snapshot's revision
func (s Snapshot) del(path string) error {
	return s.conn.Del(path, s.Rev)
}

// update checks if the specified path exists, and if so, does a (*Snapshot).Set with the passed value.
func (s Snapshot) update(path string, val string) (Snapshot, error) {
	exists, _, err := s.exists(path)
	if err != nil {
		return s, err
	}
	if !exists {
		return s, NewError(ErrNoEnt, fmt.Sprintf("path '%s' does not exist at %d", path, s.Rev))
	}
	return s.set(path, val)
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

// getLatest returns the latest value for the given path
func getLatest(s Snapshot, path string, codec Codec) (file *File, err error) {
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
