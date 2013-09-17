# immutable env design

**authors**: @xla  
**status**: implemented  
**date**: 2013-07-15  

### tl;dr
* control on instance level
* immutable
* configuration creation api
* atomic configuration changes
* reasoning about runtime behaviour

### reasoning
The premise is to move away from an app global, mutable and non-atomic alterable implementation towards an immutable, trackable implementation. Primary goal is the ease of reasoning. Secondary goal is to provide a small api surface for configuration handling in bazooka.

### current
#### visor
Currently the env of running instance is inherited by their respective applications. Mutations are done on an `App` object and on `key` level, which allows partial mutation of the whole `env`. Underneath it's a member of `App` implemented as map[string]string. There is no mutation history stored. The api methods are:

``` go
func (a *App) DelEnvironmentVar(k string) (*App, error)
func (a *App) EnvironmentVars() (vars Env, err error)
func (a *App) GetEnvironmentVar(k string) (value string, err error)
```

The resulting file structure looks like the following:

```
/apps/<APP>/env/<KEY0> = <VALUE0>
                <KEY1> = <VALUE1>
```

#### cli
The cli is a base exposure of the visor api methods. Except for the retrieval of the entire env all cli calls operate on single keys:

```
$ # get entire env
$ bazooka env get
  KEY0=VALUE0
  KEY1=VALUE1
$ # get single key
$ bazooka env get KEY1
  VALUE1
$ # set single key
$ printf "VALUE2" | bazooka env set KEY2
$ # del single key
$ bazooka env del KEY1
```

#### runtime
Currently configuration is passed to the running instance via its environment and can be accessed through the runtime's mechanism of reading the process environment. Thus the same limitations which apply to environ keys and variables apply to the keys and variables of bazooka env. Those limitations are implicit and not part of the contract.

### future
#### visor
The new implementation introduces envs as first class types. Identifiable by their `ref` envs are created in the context of an `App`. It will be guaranteed that an `Env` registered with its ref won't be overwritten or changed until removed with `Unregister`. `Register` will also fail if one of the keys is either of length zero or contains `=`. Additionally `Scale` gets another parameter to control which `Env` should be used on ticket creation.

``` go
func (a *App) NewEnv(ref string, vars map[string]string) *Env
func (a *App) GetEnv(ref string) (*Env, error)
func (a *App) GetEnvs() ([]*Env, error)

type Env struct {
  dir        *cp.Dir
  App        *App
  Ref        string
  Vars       map[string]string
  Registered time.Time
}
func (e *Env) Register() (*Env, error)
func (e *Env) Unregister() error
func (s *Store) Scale(app, rev, proc, env string, factor int) ([]*Instance, int, error)
func (s *Store) RegisterInstance(app, rev, proc, env string) (*Instance, error)
```

This should result in a file structure like this:
```
/apps/<APP>/envs/REF0/
                  vars       = {"KEY0":"VALUE0","KEY1":"VALUE1"}
                  registered = "2013-08-22T21:20:17+02:00"
                 REF1/
                  vars       = {"KEY0":"VALUE0","KEY1":"VALUE1"}
                  registered = 2013-08-22T21:20:18+02:00 
```
#### cli
The cli is shifting from key level CRUD-like operations to env-level. Outputs which show instance listings will be extended by one column showing the `Env` of the instance. The command to delete env (`delete`) will guard against the case of an env deletion while there are still instances running which are configured with this env.

```
$ # get all env list
$ bazooka env list
REF      REGISTERED
0000000  0000-00-00
…
$ # create single env
$ cat PATH_TO_ENV | bazooka env create REF
$ # create single env from json formatted input
$ printf '{"KEY0":"VALUE0","KEY1":"VALUE1"}' | bazooka env create -format="json" REF
$ # get single env
$ bazooka env get REF
  KEY0=VALUE0
  KEY1=VALUE1
$ # get single env json representation
$ bazooka env get -format="json" REF
{"KEY0":"VALUE0","KEY1":"VALUE1"}
$ # delete an env
$ bazooka env delete REF
```

The scale command will take another option which is the env to explicitly scale with a certain env. If no env is given the cli will choose the latest created env(registered-at timestamp). If an env is given, which doesn't exist the command will fail.

```
$ # scale with env
$ bazooka scale -e REF -r REV -n FACTOR PROC
PROC@REV@REF 0 -> FACTOR ...
```
#### runtime
The contract between scheduler and instances is still satisfied via environment variables. With the new implementation it is guaranteed that KEYS are validated and won't break during runtime.

#### migration
All current app envs vars need to be put an initial new env. This new will be used for all further scaling commands until another env is created by the user. It is advisable to reschedule all instances on updated pms to reflect the new env handling per instance. During the transition period pms need to support both ways to acquire and pass an env to an instance.

To avoid inconsistencies the user-level access to bazooka components should be blocked while the env migration is in process.

It should be considered to update the schema version of the bazooka store to force upgrades of the cli.
