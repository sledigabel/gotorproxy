package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/elazarl/goproxy"
)

type ProxyRunner struct {
	addr      string
	log       *log.Logger
	Verbose   bool
	proxy     *goproxy.ProxyHttpServer
	torRunner *TorRunner
	shutdown  chan struct{}
}

func NewProxyRunner() *ProxyRunner {
	p := &ProxyRunner{}
	p.log = log.New(os.Stdout, "[Proxy] ", 3)
	p.proxy = goproxy.NewProxyHttpServer()
	p.proxy.Logger = p.log
	p.torRunner = NewTorRunner()
	p.torRunner.Verbose = false
	p.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	p.proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			if p.Verbose {
				p.log.Printf("Received: %v", r)
			}
			// let's send it via Tor
			resp, err := p.torRunner.HandleRequest(r)
			if err != nil {
				p.log.Printf("ERROR - Failed to connect to %v: %v\n", r.RequestURI, err)
				return r, nil
			}
			if p.Verbose {
				p.log.Printf("%s %s %s %d %d\n", r.RemoteAddr, r.Method, r.URL, resp.ContentLength, resp.StatusCode)
			}
			return r, resp
		})
	p.shutdown = make(chan struct{})
	return p
}

func (p *ProxyRunner) ProxyStart() {

	p.torRunner.Verbose = p.Verbose

	p.log.Println("Starting TOR")
	go p.torRunner.TorStart()
	p.torRunner.WaitTillReady()

	p.log.Println("Starting proxy")
	srv := http.Server{Addr: ":8080", Handler: p.proxy, ErrorLog: p.log}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			p.log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	<-p.shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		p.log.Panicf("Error while stopping the proxy: %v", err)
	}
	p.torRunner.shutdown <- struct{}{}

}
