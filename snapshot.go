package visor

import (
	"github.com/soundcloud/doozer"
	"net"
)

// Snapshot represents a specific point in time
// within the coordinator state. It is used by
// all time-aware interfaces to the coordinator.
type Snapshot struct {
	Rev  int64
	conn *Conn
}

// Snapshotable is implemented by any type which
// is time-aware, and can be moved forward in time
// by calling createSnapshot with a new revision.
type Snapshotable interface {
	createSnapshot(rev int64) Snapshotable
}

// Dial calls doozer.Dial and returns a Snapshot of the coordinator
// at the latest revision.
func Dial(addr string, root string) (s Snapshot, err error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	dconn, err := doozer.Dial(addr)
	if err != nil {
		return
	}

	rev, err := dconn.Rev()
	if err != nil {
		return
	}

	s = Snapshot{rev, &Conn{tcpaddr, root, dconn}}
	return
}

func (s Snapshot) Conn() *Conn {
	return s.conn
}

func (s Snapshot) createSnapshot(rev int64) Snapshotable {
	return Snapshot{rev, s.conn}
}

func (s Snapshot) FastForward(rev int64) (ns Snapshot) {
	return s.fastForward(s, rev).(Snapshot)
}

// fastForward either calls *createSnapshot* on *obj* or returns *obj* if it
// can't advance the object in time. Note that fastForward can never fail.
func (s *Snapshot) fastForward(obj Snapshotable, rev int64) Snapshotable {
	var err error

	if rev == -1 {
		rev, err = s.conn.Rev()
		if err != nil {
			return obj
		}
	} else if rev < s.Rev {
		return obj
	}
	return obj.createSnapshot(rev)
}

// Get returns the value for the given path
func Get(s Snapshot, path string, codec Codec) (file *File, err error) {
	evalue, _, err := s.conn.Get(path, &s.Rev)
	if err != nil {
		return
	}

	value, err := codec.Decode(evalue)
	if err != nil {
		return
	}

	file = &File{Path: path, Value: value, Codec: codec, Snapshot: s}

	return
}
