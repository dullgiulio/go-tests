package sima

import (
	"flag"
)

type config struct {
	host string
}

func makeConfig() *config {
	c := &config{}
	flag.StringVar(&c.host, "sima:host", ":8888", "Where to listen to for RPC calls")
	flag.Parse()
	return c
}
