package main

import (
	"context"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/cretz/bine/process/embedded"
	"github.com/cretz/bine/tor"
)

type TorRunner struct {
	Verbose     bool
	torInstance *tor.Tor
	torConfig   *tor.StartConf
	log         *log.Logger
	dialer      *tor.Dialer
	httpClient  *http.Client
	cookieJar   *cookiejar.Jar
	Ready       bool
	Started     bool
	shutdown    chan struct{}
}

func NewTorRunner() *TorRunner {
	t := &TorRunner{}
	t.log = log.New(os.Stdout, "[TOR client] ", 3)
	t.torConfig = &tor.StartConf{ProcessCreator: embedded.NewCreator()}
	t.shutdown = make(chan struct{})
	t.Ready = false
	t.cookieJar, _ = cookiejar.New(&cookiejar.Options{})
	return t
}

func (t *TorRunner) TorStart() {
	if t.Verbose {
		t.torConfig.DebugWriter = os.Stdout
	} else {
		t.torConfig.ExtraArgs = []string{"--quiet"}
	}
	var err error
	t.log.Println("Starting TOR...")
	t.torInstance, err = tor.Start(nil, t.torConfig)
	t.torInstance.DeleteDataDirOnClose = true
	defer t.torInstance.Close()
	if err != nil {
		t.log.Panicf("Error while initialising TOR: %v", err)
	}
	dialCtx, dialCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer dialCancel()
	t.log.Println("Initiating Dialer")
	t.dialer, err = t.torInstance.Dialer(dialCtx, nil)
	if err != nil {
		t.log.Panicf("Error while dialing in: %v", err)
	}
	t.httpClient = &http.Client{
		Jar: t.cookieJar,
		Transport: &http.Transport{
			DialContext: t.dialer.DialContext,
		},
		Timeout: 5 * time.Minute,
	}

	t.Ready = true
	t.log.Println("Connected.")
	<-t.shutdown
	t.log.Println("Shutting down...")
	t.Ready = false
	t.torInstance.Close()
	t.Started = false
}

func (t *TorRunner) WaitTillReady() {
	for !t.Ready {
	}
}

func (t *TorRunner) HandleRequest(req *http.Request) (*http.Response, error) {
	if t.Verbose {
		t.log.Printf("%s %s %s %s\n", req.RemoteAddr, req.Method, req.URL, req.RequestURI)
	}
	// Forgetting the requestURI as this is set at receiving time.
	req.RequestURI = ""
	// send it to TOR
	resp, err := t.httpClient.Do(req)

	return resp, err
}
