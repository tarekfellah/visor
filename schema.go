package visor

import (
	"strconv"
)

const schemaPath = "/internal/schema"

type SchemaEvent struct {
	Version int
}

// WatchSchema notifies the specified ch channel on schema change,
// and errch on error. If an error occures, WatchSchema exits.
func WatchSchema(s Snapshot, ch chan SchemaEvent, errch chan error) {
	rev := s.Rev
	for {
		ev, err := s.conn.Wait(schemaPath, rev+1)
		if err != nil {
			errch <- err
			return
		}
		rev = ev.Rev

		v, err := strconv.Atoi(string(ev.Body))
		if err != nil {
			errch <- err
			return
		}
		ch <- SchemaEvent{v}
	}
}

// VerifySchema checks that visor's schema version is compatible with
// the coordinator's. If this is not the case, ErrSchemaMism is returned.
//
// See WatchSchema to get notified on schema change.
func VerifySchema(s Snapshot) error {
	return verifySchemaVersion(s, SchemaVersion)
}

func SetSchemaVersion(s Snapshot, version int) (newSnapshot Snapshot, err error) {
	strVersion := strconv.Itoa(version)

	newSnapshot, err = s.set(schemaPath, strVersion)

	return
}

func verifySchemaVersion(s Snapshot, version int) error {
	exists, _, err := s.exists(schemaPath)
	if err != nil {
		return err
	}

	if !exists {
		return ErrSchemaMism
	}

	value, _, err := s.get(schemaPath)
	if err != nil {
		return err
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	if intValue != version {
		return ErrSchemaMism
	}

	return nil
}
