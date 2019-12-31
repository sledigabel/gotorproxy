#!/bin/bash

set -u -e -o pipefail

CHECK_URL=${URL:-"http://thehub7xbw4dc5r2.onion/"}

curl \
    --head \
    -x localhost:8081 \
    --retry 3 \
    --retry-delay 1 \
    --retry-max-time 120 \
    ${CHECK_URL} \
    || exit 1