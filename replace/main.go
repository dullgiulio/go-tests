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
	cmppos    int
	buf, rbuf bytes.Buffer
}

// Reset the processor
func (m *Processor) Reset(src io.Reader) {
	m.Src = src
	m.buf.Reset()
}

func (m *Processor) Read(p []byte) (int, error) {
	if m.ChunkSize == 0 {
		m.ChunkSize = DefaultChunkSize
	}

	if m.cmppos == 0 {
		n, err := io.CopyN(&m.buf, m.Src, int64(m.ChunkSize))
		if err != nil && err != io.EOF {
			return n, err
		}
	}

	for i := 0; i < n; i++ {
		if m.Old[m.cmppos] != m.buf[i] {
			m.cmppos = 0
			continue
		}
		m.cmppos++
		if m.cmppos > len(m.Old) {
			break
		}
	}

	if m.cmppos > 0 {
		// Can write everything before the part to replace
		n = write(m.buf[:len(m.buf)-m.cmppos])
		m.buf = m.buf[m.cmppos:]

		if m.cmpppos > len(m.Old) {
			m.cmppos = 0
			m = write(m.New)
			n += m
			m.buf = m.buf[len(m.New):]
		}
	}

	return write(m.buf)
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
