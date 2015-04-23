package sima

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

type headerWriter struct {
	w      io.Writer
	prefix string
}

func newHeaderWriter(w io.Writer) *headerWriter {
	return &headerWriter{prefix: "sima", w: w}
}

func (h *headerWriter) put(key, val string) {
	// TODO: Validate key format (no spaces etc)
	fmt.Fprintf(h.w, "%s: %s: %s\n", h.prefix, key, val)
}

type headerReader struct {
	prefix string
	values map[string]string
}

func newHeaderReader() *headerReader {
	return &headerReader{prefix: "sima", values: make(map[string]string)}
}

func (g *headerReader) readAll(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}

		log.Print("read: ", line)

		if line[0:len(g.prefix)] == g.prefix {
			line = line[len(g.prefix)+2:]
			end := strings.IndexByte(line, ':')
			if end < 0 {
				continue
			}

			g.values[line[0:end]] = line[end+2:]
		}
	}
}

func (g *headerReader) get(key string) string {
	v, _ := g.values[key]
	return v
}
