package main

import (
	"fmt"
	"github.com/dullgiulio/sima"
	"log"
)

func runPlugin(s string) {
	p, err := sima.NewPlugin("unix", s)
	if err != nil {
		log.Fatal(err)
	}
	p.Start()
	defer p.Stop()

	objs, err := p.Objects()
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Printf("Objects: %s\n", objs)

	var resp string

	if err := p.Call("Plugin.SayHello", "Giulio", &resp); err != nil {
		log.Print(err)
	} else {
		fmt.Printf("%s\n", resp)
	}
}

func main() {
	runPlugin("bin/examples/sima-hello-world")
	runPlugin("bin/examples/sima-sleep")
}
