#!/bin/bash

# if the certs aren't found we'll generte them
if [ ! -f "/ca/cacert.pem" ]
then
    echo "Generating CA Cert for Tor Proxy"
    /mkcert -cert-file /ca/cacert.pem -key-file /ca/cakey.pem ${DOMAIN:-mydomain.org}
fi

/gotorproxy -cacert /ca/cacert.pem -cakey /ca/cakey.pem -addr :8081 &
PID=$!

echo "Process started with PID: ${PID}"

function terminate_properly()
{
    # because of CGO, the process needs to receive TERM twice
    # to be handled properly
    echo "Trap called"
    kill -TERM ${PID}
    sleep .5
    ps -fp ${PID} >/dev/null && kill -TERM ${PID}
}

trap terminate_properly EXIT INT

wait ${PID}
