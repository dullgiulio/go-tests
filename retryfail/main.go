package main

import (
	"log"
	"math/rand"
)

type retryWorker struct {
	workers  int
	in       chan int
	failed   chan int
	finished chan struct{}
	exit     chan struct{}
}

func newRetryWorker(workers int) *retryWorker {
	r := &retryWorker{
		workers:  workers,
		in:       make(chan int, 0),
		failed:   make(chan int, workers), // Any buffering size, really
		finished: make(chan struct{}, workers-1),
		exit:     make(chan struct{}, 0),
	}
	for i := 0; i < workers-1; i++ {
		go r.run()
	}
	go r.retry()
	return r
}

func (r *retryWorker) run() {
	for i := range r.in {
		if !r.calc(i) {
			r.failed <- i
		}
	}
	r.finished <- struct{}{}
}

func (r *retryWorker) retry() {
	go func() {
		for i := 0; i < r.workers-1; i++ {
			<-r.finished
		}
		close(r.failed)
	}()
	for i := range r.failed {
		if !r.calc(i) {
			log.Printf("failed twice to calc(%d), not retrying", i)
		}
	}
	r.exit <- struct{}{}
}

func (r *retryWorker) calc(i int) bool {
	if i%(rand.Intn(r.workers)+1) == 0 {
		return false
	}
	return true
}

func (r *retryWorker) do(i int) {
	r.in <- i
}

func (r *retryWorker) done() {
	close(r.in)
	<-r.exit
}

func main() {
	r := newRetryWorker(4)
	for i := 0; i < 100; i++ {
		r.do(rand.Intn(100))
	}
	r.done()
}
