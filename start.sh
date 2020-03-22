#!/bin/bash

RENEWAL=${RENEWAL_PERIOD:-14400} # 4 hours

# if the certs aren't found we'll generte them
if [ ! -f "/ca/cacert.pem" ]
then
    echo "Generating CA Cert for Tor Proxy"
    CAROOT=/ca /mkcert -install
    mv /ca/rootCA.pem /ca/cacert.pem
    mv /ca/rootCA-key.pem /ca/cakey.pem
fi


function terminate_properly()
{
    # because of CGO, the process needs to receive TERM twice
    # to be handled properly
    echo "Trap called"
    kill -TERM ${PID}
    sleep .5
    ps -fp ${PID} >/dev/null && kill -TERM ${PID}
    wait ${PID}
}

function terminate_for_real()
{
    terminate_properly
    exit
}

trap terminate_for_real EXIT INT

while [ 1 ]
do
    /gotorproxy -cacert /ca/cacert.pem -cakey /ca/cakey.pem -addr :8081 &
    PID=$!
    echo "Process started with PID: ${PID}"
    sleep ${RENEWAL}
    terminate_properly
done

