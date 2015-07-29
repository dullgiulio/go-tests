package main

// TODO:
//  - Remove usage of pool here; pool is not the right object.

import (
	"bytes"
	"io"
	"os"
	"sync"
)

type buffersPool struct {
		// pool of writable buffers
		wp    sync.Pool
		// pool of readable buffers
		rp    sync.Pool
		ws    sync.WaitGroup
		rs    sync.WaitGroup
		ready chan struct{}
}

func newBuffersPool() *buffersPool {
	return &buffersPool{
		ready: make(chan struct{}),
	}
}

// eof must be called when there is no more data to write
func (p *buffersPool) eof() {
	p.ws.Wait()
	close(p.ready)
}

// wait waits for the readers to finish
func (p *buffersPool) wait() {
	p.rs.Wait()
}

func (p *buffersPool) reader(l loadProc) {
	p.rs.Add(1)
	go func() {
		defer p.rs.Done()
		for range p.ready {
			b := p.rp.Get().(*bytes.Buffer)
			l.proc(b)
			// This buffer is writable again
			p.wp.Put(b)
		}
	}()
}

func (p *buffersPool) writer(l loadProc) {
	p.ws.Add(1)
	p.wp.Put(&bytes.Buffer{})
	go func() {
		for {
			b := p.wp.Get().(*bytes.Buffer)
			if err := l.load(b); err == io.EOF {
				p.ws.Done()
				p.eof()
				return
			}
			p.rp.Put(b)
			// Wake up a waiting reader
			p.ready <- struct{}{}
		}
	}()
}

type loadProc interface {
	// alloc returns a newly allocated instance
	new() interface{}
	// load write into a the data to be processed.
	//
	// load returns io.EOF when no more data is available.  Any other error is ignored.
	load(a interface{}) error
	// proc processes the data in a
	//
	// proc might need to make a ready for reuse.
	proc(a interface{})
}

type procBuffer struct{
	resp chan string
}

func (p *procBuffer) new() interface{} {
	return &bytes.Buffer{}
}

func (p *procBuffer) load(a interface{}) error {
	response := <-p.resp
	if response == "" {
		return io.EOF
	}
	b := a.(*bytes.Buffer)
	io.WriteString(b, response)
	return nil
}

func (p *procBuffer) proc(a interface{}) {
	b := a.(*bytes.Buffer)
	io.Copy(os.Stdout, b)
	b.Reset()
}

func main() {
	p := newBuffersPool()
	pb := &procBuffer{
		resp: make(chan string, 3),
	}
	pb.resp <- "hello"
	pb.resp <- " world"
	pb.resp <- ""
	p.reader(pb)
	p.writer(pb)
	p.wait()
}
