package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type dialer func(n, a string) (net.Conn, error)

var stdDialer dialer

func init() {
	stdDialer = http.DefaultTransport.(*http.Transport).Dial
}

type srcHttpClient struct {
	client *http.Client
	addr   net.Addr
}

func newSrcHttpClient() *srcHttpClient {
	s := &srcHttpClient{}
	s.client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			Dial:              s.dial,
		},
	}
	return s
}

func (s *srcHttpClient) dial(n, a string) (net.Conn, error) {
	conn, err := stdDialer(n, a)
	if err == nil {
		s.addr = conn.RemoteAddr()
	}
	return conn, err
}

func main() {
	s := newSrcHttpClient()
	resp, err := s.client.Get("http://www.google.com/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	fmt.Printf("%v\n", s.addr)
}
