package main

import (
	"fmt"
	"github.com/dullgiulio/sima"
	"log"
)

func runPlugin(proto, path string) {
	p := sima.NewPlugin(proto, path)
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
	protocols := []string{"unix", "tcp"}
	for _, p := range protocols {
		fmt.Println("Running hello world plugin")

		runPlugin(p, "bin/examples/pingo-hello-world")

		fmt.Println("Plugin terminated.")
	}

	fmt.Println("Running plugin that fails to register in time")

	runPlugin("tcp", "bin/examples/pingo-sleep")

	fmt.Println("Plugin terminated.")
}
