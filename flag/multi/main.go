package main

import (
	"flag"
	"fmt"
	"strings"
)

type multi []string

func (m *multi) String() string {
	return strings.Join(*m, "; ")
}

func (m *multi) Set(value string) error {
	*m = append(*m, value)
	return nil
}

var test multi

func main() {
	flag.Var(&test, "test", "Test value")
	flag.Parse()
	fmt.Printf("%s\n", &test)
}
