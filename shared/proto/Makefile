SUBDIRS := $(wildcard *)
current_dir = $(shell pwd)

.PHONY: _protoc
_protoc: $(SUBDIRS)

.PHONY: _download
_download:
	go mod download

.PHONY: $(SUBDIRS)
$(SUBDIRS):
	test -d $@ && cd $@ && protoc -I$(shell go list -f '{{ .Dir }}' -m wz2100.net/microlobby/shared)/../ --proto_path=$$GOPATH/bin:. --micro_out=paths=source_relative:. --go_out=paths=source_relative:. $@.proto; cd ${current_dir}