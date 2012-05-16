// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	"fmt"
	"github.com/soundcloud/visor"
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
			fmt.Println(e.Type)
		}
	}
}
