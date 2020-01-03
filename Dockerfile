FROM ubuntu as builder

RUN mkdir /go
ENV GOPATH /go

RUN apt-get update && apt-get install -y git golang build-essential libtool autopoint autoconf
RUN go get -v github.com/cretz/tor-static
RUN cd /go/src/github.com/cretz/tor-static && go run build.go build-all

# RUN go get github.com/sledigabel/gotorproxy
# RUN mkdir /code && git clone https://github.com/sledigabel/gotorproxy.git /code/gotorproxy
RUN mkdir -p /code/gotorproxy
RUN go get \
    github.com/cretz/bine/process/embedded \
    github.com/elazarl/goproxy \
    golang.org/x/net/proxy \
    github.com/FiloSottile/mkcert

ADD . /code/gotorproxy
RUN cd /code/gotorproxy && go build -x -v -o /go/bin/gotorproxy .


FROM ubuntu

RUN apt-get update && apt-get install -y ca-certificates openssl curl
COPY --from=builder /go/bin/gotorproxy /
COPY --from=builder /go/bin/mkcert /

EXPOSE 8081
VOLUME /ca

ADD check.sh /
HEALTHCHECK --start-period=1m --retries=1 --interval=2m --timeout=1m CMD /check.sh 
ADD start.sh /
CMD [ "/start.sh" ]
