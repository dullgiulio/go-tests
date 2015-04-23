package sima

import (
	"fmt"
	"strings"
)

type header string

func (h header) output(key, val string) {
	fmt.Printf("%s: %s: %s\n", string(h), key, val)
}

func (h header) parse(line string) (key, val string) {
	if line == "" {
		return
	}

	if line[0:len(string(h))] != string(h) {
		return
	}

	line = line[len(string(h))+2:]
	end := strings.IndexByte(line, ':')
	if end < 0 {
		return
	}

	return line[0:end], line[end+2:]
}
