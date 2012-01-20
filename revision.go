package visor

type Revision struct {
}

func (r *Revision) Register() error {
	return nil
}
func (r *Revision) Unregister() error {
	return nil
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
