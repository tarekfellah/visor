package visor

type Revision struct {
	App *App
	ref string
}

func NewRevision(app *App, ref string) (rev *Revision, err error) {
	rev = &Revision{App: app, ref: ref}
	return
}

func (r *Revision) Register(c *Client) (err error) {
	exists, err := c.Exists(r.Path())
	if err != nil {
		return
	}
	if exists {
		return ErrKeyConflict
	}

	err = c.Set(r.Path()+"/scale", "0")

	return
}
func (r *Revision) Unregister(c *Client) (err error) {
	return c.Del(r.Path())
}
func (r *Revision) Scale(proctype string, factor int) error {
	return nil
}
func (r *Revision) Instances(proctype ProcessType) ([]Instance, error) {
	return nil, nil
}
func (r *Revision) RegisterInstance(proctype ProcessType, address string) (*Instance, error) {
	return nil, nil
}
func (r *Revision) UnregisterInstance(instance *Instance) error {
	return nil
}

func (r *Revision) Path() string {
	return r.App.Path() + "/revs/" + r.ref
}
}
