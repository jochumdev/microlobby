# STEP 1 build executable binary
FROM golang:1.18-alpine as builder

ARG SERVICE

# Install SSL ca certificates
RUN apk add --update git ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

# Copy shared/ and $PACKAGE/ to /microlobby - this speeds up builds and we build less
COPY ./go.mod /microlobby/go.mod
COPY ./go.sum /microlobby/go.sum
COPY ./shared /microlobby/shared
COPY ./service/${SERVICE} /microlobby/service/${SERVICE}

WORKDIR /microlobby/service/${SERVICE}

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix cgo -ldflags="-w -s" -o /usr/local/bin/microservice


# STEP 2 build a small image
# start from scratch
FROM busybox

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