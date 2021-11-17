.PHONY: all
all:

.PHONY: _protoc
_protoc:

.PHONY: _download
_download:
	go mod download

.PHONY: _goupdate
_goupdate:
	go mod tidy -go=1.16 && go mod tidy -go=1.17
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get -installsuffix cgo -ldflags="-w -s" -u ./...