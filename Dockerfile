FROM ubuntu as builder

RUN mkdir /go
ENV GOPATH /go

RUN apt-get update && apt-get install -y git golang build-essential libtool autopoint autoconf
RUN go get -v github.com/cretz/tor-static
RUN cd /go/src/github.com/cretz/tor-static && go run build.go build-all

# RUN go get github.com/sledigabel/gotorproxy
# RUN mkdir /code && git clone https://github.com/sledigabel/gotorproxy.git /code/gotorproxy
RUN mkdir -p /code/gotorproxy
ADD . /code/gotorproxy
RUN go get \
    github.com/cretz/bine/process/embedded \
    github.com/elazarl/goproxy \
    golang.org/x/net/proxy \
    github.com/FiloSottile/mkcert

RUN cd /code/gotorproxy && go build -x -v -o /go/bin/gotorproxy .


FROM ubuntu

RUN apt-get update && apt-get install -y ca-certificates openssl
COPY --from=builder /go/bin/gotorproxy /
COPY --from=builder /go/bin/mkcert /

EXPOSE 8081
ENV DOMAIN mydomain.org

RUN mkdir /ca

ADD start.sh /
CMD [ "/start.sh" ]
