# gotorproxy

A Simple web proxy to TOR.

## Usage

TBD

## Compiling and running

`gotorproxy` uses [cretz`s Embedded Tor](https://github.com/cretz/bine) and requires you to follow [those steps](https://github.com/cretz/tor-static) to compile `tor-static`. It uses CGO to embedd tor into the gotorproxy binary.

Once compiled, you can download and compile gotorproxy:
```
go get github.com/sledigabel/gotorproxy
```

## Deps

- github.com/elazarl/goproxy
  Very advaced web proxying for go. Powerful features.

- github.com/cretz/bine
  ToR interfacing with Go. Provides a nice transport layer and does all the go heavylifting.
  
# TODO

- Custom Check for running proxy
  
