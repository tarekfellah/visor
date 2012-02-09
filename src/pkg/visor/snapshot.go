package visor

type Snapshot struct {
	rev  int64
	conn *Conn
}

type Snapshotable interface {
	CreateSnapshot(rev int64) Snapshotable
}

func (s *Snapshot) FastForward(obj Snapshotable, rev int64) (newobj Snapshotable) {
	var err error

	if rev == -1 {
		rev, err = s.conn.Rev()
		if err != nil {
			return obj
		}
	} else if rev < s.rev {
		return obj
	}
	newobj = obj.CreateSnapshot(rev)

	return
}
