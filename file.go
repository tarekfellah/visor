// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
)

// File represents a coordinator file
// at a specific point in time.
type file struct {
	Snapshot
	dir   string
	Value interface{}
	codec codec
}

func createFile(snapshot Snapshot, path string, value interface{}, codec codec) (*file, error) {
	file := &file{dir: path, Value: value, codec: codec, Snapshot: snapshot}
	return file.Create()
}

func (f *file) createSnapshot(rev int64) (file snapshotable) {
	tmp := *f
	tmp.Snapshot = Snapshot{rev, f.Snapshot.conn}
	return &tmp
}

// FastForward advances the file in time. It returns
// a new instance of File with the supplied revision.
func (f *file) FastForward(rev int64) *file {
	if rev == -1 {
		var err error
		_, rev, err = f.Snapshot.conn.Stat(f.dir, nil)
		if err != nil {
			return f
		}
	}
	return f.Snapshot.fastForward(f, rev).(*file)
}

// Del deletes a file
func (f *file) Del() error {
	return f.Snapshot.del(f.dir)
}

// Create creates a file from its Value attribute
// TODO: Rename to 'Save'
func (f *file) Create() (*file, error) {
	return f.Set(f.Value)
}

// Set sets the value at this file's path to a new value.
func (f *file) Set(value interface{}) (file *file, err error) {
	bytes, err := f.codec.Encode(value)
	if err != nil {
		return
	}

	s, err := f.Snapshot.setBytes(f.dir, bytes)

	if s.Rev > 0 {
		file = f.FastForward(s.Rev)
	} else {
		file = f
	}

	if err != nil {
		return
	}
	file.Value = value
	file.Snapshot = s

	return
}

func (f *file) String() string {
	return fmt.Sprintf("%#v", f)
}
