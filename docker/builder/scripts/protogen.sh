#!/bin/sh
set -ex

# Build shared
cd /microlobby/shared
if test -d proto; then
    cd proto
    for d in *; do
        if test -d $d; then
            cd /microlobby/shared/proto/$d
            protoc -I/microlobby --proto_path=$GOPATH/bin:. --micro_out=paths=source_relative:. --go_out=paths=source_relative:. $d.proto
            cd /microlobby/shared/proto
        fi
    done
fi

# Build all given services
for svc in $1; do
    cd /microlobby/service/${svc}
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get -installsuffix cgo -ldflags="-w -s" ./...
    if test -d proto; then 
        cd proto
        for d in *; do
            if test -d $d; then
                cd /microlobby/service/${svc}/proto/$d
                protoc -I/microlobby --proto_path=$GOPATH/bin:. --micro_out=paths=source_relative:. --go_out=paths=source_relative:. $d.proto
                cd /microlobby/service/${svc}/proto
            fi
        done
    fi
done