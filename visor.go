// Visor is a library which provides an abstract interface
// over a global process state.
//
// This process state is referred to as the registry.
//
// Example usage:
//
//     package main
//
//     import "soundcloud/visor"
//
//     func main() {
//         client, err := visor.Dial("coordinator:8046", "/", new(visor.StringCodec))
//         if err != nil {
//           panic(err)
//         }
//
//         l := make(chan *visor.Event)
//
//         // Watch for changes in the global process state
//         go visor.WatchEvent(client.Snapshot, l)
//
//         for {
//             fmt.Println(<-l)
//         }
//     }
//
package visor

import (
	"path"
)

const DEFAULT_ADDR string = "localhost:8046"
const DEFAULT_ROOT string = "/visor"

type ProcessName string
type Stack string
type State int

func (s State) String() string {
	switch s {
	case InsStateInitial:
		return "initial"
	case InsStateStarted:
		return "started"
	case InsStateReady:
		return "ready"
	case InsStateFailed:
		return "failed"
	case InsStateDead:
		return "dead"
	case InsStateExited:
		return "exited"
	}
	return "?"
}

func ProcPath(app string, revision string, processName string, attributes ...string) string {
	// ...
	return path.Join(append([]string{APPS_PATH, app, REVS_PATH, revision, PROCS_PATH, processName}, attributes...)...)
}
