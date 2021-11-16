SUBDIRS=shared service_auth service_lobby service_main

.PHONY: all
all:


.PHONY: update-modules
update-modules:
	for dir in ${SUBDIRS}; do ${MAKE} -C $$dir $@ || exit 1; done