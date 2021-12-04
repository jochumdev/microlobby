# STEP 1 build executable binary
FROM golang:1.17-alpine as builder

ARG SERVICE
ARG EXECUTABLE

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

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix cgo -ldflags="-w -s" -o /usr/local/bin/${EXECUTABLE}


# STEP 2 build a small image
# start from scratch
FROM busybox

ARG SERVICE
ARG EXECUTABLE

LABEL maintainer="René Jochum <rene@jochum.dev>"

# Copy certs, passwd and binary from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /usr/local/bin/${EXECUTABLE} /usr/local/bin/

# Run as appuser
USER appuser

# For future use
ENV SERVICE=${SERVICE}

# Create the ENV var EXECUTABLE and run $EXECUTABLE
ENV EXECUTABLE=${EXECUTABLE}
CMD [ `/usr/local/bin/$EXECUTABLE` ]