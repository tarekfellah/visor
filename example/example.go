package main

import (
	"fmt"
	"visor"
)

func main() {
	client, err := visor.Dial("localhost:8046/bazooka")

	if err != nil {
		panic(err)
	}

	channel := make(chan *visor.Event)

	go client.WatchEvent(channel)

	for {
		select {
		case e := <-channel:
			fmt.Println(e)
		}
	}
}
