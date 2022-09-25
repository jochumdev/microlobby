ARG DOCKER_IO=docker.io
ARG DOCKER_ORG_JO_MICRO=docker.io/jo-micro

# STEP 1 build executable binary
FROM ${DOCKER_ORG_JO_MICRO}/builder:latest AS builder

# Create appuser (/etc/passwd entry for the runner container)
RUN useradd appuser

COPY . /code
ARG SERVICE
WORKDIR /code/service/${SERVICE}

# RUN go mod download
ENV GOPATH="/go"
ARG CACHEBUST=1
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix cgo -ldflags="-w -s" -o /usr/local/bin/microservice


# STEP 2 build a small image
# start from busybox
FROM ${DOCKER_IO}/library/busybox:latest

ARG SERVICE

LABEL maintainer="René Jochum <rene@jochum.dev>"

# Copy certs, passwd and binary from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /usr/local/bin/microservice /usr/local/bin/microservice
RUN chmod +x /usr/local/bin/microservice

# Copy migrations
COPY ./service/${SERVICE}/migrations /migrations

# Run as appuser
USER appuser

# For future use
ENV SERVICE=${SERVICE}

CMD [ "/usr/local/bin/microservice" ]