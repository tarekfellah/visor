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
func (s *Store) WatchSchema(ch chan SchemaEvent, errch chan error) {
	rev := s.snapshot.Rev
	for {
		ev, err := s.snapshot.Wait(schemaPath, rev+1)
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
func (s *Store) VerifySchema() (int, error) {
	value, _, err := s.GetSnapshot().Get(schemaPath)
	if err != nil {
		return -1, err
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return intValue, err
	}

	if intValue != SchemaVersion {
		return intValue, ErrSchemaMism
	}

	return -1, nil
}

func (s *Store) SetSchemaVersion(version int) (*Store, error) {
	strVersion := strconv.Itoa(version)

	sp, err := s.GetSnapshot().Set(schemaPath, strVersion)
	if err != nil {
		return nil, err
	}

	return storeFromSnapshotable(sp), nil
}
