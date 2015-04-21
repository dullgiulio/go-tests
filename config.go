package sima

import (
	"flag"
)

type config struct {
	proto    string
	addr     string
	discover bool
}

func makeConfig() *config {
	c := &config{}
	flag.BoolVar(&c.discover, "sima:discover", false, "Discover this plugin")
	flag.StringVar(&c.proto, "sima:proto", "unix", "Protocol to use: unix or tcp")
	flag.StringVar(&c.addr, "sima:addr", "", "Where to listen to for RPC calls")
	flag.Parse()
	return c
}
