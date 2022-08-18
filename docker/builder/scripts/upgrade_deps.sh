#!/bin/sh
set -ex

cd /microlobby;
go mod tidy -go=1.18

for svc in $1; do
    cd /microlobby/service/${svc}
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get -installsuffix cgo -ldflags="-w -s" -u ./...
done