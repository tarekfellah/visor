package visor

type dir struct {
	Snapshot
	Name string
}

func (p *dir) get(key string) (string, int64, error) {
	return p.Snapshot.get(p.prefix(key))
}

func (p *dir) set(key string, val string) (int64, error) {
	s, err := p.Snapshot.set(p.prefix(key), val)
	return s.Rev, err
}

func (p *dir) setBytes(key string, val []byte) (int64, error) {
	s, err := p.Snapshot.setBytes(p.prefix(key), val)
	return s.Rev, err
}

func (p *dir) getBytes(key string) ([]byte, int64, error) {
	return p.Snapshot.getBytes(p.prefix(key))
}

func (p *dir) del(key string) error {
	return p.Snapshot.del(p.prefix(key))
}

func (p *dir) prefix(path string, paths ...string) (result string) {
	if path == "/" {
		result = p.Name
	} else {
		result = p.Name + "/" + path
	}
	for _, p := range paths {
		result += "/" + p
	}
	return
}

func (p *dir) fastForward(rev int64) *dir {
	return &dir{p.Snapshot.FastForward(rev), p.Name}
}

func (p *dir) String() (dir string) {
	return p.Name
}
