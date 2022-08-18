FROM golang:1.18

# Install packages
RUN apt-get update && \
    apt-get --no-install-recommends --no-install-suggests --yes --quiet install \
        git-core ca-certificates make wget unzip && \
    cd /tmp && wget https://github.com/protocolbuffers/protobuf/releases/download/v21.5/protoc-21.5-linux-x86_64.zip && \
    unzip protoc-21.5-linux-x86_64.zip -d /usr/local/ && chmod +x /usr/local/bin/protoc


# Create the Volumes
RUN mkdir /microlobby && chown -R root: /microlobby && cd /microlobby

COPY ./scripts /scripts

# Run as user from here
WORKDIR /microlobby

VOLUME [ "/go", "/microlobby" ]