// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"time"
)

const SchemaVersion = 2

const (
	DefaultUri   string = "doozer:?ca=localhost:8046"
	DefaultAddr  string = "localhost:8046"
	DefaultRoot  string = "/visor"
	startPort    int    = 8000
	nextPortPath string = "/next-port"
	uidPath      string = "/uid"
	proxyDir            = "/proxies"
	pmDir               = "/pms"
)

// Set *automatically* at link time (see Makefile)
var Version string

func Init(s Snapshot) (rev int64, err error) {
	exists, _, err := s.conn.Exists(nextPortPath)
	if err != nil {
		return
	}

	if !exists {
		s, err = s.set(nextPortPath, strconv.Itoa(startPort))
		if err != nil {
			return
		}
	}

	s, err = SetSchemaVersion(s, SchemaVersion)
	if err != nil {
		return
	}

	return s.Rev, nil
}

func ClaimNextPort(s Snapshot) (port int, err error) {
	for {
		f, err := getLatest(s, nextPortPath, new(intCodec))
		if err == nil {
			port = f.Value.(int)

			f, err = f.Set(port + 1)
			if err == nil {
				break
			} else {
				s = f.Snapshot
				time.Sleep(time.Second / 10)
			}
		} else {
			return -1, err
		}
	}

	return
}

func Scale(app string, revision string, processName string, factor int, s Snapshot) (tickets []*Instance, current int, err error) {
	if factor < 0 {
		return nil, -1, errors.New("scaling factor needs to be a positive integer")
	}

	exists, _, err := s.conn.Exists(path.Join(appsPath, app, revsPath, revision))
	if !exists || err != nil {
		return nil, -1, fmt.Errorf("%s@%s not found", app, revision)
	}
	exists, _, err = s.conn.Exists(path.Join(appsPath, app, procsPath, processName))
	if !exists || err != nil {
		return nil, -1, fmt.Errorf("proc '%s' doesn't exist", processName)
	}

	list, err := getInstanceIds(s, app, revision, processName)
	if err != nil {
		return nil, -1, err
	}
	current = len(list)

	if factor > current {
		// Scale up
		ntickets := factor - current

		for i := 0; i < ntickets; i++ {
			var ticket *Instance

			ticket, err = RegisterInstance(app, revision, processName, s)
			if err != nil {
				return
			}
			tickets = append(tickets, ticket)

			s = s.FastForward(ticket.Dir.Snapshot.Rev)
		}
	} else if factor < current {
		// Scale down
		stops := current - factor
		for i := 0; i < stops; i++ {
			s, err = StopInstance(list[i], s)
			if err != nil {
				panic(err)
			}
		}
	}
	return
}

func Lock(key, value string, s Snapshot) (Snapshot, error) {
	s1, e := s.set(path.Join("locks", key), value)
	if err, ok := e.(*Error); ok {
		if err.Err == ErrRevMismatch {
			return s, ErrLocked
		}
	}
	return s1, nil
}

func Unlock(key string, s Snapshot) error {
	return s.del(path.Join("locks", key))
}

func timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
