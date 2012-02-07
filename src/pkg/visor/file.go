package visor

import (
	"fmt"
	"reflect"
)

// File represents a coordinator file
// at a specific point in time.
type File struct {
	Path  string
	Rev   int64
	Value reflect.Value
}

// NewFile returns a new file object.
func NewFile(path string, rev int64, value reflect.Value) *File {
	f := &File{path, rev, value}
	return f
}

// Update sets the value at this file's path to a new value.
func (f *File) Update(client *Client, value interface{}) (file *File, err error) {
	client, err = client.FastForward(f.Rev)
	if err != nil {
		return
	}
	file, err = client.Set(f.Path, value)
	return
}

func (f *File) String() string {
	return fmt.Sprintf("%#v", f)
}
