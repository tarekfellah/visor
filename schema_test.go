// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
  "strconv"
  "time"
)


func cleanSchemaVersion(s Snapshot, t *testing.T) (Snapshot) {
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

func TestEnsureSchemaOnInitialRun(t *testing.T) {
	s, err := Dial(DefaultAddr, DefaultRoot)
  version := Schema{1}

	if err != nil {
		panic(err)
	}

  s = cleanSchemaVersion(s, t)

  if s, err = EnsureSchemaCompat(s, version); err != nil {
    t.Error("Error ensuring version on a clean tree: " + err.Error())
  }

  value, _, err := s.get(schemaPath)
  if err != nil {
    t.Fatal("error calling get on '" + schemaPath + "': " + err.Error())
  }

  intValue, err := strconv.Atoi(value)
  if err != nil {
    t.Fatal("error converting '" + value + "' to an integer")
  }

  if intValue != version.Version {
    t.Error("Wrong schema version was written to coordinator: expected '" + strconv.Itoa(version.Version) +"', got: '" + value + "'")
  }

	return
}

func TestEnsureSchemaNormal(t *testing.T) {
	s, err := Dial(DefaultAddr, DefaultRoot)
  version := Schema{1}

	if err != nil {
		panic(err)
	}

  s = cleanSchemaVersion(s, t)

  if s, err = EnsureSchemaCompat(s, version); err != nil {
    t.Error("Error ensuring version on a clean tree: " + err.Error())
  }

  value, _, err := s.get(schemaPath)
  if err != nil {
    t.Fatal("error calling get on '" + schemaPath + "': " + err.Error())
  }

  intValue, err := strconv.Atoi(value)
  if err != nil {
    t.Fatal("error converting '" + value + "' to an integer")
  }

  if intValue != version.Version {
    t.Error("Wrong schema version was written to coordinator: expected '" + strconv.Itoa(version.Version) +"', got: '" + value + "'")
  }

  if s, err = EnsureSchemaCompat(s, version); err != nil {
    t.Error("Error ensuring equal version: " + err.Error())
  }

  if s, err = EnsureSchemaCompat(s, version); err != nil {
    t.Error("Error ensuring equal version (2): " + err.Error())
  }

  version.Version += 1
  if s, err = EnsureSchemaCompat(s, version); err != nil {
    t.Error("Error ensuring version on a clean tree: " + err.Error())
  }

  value, _, err = s.get(schemaPath)
  if err != nil {
    t.Fatal("error calling get on '" + schemaPath + "': " + err.Error())
  }

  intValue, err = strconv.Atoi(value)
  if err != nil {
    t.Fatal("error converting '" + value + "' to an integer")
  }

  if intValue != version.Version {
    t.Error("Wrong schema version was written to coordinator: expected '" + strconv.Itoa(version.Version) +"', got: '" + value + "'")
  }

	return
}

func TestProvokeSchemaMismatch(t *testing.T) {
	s, err := Dial(DefaultAddr, DefaultRoot)
  version := Schema{2}

	if err != nil {
		panic(err)
	}

  s = cleanSchemaVersion(s, t)

  if s, err = EnsureSchemaCompat(s, version); err != nil {
    t.Error("Error ensuring version on a clean tree: " + err.Error())
  }

  version.Version -= 1
  if _, err = EnsureSchemaCompat(s, version); err != ErrSchemaMism {
    t.Error("Lower error schema version did not trigger error")
  }

	return
}

func TestSchemaWatcher(t *testing.T) {
	s, err := Dial(DefaultAddr, DefaultRoot)
  version := Schema{2}

	if err != nil {
		panic(err)
	}

  s = cleanSchemaVersion(s, t)

  if s, err = EnsureSchemaCompat(s, version); err != nil {
    t.Error("Error ensuring version on a clean tree: " + err.Error())
  }

  schemaEvents := make(chan Schema, 1)
  errch := make(chan error, 1)
  finished := make(chan bool)

  go func() {
    WatchSchema(s, schemaEvents, errch)
  }()

  go func() {
    select {
    case <-time.After(1 * time.Second):
      t.Error("Error watching for a schema update event")
    case e := <-errch:
      t.Error("SchemaWatch returned error: " + e.Error())
    case <- schemaEvents:
    }

    finished <-true
  }()

  version.Version += 1
  if _, err = EnsureSchemaCompat(s, version); err != nil {
    t.Fatal(err.Error())
  }

  <-finished
}

func TestSchemaWatcherMismatch(t *testing.T) {
	s, err := Dial(DefaultAddr, DefaultRoot)
  version := Schema{2}

	if err != nil {
		panic(err)
	}

  s = cleanSchemaVersion(s, t)

  if s, err = EnsureSchemaCompat(s, version); err != nil {
    t.Error("Error ensuring version on a clean tree: " + err.Error())
  }

  schemaEvents := make(chan Schema, 1)
  errch := make(chan error, 1)
  finished := make(chan bool)

  go func() {
    WatchSchema(s, schemaEvents, errch)
  }()

  go func() {
    select {
    case <-time.After(1 * time.Second):
    case e := <-errch:
      t.Error("SchemaWatch returned error: " + e.Error())
    case <- schemaEvents:
      t.Error("Version mismatch triggered event!")
    }

    finished <-true
  }()

  version.Version -= 1
  if _, err = EnsureSchemaCompat(s, version); err != ErrSchemaMism {
    t.Fatal(err.Error())
  }

  <-finished
}
