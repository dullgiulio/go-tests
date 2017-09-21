package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	const n = 100
	var done sync.WaitGroup
	for i := 0; i < n; i++ {
		done.Add(1)
		go func(i int) {
			fmt.Printf("done %d\n", i)
			done.Done()
		}(i)
		time.Sleep(1)
	}
	done.Wait()
	fmt.Printf("Waited %v\n", n)
}
