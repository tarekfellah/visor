// Visor is a doozer client which provides an abstract interface
// to a doozer cluster containing global process state information.
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
//         client, err := visor.Dial("coordinator:8046", "/")
//         if err != nil {
//           panic(err)
//         }
//
//         l := make(chan *visor.Event)
//
//         // Watch for changes in the global process state
//         go visor.WatchEvent(client, l, 0)
//
//         for {
//             e := <-l
//             fmt.Println(e)
//         }
//     }
//
package visor

const DEFAULT_ADDR string = "localhost:8046"
const DEFAULT_ROOT string = "/visor"

type ProcessType string
type Stack string
type State int

// Dial connects to the coordinator over 'tcp'
func Dial(addr string, root string, codec Codec) (c *Client, err error) {
	snapshot, err := DialConn(addr, root)
	if err != nil {
		return
	}

	c = NewClient(snapshot.conn, "/", snapshot.Rev, codec)

	return
}
