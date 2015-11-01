package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type wrappedReader struct {
	reader    io.Reader
	pre, post *bytes.Reader
}

func newWrappedReader(r io.Reader, pre, post []byte) *wrappedReader {
	return &wrappedReader{
		reader: r,
		pre:    bytes.NewReader(pre),
		post:   bytes.NewReader(post),
	}
}

func (r *wrappedReader) Read(p []byte) (n int, err error) {
	if r.pre != nil {
		n, err = r.pre.Read(p)
		if err != io.EOF {
			return
		}
		r.pre = nil
	}
	if r.reader != nil {
		n, err = r.reader.Read(p)
		if err != io.EOF {
			return
		}
		r.reader = nil
	}
	if r.post != nil {
		n, err = r.post.Read(p)
		if err != io.EOF {
			return
		}
		r.post = nil
	}
	return
}

func main() {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "Hello, playground")

	wr := newWrappedReader(&buf, []byte("pre\n"), []byte("post\n"))

	io.Copy(os.Stdout, wr)
}
