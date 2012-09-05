package visor

type dir struct {
	Snapshot
	Name string
}

func (p *dir) Get(key string) (string, int64, error) {
	return p.Snapshot.get(p.Prefix(key))
}

func (p *dir) Set(key string, val string) (int64, error) {
	s, err := p.Snapshot.set(p.Prefix(key), val)
	return s.Rev, err
}

func (p *dir) Del(key string) error {
	return p.Snapshot.del(p.Prefix(key))
}

func (p *dir) Prefix(path string, paths ...string) (result string) {
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

func (p *dir) String() (dir string) {
	return p.Name
}
