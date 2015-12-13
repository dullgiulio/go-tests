package main

import (
	"fmt"
	"sync"
)

type rslice struct {
	s []int
	l sync.RWMutex
}

func (r *rslice) writeSlice(i int) {
	r.l.Lock()
	r.s = append(r.s, i)
	r.l.Unlock()
}

func (r *rslice) readSlice() {
	r.l.RLock()
	s := r.s
	r.l.RUnlock()

	for _, i := range s {
		fmt.Printf("%d\n", i)
	}
}

func main() {
	r := rslice{s: make([]int, 0)}

	go func() {
		for i := 0; i < 10; i++ {
			r.writeSlice(i)
		}
	}()

	r.readSlice()
}
