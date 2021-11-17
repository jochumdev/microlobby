#!/bin/sh
set -ex

cd /microlobby;
go mod tidy -go=1.16 && go mod tidy -go=1.17

for svc in $1; do
    cd /microlobby/service/${svc}
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get -installsuffix cgo -ldflags="-w -s" -u ./...
done