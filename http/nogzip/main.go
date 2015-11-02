package main

import (
	//"fmt"
	"net/http"
	//"net/http/httputil"
    "log"
)

func main() {
    reqUrl := "http://localhost:8080/"
    tr := &http.Transport{
	//	DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
    req, err := http.NewRequest(
		"GET",
		reqUrl,
        nil,
    )
	if err != nil {
		log.Fatal("newreq: ", err)
	}
    //req.Header.Set("Accept-Encoding", "identity")
    _, err = client.Do(req)
    if err != nil {
		log.Fatal("req: ", err)
	}
	/*
    out, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal("dump: ", err)
	}
	fmt.Printf("%s\n", out)
    */
}
