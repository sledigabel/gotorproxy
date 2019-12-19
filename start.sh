#!/bin/bash

# ln /go/bin/gotorproxy /
/gotorproxy &
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
