#!/bin/sh
set -ex

# Build shared
cd /microlobby/shared
if test -d proto; then
    for proto in $(find proto/ -type f -name '*.proto' | xargs -0); do
        cd /microlobby/shared/$(dirname ${proto})
        protoc -I/microlobby --proto_path=$GOPATH/bin:. --micro_out=paths=source_relative:. --go_out=paths=source_relative:. $(basename ${proto})
    done
    cd /microlobby/shared
fi

# Build all given services
for svc in $1; do
    cd /microlobby/service/${svc}
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get -installsuffix cgo -ldflags="-w -s" ./...
    if test -d proto; then 
        for proto in $(find proto/ -type f -name '*.proto' | xargs -0); do
            cd cd /microlobby/service/${svc}/proto/$(dirname ${proto})
            protoc -I/microlobby --proto_path=$GOPATH/bin:. --micro_out=paths=source_relative:. --go_out=paths=source_relative:. $(basename ${proto})
        done
    fi
done

cd /microlobby
