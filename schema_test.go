// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"github.com/soundcloud/cotterpin"
	"testing"
	"time"
)

func cleanSchemaVersion(s *Store, t *testing.T) *Store {
	exists, _, err := s.GetSnapshot().Exists(schemaPath)
	if err != nil {
		t.Fatal("error calling del on '%s': %s", schemaPath, err.Error())
	}
	if exists {
		if err := s.GetSnapshot().Del(schemaPath); err != nil {
			t.Fatal("error calling get on '" + schemaPath + "': " + err.Error())
		}
	}
	s, err = s.FastForward()
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func TestSchemaMissing(t *testing.T) {
	s, err := DialUri(DefaultUri, DefaultRoot)

	if err != nil {
		panic(err)
	}

	s = cleanSchemaVersion(s, t)

	if _, err := s.VerifySchema(); !cotterpin.IsErrNoEnt(err) {
		if err == nil {
			t.Error("missing schema version did not error out")
		} else {
			t.Error("missing schema version returned incorrect error: " + err.Error())
		}
	}
}

func TestSetVersion(t *testing.T) {
	s, err := DialUri(DefaultUri, DefaultRoot)

	if err != nil {
		panic(err)
	}

	s = cleanSchemaVersion(s, t)

	if s, err = s.SetSchemaVersion(SchemaVersion); err != nil {
		t.Fatal("setting schema version failed")
	}

	if _, err := s.VerifySchema(); err != nil {
		t.Error("setting new version failed: " + err.Error())
	}
}

func TestVersionTooNew(t *testing.T) {
	s, err := DialUri(DefaultUri, DefaultRoot)
	coordinatorVersion := 0

	if err != nil {
		panic(err)
	}

	s = cleanSchemaVersion(s, t)

	if s, err = s.SetSchemaVersion(coordinatorVersion); err != nil {
		t.Fatal("setting schema version failed")
	}

	if _, err := s.VerifySchema(); err != ErrSchemaMism {
		if err == nil {
			t.Error("newer schema version did not error out")
		} else {
			t.Error("newer schema version returned incorrect error: " + err.Error())
		}
	}
}

func TestVersionTooOld(t *testing.T) {
	s, err := DialUri(DefaultUri, DefaultRoot)
	coordinatorVersion := 99

	if err != nil {
		panic(err)
	}

	s = cleanSchemaVersion(s, t)

	if s, err = s.SetSchemaVersion(coordinatorVersion); err != nil {
		t.Fatal("setting schema version failed")
	}

	if _, err := s.VerifySchema(); err != ErrSchemaMism {
		if err == nil {
			t.Error("older schema version did not error out")
		} else {
			t.Error("older schema version returned incorrect error: " + err.Error())
		}
	}
}

func TestSchemaWatcher(t *testing.T) {
	s, err := DialUri(DefaultUri, DefaultRoot)
	coordinatorVersion := 2

	if err != nil {
		panic(err)
	}

	s = cleanSchemaVersion(s, t)

	if s, err = s.SetSchemaVersion(coordinatorVersion); err != nil {
		t.Fatal(err)
	}

	schemaEvents := make(chan SchemaEvent, 1)
	errch := make(chan error, 1)

	go s.WatchSchema(schemaEvents, errch)

	go func() {
		if s, err = s.SetSchemaVersion(coordinatorVersion + 1); err != nil {
			t.Fatal(err)
		}
	}()

	select {
	case <-time.After(time.Second):
		t.Error("waiting for schema change timed out")
	case e := <-errch:
		t.Error(e)
	case <-schemaEvents:
		// Ok
	}
}
