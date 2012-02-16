package visor

import (
	"fmt"
	"reflect"
)

// File represents a coordinator file
// at a specific point in time.
type File struct {
	Snapshot
	Path  string
	Value reflect.Value
	Codec Codec
}

// NewFile returns a new file object.
func NewFile(path string, value reflect.Value, codec Codec, snapshot Snapshot) *File {
	f := &File{Path: path, Value: value, Codec: codec, Snapshot: snapshot}
	return f
}

func (f *File) createSnapshot(rev int64) (file Snapshotable) {
	file = &File{Path: f.Path, Value: f.Value, Codec: f.Codec, Snapshot: Snapshot{rev, f.conn}}
	return
}

// FastForward advances the file in time. It returns
// a new instance of File with the supplied revision.
func (f *File) FastForward(rev int64) *File {
	return f.Snapshot.fastForward(f, rev).(*File)
}

func (f *File) Del() error {
	return f.conn.Del(f.Path, f.Rev)
}

// Update sets the value at this file's path to a new value.
func (f *File) Update(value interface{}) (file *File, err error) {
	evalue, err := f.Codec.Encode(value)
	if err != nil {
		return
	}

	rev, err := f.conn.Set(f.Path, f.Rev, evalue)
	if err != nil {
		return f, err
	}
	file = f.FastForward(rev)

	return
}

func (f *File) String() string {
	return fmt.Sprintf("%#v", f)
}
