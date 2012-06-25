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
type Conn struct {
	Addr string
	Root string
	conn *doozer.Conn
}

// Set calls (*doozer.Conn).Set with a prefixed path
func (c *Conn) Set(path string, rev int64, value []byte) (newrev int64, err error) {
	path = c.prefixPath(path)
	newrev, err = c.conn.Set(path, rev, value)
	if err != nil {
		if newrev == 0 {
			_, newrev, _ = c.Stat(path)
		}
		err = errors.New(fmt.Sprintf("error setting file '%s' to '%s': %s", path, string(value), err.Error()))
	}
	return
}

// Stat calls (*doozer.Conn).Stat with a prefixed path
func (c *Conn) Stat(path string) (len int, pathrev int64, err error) {
	return c.conn.Stat(c.prefixPath(path), nil)
}

// Exists returns true or false depending on if the path exists
func (c *Conn) Exists(path string) (exists bool, pathrev int64, err error) {
	_, pathrev, err = c.conn.Stat(c.prefixPath(path), nil)
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
func (c *Conn) Create(path string, value []byte) (newrev int64, err error) {
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
func (c *Conn) Rev() (int64, error) {
	return c.conn.Rev()
}

// Get is a wrapper around (*doozer.Conn).Get with a prefixed path.
func (c *Conn) Get(path string, rev *int64) (value []byte, filerev int64, err error) {
	value, filerev, err = c.conn.Get(c.prefixPath(path), rev)
	if filerev == 0 {
		if rev == nil {
			err = fmt.Errorf("path \"%s\" not found at latest revision", path)
		} else {
			err = fmt.Errorf("path \"%s\" not found at %d", path, *rev)
		}
	}
	return
}

// Getdir is a wrapper around (*doozer.Conn).Getdir with a prefixed path.
func (c *Conn) Getdir(path string, rev int64) (keys []string, err error) {
	return c.conn.Getdir(c.prefixPath(path), rev, 0, -1)
}

// Wait is a wrapper around (*doozer.Conn).Wait
func (c *Conn) Wait(path string, rev int64) (event doozer.Event, err error) {
	path = c.prefixPath(path)
	event, err = c.conn.Wait(path, rev)
	event.Path = strings.Replace(event.Path, c.Root, "", 1)
	return
}

// Wait is a wrapper around (*doozer.Conn).Close
func (c *Conn) Close() {
	c.conn.Close()
}

// Del is a wrapper around (*doozer.Conn).Del which also supports
// deleting directories.
func (c *Conn) Del(path string, rev int64) (err error) {
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
func (c *Conn) GetMulti(path string, keys []string, rev int64) (values map[string][]byte, err error) {
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
func (c *Conn) SetMulti(path string, kvs map[string][]byte, rev int64) (newrev int64, err error) {
	for k, v := range kvs {
		newrev, err = c.Set(path+"/"+k, rev, v)
		if err != nil {
			break
		}
	}
	return
}

func (c *Conn) prefixPath(p string) (path string) {
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
