SUBDIRS=shared service

.PHONY: all
all: builders protoc build run
#	for dir in ${SUBDIRS}; do ${MAKE} -C $$dir $@ || exit 1; done

.PHONY: builders
builders:
	docker-compose --profile tools build

.PHONY: protoc
protoc:
	docker-compose --profile tools run --rm builder make _protoc

.PHONY: build
build:
	docker-compose --profile app build

.PHONY: run
run:
	docker-compose --profile app up -d

.PHONY: down
down:
	docker-compose --profile app down

.PHONY: goupdate
goupdate:
	docker-compose --profile tools run --rm builder make _goupdate

.PHONY: _goupdate
_goupdate:
	for dir in ${SUBDIRS}; do ${MAKE} -C $$dir $@ || exit 1; done

.PHONY: _download
_download:
	go install github.com/asim/go-micro/cmd/protoc-gen-micro/v4@v4.0.0
	for dir in ${SUBDIRS}; do ${MAKE} -C $$dir $@ || exit 1; done

.PHONY: _protoc
_protoc: _download
	for dir in ${SUBDIRS}; do ${MAKE} -C $$dir $@ || exit 1; done
