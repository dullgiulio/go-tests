package main

import (
	"fmt"
	"log"
)

func main() {
	s := "[arg0, arg0.1], (arg1, (arg2.1, arg2.2),  arg3, arg4, (arg5.1,arg5.2,  arg5.3 ) )"
	p := newParser(s)
	if err := p.parse(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", s)
	printEntries(p.current, 0)
}
