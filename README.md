# Visor

Interaction layer for SoundClouds global process state referred to as registry.

## Usage

``` go
package main

import "soundcloud/visor"

func main() {
  client, err := visor.Dial("coordinator:8046")
  if err != nil {
    panic(err)
  }

  c := make(chan *visor.Event)

  go client.WatchEvent(c)

  // reading one event from the channel
  e := <-c
  fmt.Printf("%s", e.String())
}
```

## Visor API

### Dial(addr string) (*Client, error)

Establishes a connection to the registry state and returns a `Client`.

## Event

## Ticket

## Client API

### (c *Client) Close() (error)

Disconnects `Client` gracefully.

### (c *Client) Apps() ([]visor.App, error)

Returns all `Apps` registered in registry.

### (c *Client) RegisterApp(**TODO**) (*visor.App, error)

Registers a new application with the registry.

### (c *Client) UnregisterApp(app *visor.App) (error)

Removes application from the registry.

### (c *Client) Instances() ([]visor.Instance, error)

Returns all Instances registered.

### (c *Client) HostInstances(addr string) ([]visor.Instance, error)

Returns all Instances running on `addr`.

### (c *Client) Tickets() ([]visor.Ticket, error)

Returns all Tickets.

### (c *Client) HostTickets(addr string) ([]visor.Ticket, error)

Returns all Tickets claimed by `addr`.

### (c *Client) WatchEvent(ch chan *visor.Event) (error)

Watches for new `Events` inside of the registry.

### (c *Client) WatchTicket(ch chan *visor.Ticket) (error)

Watch for new `Ticket` created.

## App API

### (a *App) Register() (error)

Registers the `App` in the registry.

### (a *App) Unregister() (error)

Removes application from the registry.

### (a *App) Revisions() ([]visor.Revision, error)

Returns all `Revisions` for the `App`.

### (a *App) RegisterRevision(rev string) (*visor.Revision, error)

Registers a new `Revision` for the `App`.

### (a *App) UnregisterRevision(r *visor.Revision) (error)

Removes a `Revision` from the `App`.

### (a *App) EnvironmentVariables() (*map[string]string, error)

Returns the stored `Environment` as a `Map`.

### (a *App) GetEnvironmentVariable(k string) (string, error)

Returns the value for the variable stored at `k`.

### (a *App) SetEnvironmentVariable(k string, v string) (error)

Stores the value `v` for the key `k`.

## Revision API

### (r *Revision) Register() (error)

Registers the `Revision` for it's `App`.

### (r *Revision) Unregister() (error)

Removes the `Revision` from it's `App`.

### (r *Revision) Scale(p string, s int) (error)

Sets the scaling factor of the process type `p` to the amount of `s`.

### (r *Revision) Instances() ([]visor.Instance, error)

Returns all `Instances` for the `Revision`.

### (r *Revision) RegisterInstance(p string, addr string) (*visor.Instance, error)

Registers new `Instance` for `Revision`.

### (r *Revision) UnregisterInstance(*visor.Instance) (error)

Remvoes the `Instance` from the `Revision`.

## Instance API

``` go
type Instance struct {
  Rev: *visor.Instance
  Host: string
  Port: int
  State: visor.State
  ProcessType: string
}
```

### (i *Instance) Register() (error)

Registers the `Instance` for it's `Revision`.

### (i *Instance) Register() (error)

Removes the `Instance` from it's `Revision`.

## Development

### Setup

#### Dependencies

#### Installation

#### Run

### Testing

### Branching

See [Guide](https://github.com/soundcloud/soundcloud/wiki/conventions-git#wiki-using-git-flow).

### Versioning

This project is versioned with the help of the [Semantic Versioning Specification](http://semver.org/) using `0.0.0` as the initial version. Please make sure you have read the guidelines before increasing a version number either for a release or a hotfix.
