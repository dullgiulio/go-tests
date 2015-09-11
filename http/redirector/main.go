package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type Redirect struct {
	StatusCode int
	URL        *url.URL
}

func (r *Redirect) String() string {
	return fmt.Sprintf("%d %s", r.StatusCode, r.URL)
}

type Redirector struct {
	rdirs map[*http.Request]*Redirect
	mux   sync.Mutex
}

func NewRedirector() *Redirector {
	return &Redirector{
		rdirs: make(map[*http.Request]*Redirect),
	}
}

func (r *Redirector) Get(req *http.Request) *Redirect {
	r.mux.Lock()
	defer r.mux.Unlock()

	redir := r.rdirs[req]
	delete(r.rdirs, req)
	return redir
}

func (r *Redirector) Put(req *http.Request, redir *Redirect) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.rdirs[req] = redir
}

func (r *Redirector) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		redir := &Redirect{}
		redir.StatusCode = resp.StatusCode
		if location := resp.Header.Get("Location"); location != "" {
			if redir.URL, err = url.Parse(location); err != nil {
				return
			}
		}
		r.Put(req, redir)
	}
	return
}

func (r *Redirector) CheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if len(via) > 0 {
		// Update entry to reflect actual request visible to client
		if rdir := r.Get(via[len(via)-1]); rdir != nil {
			r.Put(req, rdir)
		}
	}
	return nil
}

func main() {
	flag.Parse()
	rt := NewRedirector()
	client := http.Client{
		Transport:     rt,
		CheckRedirect: rt.CheckRedirect,
	}
	resp, err := client.Get(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if redir := rt.Get(resp.Request); redir != nil {
		fmt.Printf("%s\n", redir)
	}
}
