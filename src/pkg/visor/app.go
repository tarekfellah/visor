package visor

type App struct {
	Name    string
	RepoUrl string
	Stack   Stack
}

func (a *App) Revisions() []Revision {
	return nil
}
func (a *App) RegisterRevision(rev string) (*Revision, error) {
	return nil, nil
}
func (a *App) UnregisterRevision(r *Revision) error {
	return nil
}
func (a *App) Register() error {
	return nil
}
func (a *App) Unregister() error {
	return nil
}
func (a *App) EnvironmentVariables() (*map[string]string, error) {
	return nil, nil
}
func (a *App) GetEnvironmentVariable(k string) (string, error) {
	return "", nil
}
func (a *App) SetEnvironmentVariable(k string, v string) error {
	return nil
}
