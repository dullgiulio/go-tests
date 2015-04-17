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
	host   string
}

func NewClientHTTP(host string) *ClientHTTP {
	return &ClientHTTP{host: host}
}

func (c *ClientHTTP) Start() (err error) {
	c.client, err = rpc.DialHTTP("tcp", c.host)
	return
}

func (c *ClientHTTP) Stop() error {
	return c.client.Close()
}

func (c *ClientHTTP) Call(name string, n int, data interface{}) error {
	return c.client.Call(name, n, data)
}
