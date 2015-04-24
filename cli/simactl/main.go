package main

import (
	"fmt"
	"github.com/dullgiulio/sima"
	"log"
)

func main() {
	m := sima.NewManager()

	// TODO: Request an object from a plugin.
	p := sima.NewPlugin("bin/examples/sima-hello-world")
	m.Register(p)

	var resp string

	if err := m.Call("Plugin.SayHello", "Giulio", &resp); err != nil {
		log.Print(err)
	} else {
		fmt.Printf("%s\n", resp)
	}

	m.Stop()
}
