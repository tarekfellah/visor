package visor

import(
  "strconv"
)

const schemaPath = "/internal/schema"

type Schema struct {
	Version int
}

// WatchSchema watches for schema changes. Use it to react on schema updates
//
// If an event occurs, client can simply call EnsureSchemaCompat again to check
// whether they are still up to date
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
			continue
		}
		ch <- Schema{v}
	}
}

// EnsureSchemaCompat makes sure that the schema in the coordinator is equal to
// or smaller than the passed in version.
//
// If the current schema is smaller than version, the schema version
// is updated, if the current schema is greater than version, ErrSchemaMism
// is returned
//
// After a successful call, clients should watch out for schema changes
// using the WatchSchema function and not write anymore if the schema
// version is incremented.
func EnsureSchemaCompat(s Snapshot, version Schema) (Snapshot, error) {
  exists, _, err := s.exists(schemaPath)
  if err != nil {
    return s, err
  }

  strVersion := strconv.Itoa(version.Version)

  if !exists {
    s, err = s.set(schemaPath, strVersion)
    return s, err
  } else {
    value, _, err := s.get(schemaPath)
    if err != nil {
      return s, err
    }

    intValue, err := strconv.Atoi(value)
    if err != nil {
      return s, err
    }

    if intValue > version.Version {
      return s, ErrSchemaMism
    }

    if intValue < version.Version {
      s, err = s.set(schemaPath, strVersion)
      return s, err
    }
  }

  return s, nil
}
