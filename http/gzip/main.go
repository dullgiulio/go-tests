package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/klauspost/compress/gzip"
)

type deferredWriter struct {
	io.Writer
}

func newDeferredWriter() *deferredWriter {
	return &deferredWriter{ioutil.Discard}
}

func (d *deferredWriter) setWriter(w io.Writer) {
	d.Writer = w
}

type deferredGzip struct {
	*gzip.Writer
	w *deferredWriter
}

func newDeferredGzip(w *deferredWriter) *deferredGzip {
	gw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
	return &deferredGzip{gw, w}
}

func (dg *deferredGzip) setWriter(w io.Writer) {
	dg.w.Writer = w
}

func (dg *deferredGzip) reset() {
	dw := newDeferredWriter()
	dg.Reset(dw)
	dg.w = dw
}

type handler struct {
	zips chan *deferredGzip
	used chan *deferredGzip
}

func newHandler(n int) *handler {
	h := &handler{
		zips: make(chan *deferredGzip, n),
		used: make(chan *deferredGzip, n),
	}
	for i := 0; i < n; i++ {
		h.zips <- newDeferredGzip(newDeferredWriter())
	}
	return h
}

func (h *handler) httpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)

	dg := <-h.zips
	dg.setWriter(w)
	dg.Write(data)
	dg.Close()
	h.used <- dg
}

func (h *handler) recycle() {
	for dg := range h.used {
		dg.reset()
		h.zips <- dg
	}
}

func main() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	h := newHandler(n)
	http.HandleFunc("/", h.httpHandler)
	go h.recycle()

	if *cpuprofile != "" {
		go http.ListenAndServe(":8080", nil)
		<-time.After(3 * time.Second)
	} else {
		http.ListenAndServe(":8080", nil)
	}
}
