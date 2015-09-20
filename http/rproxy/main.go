package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var httpDirector func(req *http.Request)
var (
	debug = flag.Bool("debug", false, "Dump requests to stdout")
	host  = flag.String("host", "", "Host to proxy")
	local = flag.String("local", "", "Local IP:Port to listen to")
)

type remoteHost url.URL

func (r *remoteHost) director(req *http.Request) {
	req.Host = r.Host
	httpDirector(req)
	if *debug {
		buf, err := httputil.DumpRequestOut(req, false)
		if err == nil {
			_, err = os.Stdout.Write(buf)
		}
		if err != nil {
			log.Print(err)
		}
	}
}

func main() {
	flag.Parse()
	if *host == "" {
		log.Fatal("Please specify a host with -host")
	}
	if *local == "" {
		log.Fatal("Please specify a local host:port with -local")
	}
	target := url.URL{
		Scheme: "http",
		Host:   *host,
	}
	rh := remoteHost(target)
	rp := httputil.NewSingleHostReverseProxy(&target)
	httpDirector = rp.Director
	rp.Director = rh.director
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rp.ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(*local, nil))
}
