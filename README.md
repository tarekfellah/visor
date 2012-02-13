# Visor

Interaction layer for SoundClouds global process state referred to as registry.

## Usage

``` go
package main

import "soundcloud/visor"

func main() {
  client, err := visor.Dial("coordinator:8046", "/", new(visor.ByteCodec))
  if err != nil {
    panic(err)
  }

  c := make(chan *visor.Event)

  go visor.WatchEvent(client.Snapshot, c)

  // Read one event from the channel
  fmt.Println(<-c)
}
```

## Visor API

### Dial(addr string) (*Client, error)

Establishes a connection to the registry state and returns a `Client`.

## Event

Abstaction of an activity in the registry.

``` go
type Event struct {
  Type EventType
  Body string
  Source *doozer.Event
}
```

### (ev *Event) String() string

Returns human readable representation of the `Event`.

## EventTYpe

``` go
type EventType int

const (
  EV_APP_REG = iota
  EV_APP_UNREG
  EV_REV_REG
  EV_REV_UNREG
  EV_INS_REG
  EV_INS_UNREG
  EV_INS_STATE_CHANGE
)
```

## Ticket

``` go
type Ticket struct {
  Type visor.TicketType
  App *visor.App
  Rev *visor.Revison
  ProcessType visor.ProcessType
  Addr net.TCPAddr
  Source *doozer.Event
}
```

## TicketType

``` go
type TicketType int

const (
  T_START = iota
  T_STOP
)
```

## ProcessType

``` go
type ProcessType string
```

## Client API

``` go
type Client struct
```

### (c *Client) Close() error

Disconnects `Client` gracefully.

### (c *Client) Apps() ([]visor.App, error)

Returns all `Apps` registered in registry.

### (c *Client) RegisterApp(rUrl url.Url, stack string) (*visor.App, error)

Registers a new application with the registry.

### (c *Client) UnregisterApp(app *visor.App) error

Removes application from the registry.

### (c *Client) Instances() ([]visor.Instance, error)

Returns all Instances registered.

### (c *Client) HostInstances(addr string) ([]visor.Instance, error)

Returns all Instances running on `addr`.

### (c *Client) Tickets() ([]visor.Ticket, error)

Returns all Tickets.

### (c *Client) HostTickets(addr string) ([]visor.Ticket, error)

Returns all Tickets claimed by `addr`.

### (c *Client) WatchEvent(ch chan *visor.Event) error

Watches for new `Events` inside of the registry.

### (c *Client) WatchTicket(ch chan *visor.Ticket) error

Watch for new `Ticket` created.

## App API

``` go
type App struct {
  RepoUrl url.URL
  Stack string
}
```

### (a *App) Register() error

Registers the `App` in the registry.

### (a *App) Unregister() error

Removes application from the registry.

### (a *App) Revisions() ([]visor.Revision, error)

Returns all `Revisions` for the `App`.

### (a *App) RegisterRevision(rev string) (*visor.Revision, error)

Registers a new `Revision` for the `App`.

### (a *App) UnregisterRevision(r *visor.Revision) error

Removes a `Revision` from the `App`.

### (a *App) EnvironmentVariables() (*map[string]string, error)

Returns the stored `Environment` as a `Map`.

### (a *App) GetEnvironmentVariable(k string) (string, error)

Returns the value for the variable stored at `k`.

### (a *App) SetEnvironmentVariable(k string, v string) error

Stores the value `v` for the key `k`.

## Revision API

``` go
type Revision struct {
  Rev string
}
```

### (r *Revision) Register() error

Registers the `Revision` for it's `App`.

### (r *Revision) Unregister() error

Removes the `Revision` from it's `App`.

### (r *Revision) Scale(p string, s int) error

Sets the scaling factor of the process type `p` to the amount of `s`.

### (r *Revision) Instances() ([]visor.Instance, error)

Returns all `Instances` for the `Revision`.

### (r *Revision) RegisterInstance(p string, addr string) (*visor.Instance, error)

Registers new `Instance` for `Revision`.

### (r *Revision) UnregisterInstance(*visor.Instance) error

Remvoes the `Instance` from the `Revision`.

## Instance API

``` go
type Instance struct {
  Rev *visor.Revision
  Addr net.TCPAddr
  State visor.State
  ProcessType visor.ProcessType
}
```

### (i *Instance) Register() error

Registers the `Instance` for it's `Revision`.

### (i *Instance) Register() error

Removes the `Instance` from it's `Revision`.

## Development

### Setup

#### Dependencies

[Go](http://golang.org) (weekly)
[go-gb](http://code.google.com/p/go-gb/) (weekly)
[Doozer](https://github.com/ha/doozer) (implicit)
[Doozerd](https://github.com/soundcloud/doozerd) (testing)

#### Installation

From the root of the project run `gb`:

```
gb -g
```

If you want to install WebReduce executables & packages into your `GOROOT` run:

```
gb -g -i
```

### Testing

First start `doozerd` with default configuration. If listening run:

```
gb -t
```

### Conventions

This repository follows the code conventions dictated by [gofmt](http://golang.org/cmd/gofmt/). To automate the formatting process install this [pre-commit hook](https://gist.github.com/e689d5de0982543cce8c), which runs `gofmt` and adds the files. Don't forget to make the file executable: `chmod +x .git/hooks/pre-commit`.

### Branching

See [Guide](https://github.com/soundcloud/soundcloud/wiki/conventions-git#wiki-using-git-flow).

### Versioning

This project is versioned with the help of the [Semantic Versioning Specification](http://semver.org/) using `0.0.0` as the initial version. Please make sure you have read the guidelines before increasing a version number either for a release or a hotfix.
