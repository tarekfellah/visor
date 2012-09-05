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
type File struct {
	Snapshot
	FileRev int64 // File rev, or 0 if path doesn't exist
	Path    string
	Value   interface{}
	Codec   Codec
}

func CreateFile(snapshot Snapshot, path string, value interface{}, codec Codec) (*File, error) {
	file := &File{Path: path, Value: value, Codec: codec, Snapshot: snapshot, FileRev: -1}
	return file.Create()
}

func (f *File) createSnapshot(rev int64) (file Snapshotable) {
	tmp := *f
	tmp.Snapshot = Snapshot{rev, f.conn}
	return &tmp
}

// FastForward advances the file in time. It returns
// a new instance of File with the supplied revision.
func (f *File) FastForward(rev int64) *File {
	if rev == -1 {
		var err error
		_, rev, err = f.conn.Stat(f.Path)
		if err != nil {
			return f
		}
	}
	return f.Snapshot.fastForward(f, rev).(*File)
}

// Del deletes a file
func (f *File) Del() error {
	return f.Snapshot.del(f.Path)
}

// Create creates a file from its Value attribute
func (f *File) Create() (*File, error) {
	return f.Set(f.Value)
}

// Set sets the value at this file's path to a new value.
func (f *File) Set(value interface{}) (file *File, err error) {
	bytes, err := f.Codec.Encode(value)
	if err != nil {
		return
	}

	s, err := f.setBytes(f.Path, bytes)

	if s.Rev > 0 {
		file = f.FastForward(s.Rev)
	} else {
		file = f
	}

	if err != nil {
		return
	}
	file.Value = value
	file.FileRev = s.Rev

	return
}

func (f *File) String() string {
	return fmt.Sprintf("%#v", f)
}
