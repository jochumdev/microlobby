# MicroLobby - 3rd gen lobbyserver for Warzone 2100

MicroLobby is the next, next gen lobbyserver for Warzone 2100 after [wzlobbserver-ng](https://github.com/Warzone2100/wzlobbyserver-ng).

## Basic Architecture

It's written in Golang by using [go-micro.dev/v4](https://go-micro.dev) for simplicity. Transport, Registry and Broker is done over NATS.

The draw.io flowchart for the Architecture:
![Micro Service Architecture](/docs/micro-service-architecture.png)

## Services

### http_proxy

A very simple Proxy to MicroServices. They have to register routes with it:

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

- [Task](https://taskfile.dev/#/installation)
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

Now enjoy the [health api](http://localhost:8080/health)

### Taskfile

```bash
task -l
```

```
task: Available tasks for this project:
* build:                Build all containers
* check:toolchain:      Check if you have all tools installed
* default:              Build and run microlobby
* down:                 Stopp all containers
* download:             Download go dependencies
* generate:protoc:      Generate protobuf files
* run:                  Run all containers
* upgrade:deps:         Update all go dependencies
```

## Authors

René Jochum - rene@jochum.dev

## License

Its dual licensed:

- Apache-2.0
- GPL-2.0-or-later
