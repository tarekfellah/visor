// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
)

func fileSetup(path string, value interface{}) *File {
	s, err := Dial(DEFAULT_ADDR, "/file-test")
	if err != nil {
		panic(err)
	}

	r, _ := s.conn.Rev()
	err = s.conn.Del("/", r)

	file := &File{Path: path, Value: value, Codec: new(ByteCodec), Snapshot: s.FastForward(-1)}

	return file
}

func TestSet(t *testing.T) {
	path := "update-path"
	value := "update-val"

	f := fileSetup(path, value)

	rev, _ := f.conn.Set(path, f.Snapshot.Rev, []byte(value))
	f = f.FastForward(rev)

	f, err := f.Set([]byte(value + "!"))
	if err != nil {
		t.Error(err)
		return
	}

	if string(f.Value.([]byte)) != value+"!" {
		t.Errorf("expected (*File).Value to be update, got %s", string(f.Value.([]byte)))
	}

	val, _, err := f.conn.Get(path, &f.Rev)
	if err != nil {
		t.Error(err)
	}

	if string(val) != value+"!" {
		t.Errorf("expected %s got %s", value+"!", val)
	}
}

func TestFastForward(t *testing.T) {
	path := "ff-path"
	value := "ff-val"

	f := fileSetup(path, value)

	s, err := f.Snapshot.set(path, value)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = f.conn.Set(path+"-1", s.Rev, []byte(value))
	if err != nil {
		t.Error(err)
		return
	}

	f = f.FastForward(-1)
	if f.Snapshot.Rev != s.Rev {
		t.Errorf("expected %d got %d", s.Rev, f.Rev)
	}
}

func TestSetConflict(t *testing.T) {
	path := "conflict-path"
	value := "conflict-val"

	f := fileSetup(path, value)

	s, _ := f.Snapshot.set(path, value)
	f = f.FastForward(s.Rev)

	_, err := f.Set([]byte(value + "!"))
	if err != nil {
		t.Error(err)
		return
	}

	_, err = f.Set([]byte("!"))
	if err == nil {
		t.Error("expected update with old revision to fail")
		return
	}
}

func TestDel(t *testing.T) {
	path := "del-path"
	value := "del-val"

	f := fileSetup(path, value)

	_, err := f.Snapshot.setBytes(path, []byte{})
	if err != nil {
		t.Error(err)
	}
	exists, _, err := f.conn.Exists(path)
	if !exists {
		t.Error("path wasn't set properly")
		return
	}

	f = f.FastForward(-1)

	err = f.Del()
	if err != nil {
		t.Error(err)
	}

	exists, _, err = f.conn.Exists(path)
	if err != nil {
		t.Error(err)
	}

	if exists {
		t.Error("path wasn't deleted")
	}
}
