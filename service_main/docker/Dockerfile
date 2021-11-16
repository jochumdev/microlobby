# STEP 1 build executable binary
FROM golang:1.17-alpine as builder

ARG PACKAGE

# Install SSL ca certificates
RUN apk add --update git ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

# Copy the context into /microlobby
COPY . /microlobby
WORKDIR /microlobby/${PACKAGE}

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix cgo -ldflags="-w -s" -o /usr/local/bin/${PACKAGE}


# STEP 2 build a small image
# start from scratch
FROM busybox

ARG PACKAGE

LABEL maintainer="René Jochum <rene@jochum.dev>"

# Copy certs, passwd and binary from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /usr/local/bin/${PACKAGE} /usr/local/bin/

# Run as appuser
USER appuser

# Create the ENV var PACKAGE and run $PACKAGE
ENV PACKAGE=${PACKAGE}
CMD [ `/usr/local/bin/$PACKAGE` ]