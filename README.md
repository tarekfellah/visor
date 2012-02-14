# Visor

Visor is a library which provides an abstract interface over a global process state.

## Usage

... snapshots

### Working with snapshots

```go
// Get a snapshot of the latest coordinator state
snapshot, err := visor.DialConn("coordinator:8046", "/")

// Get the list of applications at snapshot
apps, _ := visor.Apps(snapshot)
app := apps[0] // app.Rev == snapshot.Rev == 1

// Set some environment vars on *app*. Every time state is
// changed in the coordinator, a new App snapshot is returned.
app, _ = app.SetEnvironmentVar("cow", "moo")  // app.Rev == 2
app, _ = app.SetEnvironmentVar("cat", "meow") // app.Rev == 3

// Attempt to get a recently set environment var from an old snapshot (apps[0].Rev == 1)
apps[0].GetEnvironmentVar("cat") // "", ErrKeyNotFound

// Get a recently set environment var from the latest snapshot (app.Rev == 3)
app.GetEnvironmentVar("cat")     // "meow", nil

```

### Advancing in time

```go
// Get a snapshot of the latest coordinator state
snapshot, err := visor.DialConn("coordinator:8046", "/")

apps, _ := visor.Apps(snapshot) // len(apps) == 0

app, _ := NewApp("soundcloud.com", "git://github.com/sc/soundcloud.com", "mystack", snapshot)
app.Register()

// *snapshot* still refers to the old state, so apps is still empty
apps, _ := visor.Apps(snapshot) // len(apps) == 0

// Get a snapshot of the latest coordinator state
snapshot = snapshot.FastForward(-1)

// Now that snapshot reflects the latest state, apps contains our registered app
apps, _ := visor.Apps(snapshot) // len(apps) == 1
```

### Watching for events

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
