package main

import (
	"github.com/dullgiulio/sima"
)

func main() {
	done := make(chan bool)
	m := sima.NewManager(func() {
		done <- true
	})

	// TODO: Request an object from a plugin.
	p := sima.NewPlugin("bin/examples/sima-hello-world")
	m.Register(p)

	<-done
}
