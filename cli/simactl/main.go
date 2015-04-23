package main

import (
	"fmt"
	"github.com/dullgiulio/sima"
	"log"
)

func main() {
	done := make(chan bool)
	m := sima.NewManager(func() {
		done <- true
	})

	// TODO: Request an object from a plugin.
	p := sima.NewPlugin("bin/examples/sima-hello-world")
	m.Register(p)

	if resp, err := m.Call("Plugin.SayHello", "Giulio"); err != nil {
		log.Print(err)
	} else {
		fmt.Printf("%s\n", resp)
	}

	m.Stop()

	<-done
}
