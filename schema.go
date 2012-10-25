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
func VerifySchema(s Snapshot) (int, error) {
	return verifySchemaVersion(s, SchemaVersion)
}

func SetSchemaVersion(s Snapshot, version int) (newSnapshot Snapshot, err error) {
	strVersion := strconv.Itoa(version)

	newSnapshot, err = s.set(schemaPath, strVersion)

	return
}

func verifySchemaVersion(s Snapshot, version int) (int, error) {
	value, _, err := s.get(schemaPath)
	if err != nil {
		return -1, err
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return intValue, err
	}

	if intValue != version {
		return intValue, ErrSchemaMism
	}

	return -1, nil
}
