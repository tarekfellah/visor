// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"fmt"
	"github.com/soundcloud/doozer"
	"strings"
)

// Conn is a wrapper around doozer.Conn,
// providing some additional and sometimes
// higher-level methods.
type conn struct {
	Addr string
	Root string
	conn *doozer.Conn
}

// Set calls (*doozer.Conn).Set with a prefixed path
func (c *conn) Set(path string, rev int64, value []byte) (newrev int64, err error) {
	path = c.prefixPath(path)
	if !pathRe.MatchString(path) {
		return rev, NewError(ErrBadPath, fmt.Sprintf("%s, got path: %s", ErrBadPath.Error(), path))
	}

	newrev, err = c.conn.Set(path, rev, value)
	if err != nil {
		if newrev == 0 { // err + newrev == 0: REV MISMATCH
			_, newrev, _ = c.Stat(path)
		}
		err = errors.New(fmt.Sprintf("error setting file '%s' to '%s': %s", path, string(value), err.Error()))
	}
	return
}

// Stat calls (*doozer.Conn).Stat with a prefixed path
func (c *conn) Stat(path string) (len int, pathrev int64, err error) {
	return c.conn.Stat(c.prefixPath(path), nil)
}

// Exists returns true or false depending on if the path exists
func (c *conn) Exists(path string) (exists bool, pathrev int64, err error) {
	return c.ExistsRev(path, nil)
}

// ExistsRev returns true or false depending on if the path exists
func (c *conn) ExistsRev(path string, rev *int64) (exists bool, pathrev int64, err error) {
	_, pathrev, err = c.conn.Stat(c.prefixPath(path), rev)
	if err != nil {
		return
	}

	// directories have negative revisions and should also be found
	if pathrev != 0 {
		exists = true
	}
	return
}

// Create is a wrapper around (*Conn).Set which returns an error if the file already exists.
func (c *conn) Create(path string, value []byte) (newrev int64, err error) {
	path = c.prefixPath(path)
	exists, newrev, err := c.Exists(path)
	if err != nil {
		return
	}
	if exists {
		return newrev, errors.New(fmt.Sprintf("path %s already exists", path))
	}
	return c.Set(path, newrev, value)
}

// Rev is a wrapper around (*doozer.Conn).Rev.
func (c *conn) Rev() (int64, error) {
	return c.conn.Rev()
}

// Get is a wrapper around (*doozer.Conn).Get with a prefixed path.
func (c *conn) Get(path string, rev *int64) (value []byte, filerev int64, err error) {
	value, filerev, err = c.conn.Get(c.prefixPath(path), rev)

	// If the file revision is 0 and there is no error, set the error appropriately.
	// We don't want to overwrite the error in case it is network-related.
	if filerev == 0 && err == nil {
		if rev == nil {
			err = NewError(ErrNoEnt, fmt.Sprintf("path \"%s\" not found at latest revision", path))
		} else {
			err = NewError(ErrNoEnt, fmt.Sprintf("path \"%s\" not found at %d", path, *rev))
		}
	}
	return
}

// Getdir is a wrapper around (*doozer.Conn).Getdir with a prefixed path.
func (c *conn) Getdir(path string, rev int64) (names []string, err error) {
	type reply struct {
		name  string
		err   error
		index int
	}

	if rev < 0 {
		return nil, fmt.Errorf("rev must be >= 0")
	}
	path = c.prefixPath(path)
	size, _, err := c.conn.Stat(path, &rev)

	if err == doozer.ErrNoEnt || (err != nil && err.Error() == "NOENT") {
		return nil, NewError(ErrNoEnt, fmt.Sprintf(`dir "%s" not found at %d`, path, rev))
	}
	names = make([]string, size)
	replies := make(chan reply, size)

	for i := 0; i < size; i++ {
		go func(index int) {
			names, err := c.conn.Getdir(path, rev, index, 1)
			if err != nil {
				replies <- reply{err: err}
				return
			}
			replies <- reply{name: names[0], index: index}
		}(i)
	}
	for i := 0; i < size; i++ {
		res := <-replies
		if res.err == nil {
			names[res.index] = res.name
		} else {
			return nil, res.err
		}
	}
	return
}

// Wait is a wrapper around (*doozer.Conn).Wait
func (c *conn) Wait(path string, rev int64) (event doozer.Event, err error) {
	path = c.prefixPath(path)
	event, err = c.conn.Wait(path, rev)
	event.Path = strings.Replace(event.Path, c.Root, "", 1)
	return
}

// Wait is a wrapper around (*doozer.Conn).Close
func (c *conn) Close() {
	c.conn.Close()
}

// Del is a wrapper around (*doozer.Conn).Del which also supports
// deleting directories.
// TODO: Concurrent implementation
// TODO: Better error messages on NOENT
func (c *conn) Del(path string, rev int64) (err error) {
	path = c.prefixPath(path)

	err = doozer.Walk(c.conn, rev, path, func(path string, f *doozer.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if !f.IsDir {
			e = c.conn.Del(path, rev)
			if e != nil {
				return e
			}
		}

		return nil
	})

	return
}

// GetMulti returns multiple key/value pairs organized in a map.
func (c *conn) GetMulti(path string, keys []string, rev int64) (values map[string][]byte, err error) {
	if keys == nil {
		keys, err = c.Getdir(path, rev)
	}
	if err != nil {
		return
	}
	values = make(map[string][]byte)

	for i := range keys {
		val, _, e := c.Get(path+"/"+keys[i], &rev)
		if e != nil {
			return nil, e
		}
		values[keys[i]] = val
	}
	return
}

// SetMulti stores mutliple key/value pairs under the given path.
func (c *conn) SetMulti(path string, kvs map[string][]byte, rev int64) (newrev int64, err error) {
	for k, v := range kvs {
		newrev, err = c.Set(path+"/"+k, rev, v)
		if err != nil {
			break
		}
	}
	return
}

func (c *conn) prefixPath(p string) (path string) {
	prefix := c.Root
	path = p

	if p == "/" {
		return prefix
	}

	if !strings.HasSuffix(prefix, "/") && !strings.HasPrefix(p, "/") {
		prefix = prefix + "/"
	}

	if !strings.HasPrefix(p, c.Root) {
		path = prefix + p
	}

	return path
}
