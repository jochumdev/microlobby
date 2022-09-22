ARG DOCKER_IO=docker.io
ARG TAG=14.5-alpine

FROM ${DOCKER_IO}/library/postgres:${TAG}
COPY ./docker-entrypoint-initdb.d/* /docker-entrypoint-initdb.d/