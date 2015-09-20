package main

import (
	"fmt"
	"unsafe"
)

func main() {
	var x int
	fmt.Println("%s\n", unsafe.Sizeof(x))
}
