package main

import (
	"fmt"
	"visor"
)

func main() {
	addr := visor.DEFAULT_ADDR
	root := visor.DEFAULT_ROOT
	client, err := visor.Dial(addr, root, new(visor.ByteCodec))

	if err != nil {
		panic(err)
	}

	channel := make(chan *visor.Event)

	go visor.WatchEvent(client, channel)

	for {
		select {
		case e := <-channel:
			fmt.Println(e)
		}
	}
}
