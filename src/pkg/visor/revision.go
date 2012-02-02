package visor

import (
	"fmt"
	"time"
)

// A Revision represents an application revision,
// identifiable by its `ref`.
type Revision struct {
	App *App
	ref string
}

// NewRevision returns a new instance of Revision.
func NewRevision(app *App, ref string) (rev *Revision, err error) {
	rev = &Revision{App: app, ref: ref}
	return
}

// Register registers a new Revision with the registry.
func (r *Revision) Register(c *Client) (err error) {
	exists, err := c.Exists(r.Path())
	if err != nil {
		return
	}
	if exists {
		return ErrKeyConflict
	}

	err = c.Set(r.Path()+"/registered", []byte(time.Now().UTC().String()))
	if err != nil {
		return
	}
	err = c.Set(r.Path()+"/scale", []byte("0"))

	return
}

// Unregister unregisters a revision from the registry.
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

// Path returns this revision's directory path in the registry.
func (r *Revision) Path() string {
	return r.App.Path() + "/revs/" + r.ref
}

func (r *Revision) String() string {
	return fmt.Sprintf("%#v", r)
}

// Revisions returns an array of all registered revisions.
func Revisions(c *Client) (revisions []*Revision, err error) {
	apps, err := Apps(c)
	if err != nil {
		return
	}

	revisions = []*Revision{}

	for i := range apps {
		revs, e := AppRevisions(c, apps[i])
		if e != nil {
			return nil, e
		}
		revisions = append(revisions, revs...)
	}

	return
}

// AppRevisions returns an array of all registered revisions belonging
// to the given application.
func AppRevisions(c *Client, app *App) (revisions []*Revision, err error) {
	refs, err := c.Keys(app.Path() + "/revs")
	if err != nil {
		return
	}
	revisions = make([]*Revision, len(refs))

	for i := range refs {
		r, e := NewRevision(app, refs[i])
		if e != nil {
			return nil, e
		}

		revisions[i] = r
	}

	return
}
