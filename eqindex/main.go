package main

import (
	"fmt"
)

func main() {
	a := []int{-1, 3, -4, 5, 1, -6, 2, 1}
	total := 0
	acc := 0
	for i := 0; i < len(a); i++ {
		total += a[i]
	}
	for i := 0; i < len(a); i++ {
		total -= a[i]
		if total == acc {
			fmt.Printf("eq point at %d\n", i)
		}
		acc += a[i]
	}
	if total == acc {
		fmt.Printf("eq point at %d\n", len(a)-1)
	}
}
