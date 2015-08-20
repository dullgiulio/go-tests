package main

import (
	"bufio"
	"fmt"
	"os"
)

func sumBytes(bs []byte) int {
	var t int
	for _, b := range bs {
		t += int(b)
	}
	return t
}

func main() {
	var lines int
	div := 3
	scanner := bufio.NewScanner(os.Stdin)
	results := make(map[int]int)
	for scanner.Scan() {
		d := sumBytes(scanner.Bytes())
		results[d%div]++
		lines++
	}
	for val, hits := range results {
		fmt.Printf("%d: %d (%d%%)\n", val, hits, hits*100/lines)
	}
}
