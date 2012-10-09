// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
	"time"
)

func cleanSchemaVersion(s Snapshot, t *testing.T) Snapshot {
	var exists bool
	var err error

	if exists, _, err = s.exists(schemaPath); err != nil {
		t.Fatal("error calling del on '%s': %s", schemaPath, err.Error())
	}

	if exists {
		if err := s.del(schemaPath); err != nil {
			t.Fatal("error calling get on '" + schemaPath + "': " + err.Error())
		}
		s = s.FastForward(-1)
	}

	return s
}

func TestSchemaMissing(t *testing.T) {
	s, err := Dial(DefaultAddr, DefaultRoot)

	if err != nil {
		panic(err)
	}

	s = cleanSchemaVersion(s, t)

	if err := VerifySchema(s); err != ErrSchemaMism {
		if err == nil {
			t.Error("missing schema version did not error out")
		} else {
			t.Error("missing schema version returned incorrect error: " + err.Error())
		}
	}
}

func TestSetVersion(t *testing.T) {
	s, err := Dial(DefaultAddr, DefaultRoot)

	if err != nil {
		panic(err)
	}

	s = cleanSchemaVersion(s, t)

	if s, err = SetSchemaVersion(s, SchemaVersion); err != nil {
		t.Fatal("setting schema version failed")
	}

	if err := VerifySchema(s); err != nil {
		t.Error("setting new version failed: " + err.Error())
	}
}

func TestVersionTooNew(t *testing.T) {
	s, err := Dial(DefaultAddr, DefaultRoot)
	coordinatorVersion := 0

	if err != nil {
		panic(err)
	}

	s = cleanSchemaVersion(s, t)

	if s, err = SetSchemaVersion(s, coordinatorVersion); err != nil {
		t.Fatal("setting schema version failed")
	}

	if err := VerifySchema(s); err != ErrSchemaMism {
		if err == nil {
			t.Error("newer schema version did not error out")
		} else {
			t.Error("newer schema version returned incorrect error: " + err.Error())
		}
	}
}

func TestVersionTooOld(t *testing.T) {
	s, err := Dial(DefaultAddr, DefaultRoot)
	coordinatorVersion := 99

	if err != nil {
		panic(err)
	}

	s = cleanSchemaVersion(s, t)

	if s, err = SetSchemaVersion(s, coordinatorVersion); err != nil {
		t.Fatal("setting schema version failed")
	}

	if err := VerifySchema(s); err != ErrSchemaMism {
		if err == nil {
			t.Error("older schema version did not error out")
		} else {
			t.Error("older schema version returned incorrect error: " + err.Error())
		}
	}

}

func TestSchemaWatcher(t *testing.T) {
	s, err := Dial(DefaultAddr, DefaultRoot)
	coordinatorVersion := 2

	if err != nil {
		panic(err)
	}

	s = cleanSchemaVersion(s, t)

	if s, err = SetSchemaVersion(s, coordinatorVersion); err != nil {
		t.Fatal("setting schema version failed")
	}

	schemaEvents := make(chan SchemaEvent, 1)
	errch := make(chan error, 1)
	finished := make(chan bool)

	go func() {
		WatchSchema(s, schemaEvents, errch)
	}()

	go func() {
		select {
		case <-time.After(1 * time.Second):
			t.Error("error watching for a schema update event")
		case e := <-errch:
			t.Error("schemaWatch returned error: " + e.Error())
		case <-schemaEvents:
		}

		finished <- true
	}()

	coordinatorVersion += 1
	if s, err = SetSchemaVersion(s, coordinatorVersion); err != nil {
		t.Fatal("setting schema version failed: " + err.Error())
	}

	<-finished
}
