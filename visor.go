// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"fmt"
	cp "github.com/soundcloud/cotterpin"
	"net"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const SchemaVersion = 3

const (
	DefaultUri     = "doozer:?ca=localhost:8046"
	DefaultRoot    = "/visor"
	startPort      = 8000
	nextPortPath   = "/next-port"
	loggerDir      = "/loggers"
	proxyDir       = "/proxies"
	pmDir          = "/pms"
	UTCFormat      = "2006-01-02 15:04:05 -0700 MST"
	registeredPath = "registered"
)

// Set *automatically* at link time (see Makefile)
var Version string

type Store struct {
	snapshot cp.Snapshot
}

func DialUri(uri, root string) (*Store, error) {
	sp, err := cp.DialUri(uri, root)
	if err != nil {
		return nil, err
	}
	return &Store{sp}, nil
}

func (s *Store) GetSnapshot() cp.Snapshot {
	return s.snapshot
}

func (s *Store) FastForward() (*Store, error) {
	sp, err := s.GetSnapshot().FastForward()
	if err != nil {
		return nil, err
	}
	return &Store{sp}, nil
}

func (s *Store) Init() (*Store, error) {
	sp, err := s.GetSnapshot().FastForward()
	if err != nil {
		return nil, err
	}

	exists, _, err := sp.Exists(nextPortPath)
	if err != nil {
		return nil, err
	}

	if !exists {
		sp, err = sp.Set(nextPortPath, strconv.Itoa(startPort))
		if err != nil {
			return nil, err
		}
	}

	v, err := cp.VerifySchema(SchemaVersion, sp)
	if cp.IsErrNoEnt(err) {
		sp, err = cp.SetSchemaVersion(SchemaVersion, sp)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		if cp.IsErrSchemaMism(err) {
			err = fmt.Errorf("%s (%d != %d)", err, SchemaVersion, v)
		}
		return nil, err
	}

	s.snapshot = sp

	return s, nil
}

func (s *Store) Scale(app, rev, proc, env string, factor int) (tickets []*Instance, current int, err error) {
	if err := validateInput(app); err != nil {
		return nil, -1, errorf(err, "given app not valid: %s (%s)", app, err)
	}
	if err := validateInput(rev); err != nil {
		return nil, -1, errorf(err, "given rev not valid: %s (%s)", rev, err)
	}
	if err := validateInput(proc); err != nil {
		return nil, -1, errorf(err, "given proc not valid: %s (%s)", proc, err)
	}
	if err := validateInput(env); err != nil {
		return nil, -1, errorf(err, "given env not valid: %s (%s)", env, err)
	}
	if factor < 0 {
		return nil, -1, errors.New("scaling factor needs to be a positive integer")
	}

	sp, err := s.GetSnapshot().FastForward()
	if err != nil {
		return
	}

	exists, _, err := sp.Exists(path.Join(appsPath, app, revsPath, rev))
	if err != nil {
		return
	}
	if !exists {
		return nil, -1, errorf(ErrNotFound, "rev '%s' not found for app '%s'", rev, app)
	}
	exists, _, err = sp.Exists(path.Join(appsPath, app, procsPath, proc))
	if err != nil {
		return
	}
	if !exists {
		return nil, -1, errorf(ErrNotFound, "proc '%s' not found", proc)
	}

	s.snapshot = sp

	ids, err := getInstanceIds(app, rev, proc, sp)
	if err != nil {
		return nil, -1, err
	}

	is := []*Instance{}
	for _, id := range ids {
		// TODO parallelize the instance retrieval and order after (xla)
		i, err := getInstance(id, s)
		if err != nil {
			return nil, -1, err
		}

		if i.Env != env {
			continue
		}

		is = append(is, i)
	}

	current = len(is)

	if factor > current {
		// Scale up
		ntickets := factor - current

		for i := 0; i < ntickets; i++ {
			ticket, err := s.RegisterInstance(app, rev, proc, env)
			if err != nil {
				return nil, -1, err
			}
			tickets = append(tickets, ticket)

			s.snapshot = s.GetSnapshot().Join(ticket)
		}
	} else if factor < current {
		// Scale down
		stops := current - factor
		for i := 0; i < stops; i++ {
			ins := is[i]

			err = ins.Stop()
			if err != nil {
				if IsErrInvalidState(err) {
					err = errorf(ErrInvalidState, "instance '%d' isn't running", ins.Id)
				}
				return nil, -1, err
			}

			tickets = append(tickets, ins)
		}
	}
	return
}

