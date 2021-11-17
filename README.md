# MicroLobby - 3rd gen lobbyserver for Warzone 2100

MicroLobby is the next, next gen lobbyserver for Warzone 2100 after [wzlobbserver-ng](https://github.com/Warzone2100/wzlobbyserver-ng).

## Basic Architecture

It's written in Golang by using [go-micro.dev/v4](https://go-micro.dev) for simplicity Transport, Registry and Broker is done over NATS.

See this draw.io draw for the Architecture:
![Micro Service Architecture](/docs/micro-service-architecture.png)

## Services

### http_proxy

A very simple Proxy to MicroServices which register routes with it:

- Auth Service
- Lobby Service
- Settings Service

### Settings Service

Basic Key/Value Store with Permissions

### Auth Service

- Give Username + password and you get a JWT back.
- Internaly converts a JWT to a user with roles.

### Lobby Service

Register a game, get list of games and unregister it.

### EMail Service

Sends E-Mails for us.

### OAuth Service

Think it will never be implemented but be part of Profile Service which will be added later.

## Development

### Prerequesits

- GNU Make
- docker-compose 1.29+
- podman/docker

### Run

To run this you have to do the following steps:

```bash
git clone https://github.com/pcdummy/microlobby.git
cd microlobby
cp .env.sample .env
make
```

### Makefile

- **all**:
  - calls `docker-compose` to build the builder
  - builder generates `protobuf` code inside a container over `make _protoc`
  - calls `docker-compose` to build all services
  - starts all service containers
- **builders**: Builds builders
- **protoc**: Generates protobuf code with the help of a container
- **build**: Builds the code into containers
- **run**: Runs containers
- **down**: Stops them
- **goupdate**: Update all dependencies

## Authors

René Jochum - rene@jochum.dev

## License

Apache-2.0 License
