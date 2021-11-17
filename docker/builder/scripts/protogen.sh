#!/bin/sh
set -ex

for svc in $1; do
    cd /microlobby/service/${svc}
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get -installsuffix cgo -ldflags="-w -s" ./...
    if test -d proto; then 
        cd proto
        protoc -I$(shell go list -f '{{ .Dir }}' -m wz2100.net/microlobby/shared)/../ --proto_path=$$GOPATH/bin:. --micro_out=paths=source_relative:. --go_out=paths=source_relative:. $@.proto
    fi
done