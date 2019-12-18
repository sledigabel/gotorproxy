FROM ubuntu

RUN mkdir /go
ENV GOPATH /go

RUN apt-get update && apt-get install -y git golang build-essential libtool autopoint autoconf
RUN go get github.com/cretz/tor-static
RUN cd /go/src/github.com/cretz/tor-static && go run build.go build-all

RUN go get github.com/sledigabel/gotorproxy

