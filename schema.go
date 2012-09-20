package visor

import "strconv"

const schemaPath = "/internal/schema"

type Schema struct {
	Version int
}

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
			panic(err)
		}
		ch <- Schema{v}
	}
}
