package visor

import "net"

type Instance struct {
	Rev         *Revision
	Addr        net.TCPAddr
	State       State
	ProcessType ProcessType
}

func (i *Instance) String() string {
	return "<instance>"
}
func (i *Instance) Register() error {
	return nil
}
func (i *Instance) Unregister() error {
	return nil
}
