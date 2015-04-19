package sima

import (
	"net/rpc"
)

type Client interface {
	Start() error
	Stop() error
	Call(name string, n int, data interface{}) error
}

type ClientHTTP struct {
	client *rpc.Client
	proto  string
	addr   string
}

func NewClientHTTP(addr string) *ClientHTTP {
	return &ClientHTTP{proto: "tcp", addr: addr}
}

func (c *ClientHTTP) Start() (err error) {
	c.client, err = rpc.DialHTTP(c.proto, c.addr)
	return
}

func (c *ClientHTTP) Stop() error {
	return c.client.Close()
}

func (c *ClientHTTP) Call(name string, n int, data interface{}) error {
	return c.client.Call(name, n, data)
}
