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

  c := make(chan visor.Event)

  go client.WatchEvent(c)

  // reading one event from the channel
  e := <-c
  fmt.Printf("%s", e.String())
}
```

## Visor API

### Dial(addr string) (*Client, error)

Establishes a connection to the registry state and returns a `Client`.

## Client API

### (c *Client) Close() (error)

Disconnects `Client` gracefully.

### (c *Client) Apps() ([]App, error)

Returns all `Apps` registered in registry.

### (c *Client) RegisterApp(**TODO**) (*App, error)

Registers a new application with the registry.

### (c *Client) UnregisterApp(app *App) (error)

Removes application from the registry.

### (c *Client) Instances() ([]Instance, error)

Returns all Instances registered.

### (c *Client) HostInstances(addr string) ([]Instance, error)

Returns all Instances running on `addr`.

### (c *Client) Tickets() ([]Ticket, error)

Returns all Tickets.

### (c *Client) HostTickets(addr string) ([]Ticket, error)

Returns all Tickets claimed by `addr`.

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
