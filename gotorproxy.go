package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/elazarl/goproxy"
)

type Flags struct {
	Verbose    bool
	Address    string
	CACertPath string
	CAKeyPath  string
}

func main() {

	var flags Flags
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8080", "proxy listen address")
	cacert := flag.String("cacert", "", "CA Certiicate path")
	cakey := flag.String("cakey", "", "CA Key path")
	flag.Parse()

	flags.Verbose = *verbose
	flags.Address = *addr
	flags.CACertPath = *cacert
	flags.CAKeyPath = *cakey
	log.SetPrefix("[main] ")

	p := NewProxyRunner()
	p.addr = flags.Address
	p.Verbose = flags.Verbose

	done := make(chan struct{}, 1)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		select {
		case sig := <-c:
			log.Printf("Caught signal: %v", sig)
			p.shutdown <- struct{}{}
			// wait till proper shutdown
			for p.Started {
			}
			log.Println("Shutdown completed. Exiting...")
			done <- struct{}{}
		}
	}()

	// if provided custom CA
	if len(flags.CACertPath) > 0 && len(flags.CAKeyPath) > 0 {
		caCert, err := ioutil.ReadFile(flags.CACertPath)
		if err != nil {
			log.Fatalf("Error while opening CA cert %v: %v", flags.CACertPath, err)
		}
		caKey, err := ioutil.ReadFile(flags.CAKeyPath)
		if err != nil {
			log.Fatalf("Error while opening CA key %v: %v", flags.CAKeyPath, err)
		}
		goproxyCa, err := tls.X509KeyPair(caCert, caKey)
		if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
			log.Fatalf("Error while reading the certificate: %v", err)
		}

		goproxy.GoproxyCa = goproxyCa
		goproxy.GoproxyCa = goproxyCa
		goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
		goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
		goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
		goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	}

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
	log.Printf("ToR description: \n%v\n", newStr)

	<-done
	log.Println("Done")
}
