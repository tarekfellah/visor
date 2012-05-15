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
	"strconv"
)

const DEFAULT_ADDR string = "localhost:8046"
const DEFAULT_ROOT string = "/visor"
const START_PORT int = 8000
const START_PORT_PATH string = "/next-port"

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

func Init(s Snapshot) (rev int64, err error) {
	exists, _, err := s.Conn().Exists(START_PORT_PATH, &s.Rev)
	if err != nil {
		return
	}

	if !exists {
		rev, err = s.Conn().Set(START_PORT_PATH, s.Rev, []byte(strconv.Itoa(START_PORT)))
		if err != nil {
			return
		}

		return rev, err
	}
	return s.conn.Rev()
}

func ProcPath(app string, revision string, processName string, attributes ...string) string {
	// ...
	return path.Join(append([]string{APPS_PATH, app, REVS_PATH, revision, PROCS_PATH, processName}, attributes...)...)
}
