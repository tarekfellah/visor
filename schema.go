package visor

import (
	"strconv"
)

const schemaPath = "/internal/schema"

type Schema struct {
	Version int
}

// WatchSchema watches for schema changes. Use it to react on schema updates
//
// If an event occurs, call VerifySchemaVersion again to check whether you
// still support the current coordinator schema version
func WatchSchema(s Snapshot, ch chan Schema, errch chan error) {
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
		ch <- Schema{v}
	}
}

// VerifySchemaVersion makes sure that the given schema version is compatible
// with the current coordinator schema; if this is not the case ErrSchemaMism is
// returned
//
// Once the schema version is verified, make sure to use WatchSchema for schema
// changes and act accordingly
func VerifySchemaVersion(s Snapshot, version Schema) error {
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

	if intValue != version.Version {
		return ErrSchemaMism
	}

	return nil
}

func SetSchemaVersion(s Snapshot, version Schema) (newSnapshot Snapshot, err error) {
	strVersion := strconv.Itoa(version.Version)

	newSnapshot, err = s.set(schemaPath, strVersion)

	return
}
