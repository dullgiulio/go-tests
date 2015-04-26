package main

import (
	"fmt"
	"github.com/dullgiulio/sima"
	//"os"
	//"time"
)

type Plugin struct{}

func (p *Plugin) SayHello(name string, msg *string) error {
	fmt.Printf("Called rpc: %s", name)
	*msg = fmt.Sprintf("Hello %s", name)
	return nil
}

func main() {
	plugin := &Plugin{}

	/*
		go func() {
			<-time.After(3 * time.Second)
			os.Exit(0)
		}()
	*/

	sima.Register(plugin)
	sima.Run()
}