// GetScale returns the scale of an app:proc@rev tuple. If the scale isn't found, 0 is returned.
func (s *Store) GetScale(app string, revid string, proc string) (scale int, rev int64, err error) {
	sp, err := s.GetSnapshot().FastForward()
	if err != nil {
		return
	}

	path := procInstancesPath(app, revid, proc)
	count, rev, err := sp.Stat(path, &s.snapshot.Rev)

	// File doesn't exist, assume scale = 0
	if cp.IsErrNoEnt(err) {
		return 0, rev, nil
	}
	if err != nil {
		return -1, rev, err
	}

	return count, rev, nil
}

// GetLoggers gets the list of bazooka-log services endpoints.
func (s *Store) GetLoggers() ([]string, error) {
	sp, err := s.GetSnapshot().FastForward()
	if err != nil {
		return nil, err
	}
	names, err := sp.Getdir(loggerDir)
	if err != nil {
		return nil, err
	}
	for i, name := range names {
		names[i] = strings.Replace(name, "-", ":", 1)
	}
	return names, nil
}

// GetProxies gets the list of bazooka-proxy service IPs
func (s *Store) GetProxies() ([]string, error) {
	sp, err := s.GetSnapshot().FastForward()
	if err != nil {
		return nil, err
	}
	return sp.Getdir(proxyDir)
}

// GetPms gets the list of bazooka-pm service IPs
func (s *Store) GetPms() ([]string, error) {
	sp, err := s.GetSnapshot().FastForward()
	if err != nil {
		return nil, err
	}
	return sp.Getdir(pmDir)
}

func (s *Store) GetAppNames() ([]string, error) {
	sp, err := s.GetSnapshot().FastForward()
	if err != nil {
		return nil, err
	}
	return sp.Getdir("apps")
}

func (s *Store) RegisterLogger(addr, version string) (*Store, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	sp, err := s.GetSnapshot().Set(path.Join(loggerDir, host+"-"+port), timestamp()+" "+version)
	if err != nil {
		return nil, err
	}
	s.snapshot = sp
	return s, nil
}

func (s *Store) UnregisterLogger(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	return s.GetSnapshot().Del(path.Join(loggerDir, host+"-"+port))
}

func (s *Store) RegisterPm(host, version string) (*Store, error) {
	sp, err := s.GetSnapshot().Set(path.Join(pmDir, host), timestamp()+" "+version)
	if err != nil {
		return nil, err
	}
	s.snapshot = sp
	return s, nil
}

func (s *Store) UnregisterPm(host string) error {
	return s.GetSnapshot().Del(path.Join(pmDir, host))
}

func (s *Store) RegisterProxy(host string) (*Store, error) {
	sp, err := s.GetSnapshot().Set(path.Join(proxyDir, host), timestamp())
	if err != nil {
		return nil, err
	}
	s.snapshot = sp
	return s, nil
}

func (s *Store) SetSchemaVersion(version int) error {
	sp, err := s.GetSnapshot().FastForward()
	if err != nil {
		return err
	}
	_, err = cp.SetSchemaVersion(version, sp)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) VerifySchema() (int, error) {
	sp, err := s.GetSnapshot().FastForward()
	if err != nil {
		return -1, err
	}
	v, err := cp.VerifySchema(SchemaVersion, sp)
	if err != nil {
		if cp.IsErrSchemaMism(err) {
			err = fmt.Errorf("%s (%d != %d)", err, SchemaVersion, v)
		}
		return v, err
	}
	return v, nil
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

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func parseTime(val string) (time.Time, error) {
	return time.Parse(time.RFC3339, val)
}

func validateInput(s string) error {
	if len(s) < 1 {
		return errorf(ErrInvalidArgument, "input can't be zero length")
	}
	validInput := regexp.MustCompile(`^[[:alnum:]\-\.]+$`)
	if !validInput.MatchString(s) {
		return errorf(ErrInvalidArgument, "input only allows alphanumeric characters and -")
	}
	return nil
}
