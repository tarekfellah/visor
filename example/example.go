package main

import (
	"fmt"
	"visor"
)

func main() {
	addr := visor.DEFAULT_ADDR
	root := visor.DEFAULT_ROOT
	snapshot, err := visor.Dial(addr, root)

	if err != nil {
		panic(err)
	}

	channel := make(chan *visor.Event)

	go visor.WatchEvent(snapshot, channel)

	fmt.Println(<-channel)

	for {
		select {
		case e := <-channel:
			fmt.Println(e)
		}
	}
}
