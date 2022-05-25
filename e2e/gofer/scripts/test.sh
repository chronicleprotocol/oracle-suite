#!/usr/bin/env bash

start_rpcsplitter() {
    timeout=2
    endpoints="$SMOCKER,$SMOCKER,http://10.255.255.1"
    echo "rpc-splitter running with $endpoints endpoints + timeout of $timeout seconds."
    (rpc-splitter run --log.format json --eth-rpc $endpoints -v debug -t $timeout)&
    sleep 3
}

start_rpcsplitter

printenv
go test -v -parallel 1 -cpu 1 ./
