# Visor

Visor is a library which provides an abstraction over a global process state.

Visor uses [doozerd](http://github.com/soundcloud/doozerd).

## Usage

To understand how Visor works, we need to understand how it works with *time*. Each
of the Visor data-types *App*, *Revision*, *Proctype* and *Instance* are snapshots of
a specific point in time in the coordinator. When a mutating operation is successfully
performed on one of these data-types, a **new snapshot** is returned, representing the state
of the coordinator *after* the operation. If the operation would fail, the old snapshot is
returned with an error.

With the new snapshot, we can perform an operation on this new state, and so on with
every new snapshot.

```go
package main

import "github.com/soundcloud/visor"
import "log"

func main() {
    snapshot, err := visor.DialUri("doozer:?ca=localhost:8046", "/")

    name  := "rocket"
    stack := "HEAD" // Runtime stack revision
    repo  := "http://github.com/bazooka/rocket"

    app   := visor.NewApp(name, repo, stack, snapshot)

    _, err := app.Register()
    if err != nil {
        log.Fatalf("error registering app: %s", err)
    }

    rev := visor.NewRevision(app, "f84e19", snapshot)
    rev.ArchiveUrl = "http://artifacts/rocket/f84e19.img"

    _, err = rev.Register()
    if err != nil {
        log.Fatalf("error registering revision: %s", err)
    }
}
```

### Working with snapshots

```go
// Get a snapshot of the latest coordinator state
snapshot, err := visor.DialUri("doozer:?ca=localhost:8046", "/")

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

### Watching for events

``` go
package main

import "soundcloud/visor"

func main() {
  snapshot, err := visor.DialUri("doozer:?ca=localhost:8046", "/")
  if err != nil {
    panic(err)
  }

  c := make(chan *visor.Event)

  go visor.WatchEvent(snapshot, c)

  // Read one event from the channel
  fmt.Println(<-c)
}
```

## Development

### Setup

#### Dependencies

  - [Go](http://golang.org) (go1)
    
  - [doozer](https://github.com/soundcloud/doozer) (implicit)

        go get github.com/soundcloud/doozer
   
    if you run in trouble with the protobuf, do a:

        cd src/pkg/code.google.com/p/goprotobuf
        hg pull
        hg update 
        make install

  - [doozerd](https://github.com/soundcloud/doozerd) (testing)

        go get github.com/ha/doozerd

    if this fails, do the following

        cd $GOROOT/src/pkg/github.com/soundcloud/doozerd
        git remote add soundcloud git@github.com:soundcloud/doozerd
        git pull soundcloud master
        ./make.sh
        go install

#### Installation

Debian in our internal network:

    apt-get install visor

Compile yourself:

 * install [golang](http://golang.org) and `make install`

### Testing

First start `doozerd` with default configuration. Then run:

```
go test
```

### Conventions

This repository follows the code conventions dictated by [gofmt](http://golang.org/cmd/gofmt/). To automate the formatting process install this [pre-commit hook](https://gist.github.com/e689d5de0982543cce8c), which runs `gofmt` and adds the files. Don't forget to make the file executable: `chmod +x .git/hooks/pre-commit`.

### Branching

See [Guide](https://github.com/soundcloud/soundcloud/wiki/conventions-git#wiki-using-git-flow).

### Versioning

This project is versioned with the help of the [Semantic Versioning Specification](http://semver.org/) using `0.0.0` as the initial version. Please make sure you have read the guidelines before increasing a version number either for a release or a hotfix.
