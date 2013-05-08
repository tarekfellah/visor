// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"fmt"
	cp "github.com/soundcloud/cotterpin"
	"path"
	"strconv"
	"time"
)

const SchemaVersion = 2

const (
	DefaultUri   string = "doozer:?ca=localhost:8046"
	DefaultRoot  string = "/visor"
	startPort    int    = 8000
	nextPortPath string = "/next-port"
	proxyDir            = "/proxies"
	pmDir               = "/pms"
)

// Set *automatically* at link time (see Makefile)
var Version string

type Store struct {
	snapshot cp.Snapshot
}

func DialUri(uri, root string) (*Store, error) {
	snapshot, err := cp.DialUri(uri, root)
	if err != nil {
		return nil, err
	}
	return &Store{snapshot}, nil
}

func (s *Store) GetSnapshot() cp.Snapshot {
	return s.snapshot
}

// Join advances the Store in time. It returns a new
// instance of Store at the rev of the supplied
// cp.Snapshotable.
func (s *Store) Join(sp cp.Snapshotable) *Store {
	tmp := *s
	tmp.snapshot = sp.GetSnapshot()
	return &tmp
}

func (s *Store) FastForward() (*Store, error) {
	sp, err := s.snapshot.FastForward()
	if err != nil {
		return nil, err
	}
	return &Store{sp}, nil
}

func (s *Store) Init() (*Store, error) {
	s, err := s.FastForward()
	if err != nil {
		return nil, err
	}

	v, err := cp.VerifySchema(SchemaVersion, s.GetSnapshot())
	if err != nil {
		if err == cp.ErrSchemaMism {
			err = fmt.Errorf("%s (%d != %d)", err, SchemaVersion, v)
		}
		return nil, err
	}

	exists, _, err := s.GetSnapshot().Exists(nextPortPath)
	if err != nil {
		return nil, err
	}

	if !exists {
		sp, err := s.GetSnapshot().Set(nextPortPath, strconv.Itoa(startPort))
		if err != nil {
			return nil, err
		}
		s = s.Join(sp)
	}

	sp, err := cp.SetSchemaVersion(SchemaVersion, s.GetSnapshot())
	if err != nil {
		return nil, err
	}
	s = s.Join(sp)

	return s, nil
}

func (s *Store) Scale(app string, revision string, processName string, factor int) (tickets []*Instance, current int, err error) {
	if factor < 0 {
		return nil, -1, errors.New("scaling factor needs to be a positive integer")
	}

	exists, _, err := s.GetSnapshot().Exists(path.Join(appsPath, app, revsPath, revision))
	if !exists || err != nil {
		return nil, -1, fmt.Errorf("%s@%s not found", app, revision)
	}
	exists, _, err = s.GetSnapshot().Exists(path.Join(appsPath, app, procsPath, processName))
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

			ticket, err = s.RegisterInstance(app, revision, processName)
			if err != nil {
				return
			}
			tickets = append(tickets, ticket)

			s = s.Join(ticket)
		}
	} else if factor < current {
		// Scale down
		stops := current - factor
		for i := 0; i < stops; i++ {
			s, err = s.StopInstance(list[i])
			if err != nil {
				return
			}
		}
	}
	return
}

// GetScale returns the scale of an app:pty@rev tuple. If the scale isn't found, 0 is returned.
func (s *Store) GetScale(app string, revid string, pty string) (scale int, rev int64, err error) {
	path := ptyInstancesPath(app, revid, pty)
	count, rev, err := s.snapshot.Stat(path, &s.snapshot.Rev)

	// File doesn't exist, assume scale = 0
	if cp.IsErrNoEnt(err) {
		return 0, rev, nil
	}

	if err != nil {
		return -1, rev, err
	}

	return count, rev, nil
}

// GetProxies gets the list of bazooka-proxy service IPs
func (s *Store) GetProxies() ([]string, error) {
	return s.GetSnapshot().Getdir(proxyDir)
}

// GetPms gets the list of bazooka-pm service IPs
func (s *Store) GetPms() ([]string, error) {
	return s.GetSnapshot().Getdir(pmDir)
}

func (s *Store) GetApps() ([]string, error) {
	return s.GetSnapshot().Getdir("apps")
}

func (s *Store) RegisterPm(host, version string) (*Store, error) {
	sp, err := s.GetSnapshot().Set(path.Join(pmDir, host), timestamp()+" "+version)
	if err != nil {
		return nil, err
	}
	return storeFromSnapshotable(sp), nil
}

func (s *Store) UnregisterPm(host string) error {
	return s.GetSnapshot().Del(path.Join(pmDir, host))
}

func (s *Store) RegisterProxy(host string) (*Store, error) {
	sp, err := s.GetSnapshot().Set(path.Join(proxyDir, host), timestamp())
	if err != nil {
		return nil, err
	}
	return storeFromSnapshotable(sp), nil
}

func (s *Store) UnregisterProxy(host string) error {
	return s.GetSnapshot().Del(path.Join(proxyDir, host))
}

func (s *Store) reset() error {
	return s.GetSnapshot().Reset()
}

func storeFromSnapshotable(sp cp.Snapshotable) *Store {
	return &Store{sp.GetSnapshot()}
}

func timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
