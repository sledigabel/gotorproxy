package main

import (
	"bytes"
	"context"
	"encoding/pem"
	"log"
	"net/http"
	"os"

	"github.com/elazarl/goproxy"
)

type ProxyRunner struct {
	addr      string
	log       *log.Logger
	Verbose   bool
	proxy     *goproxy.ProxyHttpServer
	torRunner *TorRunner
	shutdown  chan struct{}
	Started   bool
}

func NewProxyRunner() *ProxyRunner {
	p := &ProxyRunner{}
	p.log = log.New(os.Stdout, "[Proxy] ", 3)
	p.proxy = goproxy.NewProxyHttpServer()
	p.proxy.Logger = p.log
	p.torRunner = NewTorRunner()
	p.torRunner.Verbose = false
	p.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	p.proxy.OnRequest(goproxy.DstHostIs("proxycert")).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			ctx.Warnf("Received req for certificate from: %v (%v)", r.RemoteAddr, r.UserAgent)
			resp := goproxy.NewResponse(r,
				goproxy.ContentTypeHtml, http.StatusMovedPermanently,
				"Redirecting to Certificate")
			resp.Header.Set("Location", "http://secret/cacert.pem")
			return r, resp
		})
	p.proxy.OnRequest(goproxy.UrlIs("secret/cacert.pem")).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			ctx.Warnf("Sending out the Certificate Authority to: %v (%v)", r.RemoteAddr, r.UserAgent)
			buf := new(bytes.Buffer)
			for _, c := range goproxy.GoproxyCa.Certificate {
				err := pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: c})
				if err != nil {
					p.log.Printf("Couldn't read certificate: %v", err)
				}
			}
			resp := goproxy.NewResponse(r,
				goproxy.ContentTypeText, http.StatusOK,
				string(buf.String()))
			resp.Header.Add("content-disposition", "attachment")
			return r, resp
		})
	p.proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			// let's send it via Tor
			resp, err := p.torRunner.HandleRequest(r)
			if err != nil {
				p.log.Printf("ERROR - Failed to connect to %v: %v\n", r.RequestURI, err)
				return r, nil
			}
			if p.Verbose {
				p.log.Printf("%s %s %s [%s] %d %d\n", r.RemoteAddr, r.Method, r.URL, r.Header, resp.ContentLength, resp.StatusCode)
			}
			return r, resp
		})

	p.shutdown = make(chan struct{})
	return p
}

func (p *ProxyRunner) ProxyStart() {

	p.torRunner.Verbose = p.Verbose
	p.Started = true
	p.log.Println("Starting TOR")
	go p.torRunner.TorStart()
	p.torRunner.WaitTillReady()

	srv := http.Server{Addr: p.addr, Handler: p.proxy, ErrorLog: p.log}
	go func() {
		p.log.Println("Starting proxy")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			p.log.Fatalf("ListenAndServe(): %v", err)
		}
		p.log.Println("Exiting...")
	}()
	p.log.Println("test")
	<-p.shutdown
	p.log.Println("Shutdown requested")
	ctx := context.TODO()
	if err := srv.Shutdown(ctx); err != nil {
		p.log.Panicf("Error while stopping the proxy: %v", err)
	}
	p.torRunner.shutdown <- struct{}{}
	// wait till tor shuts down.
	for p.torRunner.Started {
	}
	p.Started = false
	p.log.Println("Exiting proxy")
}
