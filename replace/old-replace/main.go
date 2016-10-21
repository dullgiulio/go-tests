package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"unicode/utf8"
)

const DefaultChunkSize = 255

// Processor implements an io.Reader
type Processor struct {
	Src       io.Reader
	Old, New  []byte
	ChunkSize int
	buf, rbuf bytes.Buffer
}

// Reset the processor
func (m *Processor) Reset(src io.Reader) {
	m.Src = src
	m.buf.Reset()
}

// Replace reimplements bytes.Replace in a way that can reuse the buffer
func Replace(s, old, new []byte, n int, buf *bytes.Buffer) error {
	m := 0

	if n != 0 {
		// Compute number of replacements.
		m = bytes.Count(s, old)
	}

	if buf == nil {
		buf = &bytes.Buffer{}
	}

	buf.Reset()

	if m == 0 {
		// Just return a copy.
		_, err := buf.Write(s)
		return err
	}

	if n < 0 || m < n {
		n = m
	}

	// Apply replacements to buffer.
	buf.Grow(len(s) + n*(len(new)-len(old)))

	start := 0
	for i := 0; i < n; i++ {
		j := start

		if len(old) == 0 {
			if i > 0 {
				_, wid := utf8.DecodeRune(s[start:])
				j += wid
			}
		} else {
			j += bytes.Index(s[start:], old)
		}

		if _, err := buf.Write(s[start:j]); err != nil {
			return err
		}

		if _, err := buf.Write(new); err != nil {
			return err
		}

		start = j + len(old)
	}

	_, err := buf.Write(s[start:])
	return err
}

func (m *Processor) Read(p []byte) (int, error) {
	if m.ChunkSize == 0 {
		m.ChunkSize = DefaultChunkSize
	}

	// first flush any buffered data to p
	n, err := m.buf.Read(p)
	if n == len(p) || (err != nil && err != io.EOF) {
		return n, err
	}

	// m.buf must be empty now
	_, err = io.CopyN(&m.buf, m.Src, int64(m.ChunkSize))
	if err != nil && err != io.EOF {
		return n, err
	}

	rerr := Replace(m.buf.Bytes(), m.Old, m.New, -1, &m.rbuf)
	m.buf.Reset()

	if rerr != nil {
		return n, rerr
	}

	copied, rerr := m.rbuf.Read(p[n:])
	n += copied
	if rerr != nil && rerr != io.EOF {
		return n, rerr
	}

	// copy anything not put in p, back to the buffer
	if _, rerr := m.buf.ReadFrom(&m.rbuf); rerr != nil {
		return n, rerr
	}

	return n, err
}

func main() {
	p := &Processor{
		Src:       strings.NewReader("hello hold the door giulio"),
		ChunkSize: 10,
		Old:       []byte("hold the door"),
		New:       []byte("hodor"),
	}
	if _, err := io.Copy(os.Stdout, p); err != nil {
		log.Fatal("copy: ", err)
	}
}
