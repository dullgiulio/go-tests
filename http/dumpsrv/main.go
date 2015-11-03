package main

import (
	"flag"
	//"fmt"
	"io"
	"log"
	"net/http"
	//"net/http/httputil"
	"os"
	"strconv"
)

type handler struct {
	w io.Writer
}

func (h *handler) handle(w http.ResponseWriter, r *http.Request) {
	/*
    out, err := httputil.DumpRequestOut(r, true)
	if err != nil {
		log.Fatal("dump: ", err)
	}
	fmt.Fprintf(h.w, "%s\n", out)
	*/
    r.Write(h.w)
    w.Write([]byte("All OK!"))
}

func main() {
	port := flag.Int("port", 8080, "Listen on port")
	flag.Parse()

	h := &handler{os.Stdout}
	http.HandleFunc("/", h.handle)
	log.Println("Starting http server on port:", *port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
