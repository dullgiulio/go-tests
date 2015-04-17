package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"github.com/dullgiulio/sima"
)

type Plugin struct{}

func NewPlugin() *Plugin{
	return &Plugin{}
}

func (p *Plugin) SimaPluginRegister(unused int, methods *sima.Methods) error {
	methods.Add("HelloWorld")
	return nil
}

func (p *Plugin) HelloWorld(name string, unused *int) error {
	log.Print("Name is: ", name)
	return nil
}

func main() {
	plugin := NewPlugin()
	rpc.Register(plugin)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":8888")
	if e != nil {
			log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}
