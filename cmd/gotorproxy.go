package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/elazarl/goproxy"
)

func main() {

	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8080", "proxy listen address")
	flag.Parse()
	log.SetPrefix("[main] ")
	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = log.New(os.Stdout, "[proxy] ", log.Flags())
	proxy.Verbose = *verbose

	log.Println("Starting new Proxy service")
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
