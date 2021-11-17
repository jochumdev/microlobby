FROM golang:1.17

# Install packages
RUN apt-get update && \
    apt-get --no-install-recommends --no-install-suggests --yes --quiet install \
        git-core ca-certificates make protobuf-compiler protoc-gen-go libprotobuf-dev

# Create the Volumes
RUN mkdir /microlobby && chown -R root: /microlobby && cd /microlobby

COPY ./scripts /scripts

# Run as user from here
WORKDIR /microlobby

VOLUME [ "/go", "/microlobby" ]