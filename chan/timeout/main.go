package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type future struct {
	ready chan struct{}
	data  string
}

func newFuture() *future {
	return &future{ready: make(chan struct{})}
}

func (f *future) set(data string) {
	f.data = data
	close(f.ready)
}

func (f *future) tryTimeout(t time.Duration) (string, bool) {
	select {
	case <-f.ready:
		return f.data, true
	case <-time.After(t):
		return "", false
	}
}

func (f *future) try() (string, bool) {
	select {
	case <-f.ready:
		return f.data, true
	default:
		return "", false
	}
}

func (f *future) get() string {
	<-f.ready
	return f.data
}

type msg struct {
	resp *future
	data string
}

func newMsg(data string) *msg {
	return &msg{
		data: data,
		resp: newFuture(),
	}
}

func slowListen(ch chan *msg, t time.Duration, wg *sync.WaitGroup) {
	for {
		m := <-ch
		if m == nil {
			break
		}
		time.Sleep(t)
		m.resp.set(fmt.Sprintf("resp: %s", m.data))
	}
	wg.Done()
}

func send(ch chan *msg, s string, t time.Duration, wg *sync.WaitGroup) error {
	var m *msg
	if s != "" {
		m = newMsg(s)
	}
	ch <- m
	if s == "" {
		wg.Done()
		return nil
	}
	resp, ok := m.resp.tryTimeout(t)
	if ok {
		fmt.Printf("direct: %s\n", resp)
		wg.Done()
		return nil
	} else {
		go func() {
			resp := m.resp.get()
			fmt.Printf("future: %s\n", resp)
			wg.Done()
		}()
		return errors.New("future")
	}
}

func main() {
	var wg sync.WaitGroup
	ch := make(chan *msg)
	wg.Add(3)
	go slowListen(ch, 2*time.Second, &wg)
	send(ch, "test1", 1*time.Second, &wg)
	send(ch, "", 1*time.Second, &wg)
	wg.Wait()
}
