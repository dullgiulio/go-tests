package main

import "github.com/dullgiulio/sima"

type Plugin struct{}

func (p *Plugin) SayHello(name string, msg *string) error {
	*msg = fmt.Sprintf("Hello %s", name)
	return nil
}

func main() {
	plugin := &Plugin{}

	sima.Register(plugin)
	sima.Run()
}
