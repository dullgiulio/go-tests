package sima

import (
	"fmt"
	"io"
)

type HeaderWriter struct {
	w      io.Writer
	prefix string
}

func NewHeaderWriter(w io.Writer) *HeaderWriter {
	return &HeaderWriter{prefix: "sima", w: w}
}

func (h *HeaderWriter) Put(key, val string) {
	// TODO: Validate key format (no spaces etc)
	fmt.Fprintf(h.w, "%s: %s: %s\n", h.prefix, key, val)
}

type HeaderReader struct {
	r      io.Reader
	prefix string
}

func NewHeaderReader(r io.Reader) *HeaderReader {
	return &HeaderReader{prefix: "sima", r: r}
}

func (g *HeaderReader) Get(key string) string {
	return ""
}
