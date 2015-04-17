package main

import (
	"github.com/dullgiulio/sima"
)

func main() {
	done := make(chan bool)
	m := sima.NewManager(func() {
		done <- true
	})

	p := sima.NewPlugin("SimaHelloWorld", sima.NewClientHTTP("localhost:8888"))
	m.RegisterPlugin(p)
	m.Debug(p)
	m.Stop()

	<-done
}
