# Boober

Interaction layer for SoundClouds global process state.

## Usage

``` go
package main

import "soundcloud/boober"

func main() {
  client, err := boober.Dial("coordinator:8046")
  if (err != nil) {
    panic(err)
  }

  c := make(chan boober.Event)

  go client.WatchEvent(c)

  // reading one event from the channel
  e := <-c
  fmt.Printf("%s", e.String())
}
```

## API

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
