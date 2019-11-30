package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {

	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8080", "proxy listen address")
	flag.Parse()
	log.SetPrefix("[main] ")

	run(*verbose, *addr)

}

func run(verbose bool, addr string) {
	p := NewProxyRunner()
	p.addr = addr
	p.Verbose = verbose

	go p.ProxyStart()

	p.torRunner.WaitTillReady()
	r, err := http.NewRequest(http.MethodGet, "https://wtfismyip.com/json", nil)
	if err != nil {
		log.Printf("Error with req: %v", err)
	}

	resp, err := p.torRunner.HandleRequest(r)
	if err != nil {
		log.Printf("Error with resp: %v", err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	newStr := buf.String()
	log.Println(newStr)

	time.Sleep(2 * time.Minute)
	p.shutdown <- struct{}{}

}
