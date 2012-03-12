package visor

import (
	"fmt"
)

// File represents a coordinator file
// at a specific point in time.
type File struct {
	Snapshot
	Path  string
	Value interface{}
	Codec Codec
}

func CreateFile(snapshot Snapshot, path string, value interface{}, codec Codec) (*File, error) {
	file := &File{Path: path, Value: value, Codec: codec, Snapshot: snapshot}
	return file.Create()
}

func (f *File) createSnapshot(rev int64) (file Snapshotable) {
	file = &File{Path: f.Path, Value: f.Value, Codec: f.Codec, Snapshot: Snapshot{rev, f.conn}}
	return
}

// FastForward advances the file in time. It returns
// a new instance of File with the supplied revision.
func (f *File) FastForward(rev int64) *File {
	if rev == -1 {
		var err error
		_, rev, err = f.conn.Stat(f.Path, nil)
		if err != nil {
			return f
		}
	}
	return f.Snapshot.fastForward(f, rev).(*File)
}

// Del deletes a file
func (f *File) Del() error {
	return f.conn.Del(f.Path, f.Rev)
}

// Create creates a file from its Value attribute
func (f *File) Create() (*File, error) {
	return f.Update(f.Value)
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
	file.Value = value

	return
}

func (f *File) String() string {
	return fmt.Sprintf("%#v", f)
}
