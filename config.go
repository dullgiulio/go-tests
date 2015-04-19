package sima

import (
	"flag"
)

type config struct {
	proto string
	addr  string
}

func makeConfig() *config {
	c := &config{}
	flag.StringVar(&c.proto, "sima:proto", "unix", "Protocol to use: unix or tcp")
	flag.StringVar(&c.addr, "sima:addr", "", "Where to listen to for RPC calls")
	flag.Parse()
	return c
}
