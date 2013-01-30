// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"path"
	"regexp"
)

// Snapshot represents a specific point in time
// within the coordinator state. It is used by
// all time-aware interfaces to the coordinator.
type Snapshot struct {
	Rev  int64
	conn *conn
}

// regular expressions to validate paths
const charPat = `[-.[:alnum:]]`

var pathRe = regexp.MustCompile(`^/$|^(/` + charPat + `+)+$`)

// Snapshotable is implemented by any type which
// is time-aware, and can be moved forward in time
// by calling createSnapshot with a new revision.
type snapshotable interface {
	createSnapshot(rev int64) snapshotable
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

	s = Snapshot{rev, &conn{addr, root, dconn}}
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

	s = Snapshot{rev, &conn{uri, root, dconn}}
	return
}

// GetScale returns the scale of an app:pty@rev tuple. If the scale isn't found, 0 is returned.
func (s Snapshot) GetScale(app string, revid string, pty string) (scale int, rev int64, err error) {
	path := ptyInstancesPath(app, revid, pty)
	count, rev, err := s.conn.Stat(path, &s.Rev)

	// File doesn't exist, assume scale = 0
	if IsErrNoEnt(err) {
		return 0, rev, nil
	}

	if err != nil {
		return -1, rev, err
	}

	return count, rev, nil
}

// GetProxies gets the list of bazooka-proxy service IPs
func (s Snapshot) GetProxies() ([]string, error) {
	return s.getdir(proxyDir)
}

// GetPms gets the list of bazooka-pm service IPs
func (s Snapshot) GetPms() ([]string, error) {
	return s.getdir(pmDir)
}

func (s Snapshot) RegisterPm(host, version string) (Snapshot, error) {
	return s.set(path.Join(pmDir, host), timestamp()+" "+version)
}

func (s Snapshot) UnregisterPm(host string) error {
	return s.del(path.Join(pmDir, host))
}

func (s Snapshot) RegisterProxy(host string) (Snapshot, error) {
	return s.set(path.Join(proxyDir, host), timestamp())
}

func (s Snapshot) UnregisterProxy(host string) error {
	return s.del(path.Join(proxyDir, host))
}

func (s Snapshot) ResetCoordinator() error {
	err := s.del("/")
	if IsErrNoEnt(err) {
		return nil
	}
	return err
}

// Getuid returns a unique ID from the coordinator
func Getuid(s Snapshot) (int64, error) {
	return s.conn.Set(uidPath, -1, []byte{})
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
func (s Snapshot) getFile(path string, codec codec) (f *file, err error) {
	f = newFile(s, path, nil, codec)

	bytes, rev, err := s.getBytes(path)
	if err != nil {
		return
	}

	value, err := codec.Decode(bytes)
	if err != nil {
		return
	}
	f.Value = value
	f.Snapshot = f.Snapshot.FastForward(rev)

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

func (s Snapshot) createSnapshot(rev int64) snapshotable {
	return Snapshot{rev, s.conn}
}

func (s Snapshot) FastForward(rev int64) (ns Snapshot) {
	return s.fastForward(s, rev).(Snapshot)
}

// fastForward either calls *createSnapshot* on *obj* or returns *obj* if it
// can't advance the object in time. Note that fastForward can never fail.
func (s *Snapshot) fastForward(obj snapshotable, rev int64) snapshotable {
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
func getLatest(s Snapshot, path string, codec codec) (f *file, err error) {
	evalue, rev, err := s.conn.Get(path, nil)
	if err != nil {
		return
	}

	value, err := codec.Decode(evalue)
	if err != nil {
		return
	}

	f = &file{dir: path, Value: value, codec: codec, Snapshot: s.FastForward(rev)}

	return
}

func getSnapshotables(list []string, fn func(string) (snapshotable, error)) (chan snapshotable, chan error) {
	ch := make(chan snapshotable, len(list))
	errch := make(chan error, len(list))

	for _, item := range list {
		go func(i string) {
			r, err := fn(i)
			if err != nil {
				errch <- err
			} else {
				ch <- r
			}
		}(item)
	}
	return ch, errch
}
