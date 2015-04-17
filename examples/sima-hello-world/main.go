package main

import (
	"github.com/dullgiulio/sima"
	"log"
)

type Plugin struct{}

func (p *Plugin) HelloWorld(name string, unused *int) error {
	log.Print("Name is: ", name)
	return nil
}

func main() {
	plugin := &Plugin{}

	sima.Register(plugin)
	sima.Run()
}
