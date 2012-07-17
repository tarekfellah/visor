package visor

type Path struct {
	Snapshot
	Dir string
}

func (p *Path) Get(key string) (string, int64, error) {
	return p.Snapshot.Get(p.Prefix(key))
}

func (p *Path) Set(key string, val string) (int64, error) {
	return p.Snapshot.Set(p.Prefix(key), val)
}

func (p *Path) Del(key string) error {
	return p.Snapshot.Del(p.Prefix(key))
}

func (p *Path) Prefix(path string, paths ...string) (fixed string) {
	if path == "/" {
		fixed = p.Dir
	} else {
		fixed = p.Dir + "/" + path
	}
	for _, p := range paths {
		fixed += "/" + p
	}
	return
}

func (p *Path) String() (dir string) {
	return p.Dir
}
