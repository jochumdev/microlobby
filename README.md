# MicroLobby - 3rd gen lobbyserver for Warzone 2100

MicroLobby is the next, next gen lobbyserver for Warzone 2100 after [wzlobbserver-ng](https://github.com/Warzone2100/wzlobbyserver-ng).

## Why another lobby

- wzlobbyserver-ng is 10 years old, time for a new one
- Allows registering names
- Allows to verify users
- Adds a new REST Protocol
- Features are easy to add, like GameNetworkSockets, TOTP and others

## Technical features

- Requires only:
  - podman
  - docker-compose
  - [Task](https://taskfile.dev/#/installation)
- Everything in containers, leaves nothing on the Host except Podman/Docker volumes. "task rm" removes everything.
- Automated migrations, migrating on start
- gRPC+Protobuf internal, JSON/XML external
- Argon2-id Hashes
- JWT Tokens
- Integrated RBAC K/V store -> settings/v1
- Loosely coupled Microservices
- Fast to copy&paste a service, easy to start a new one
- Event System as example for IRC/Discord bots
- Registry and Broker over NATS
- Scale your db and everything else scales easy as it needs no Filesystem

## Basic Architecture

It's written in Golang by using [go-micro.dev/v4](https://go-micro.dev) for simplicity. Registry and Broker is done over NATS, Transport over gRPC.

The draw.io flowchart for the Architecture:
![Micro Service Architecture](/docs/micro-service-architecture.png)

## Services

### http_proxy

A very simple Proxy to MicroServices. They have to register routes with it over the help of infoservice.
Infoservice is just a ProtoBuf Service that needs to be registered into each microservice, see [auth/v1](https://github.com/pcdummy/microlobby/blob/master/service/auth/v1/main.go#L105).

It provides 3 routes, the result will be collected from all microservices:

| METHOD | Route             | AUTH | Description           |
| ------ | ----------------- | ---- | --------------------- |
| GET    | /health           |  n   | Summary health        |
| GET    | /proxy/v1/health  |  y   | Detailed health       |
| GET    | /proxy/v1/routes  |  y   | List of all routes    |

### settings/v1 Service

Basic Key/Value Store with Permissions

### auth/v1 Service

- Give Username + password and you get a JWT back.
- Internaly converts a JWT to a user with roles.
- Does not need a Database it uses AES for the JWT to enforce it's coming from this instance.
- [jwt.io](https://jwt.io) is very helpfull to see what's in a JWT.

### gamedb/v1 Service

Register a game, get list of games and unregister it.

It provides 4 routes:
| METHOD | Route             | AUTH | Description           |
| ------ | ----------------- | ---- | --------------------- |
| GET    | /                 |  y   | List games            |
| POST   | /                 |  y   | Create a new game     |
| PUT    | /:id              |  y   | Update a game         |
| DELETE | /:id              |  y   | Delete a game         |

## Development

### Prerequesits

- [Task](https://taskfile.dev/#/installation)
- podman
- docker-compose

Latest docker-compose (v2.7.0) works with podman >=4.1.1 only, for Debian testing I've used [Method 2: Ansible](https://computingforgeeks.com/how-to-install-podman-on-debian/) way to install the latest podman.

### Run

To run this you have to do the following steps.

```bash
git clone https://github.com/pcdummy/microlobby.git
cd microlobby
# To develop you don't need to change anything, for production you have to change all passwords
cp .env.sample .env
task
# Some containers don't start on first run, start them again
task up
```

Now enjoy the [health api](http://localhost:8080/health)

### Testing the API

- Get a token:

  It exports 3 variables:
  - MICROLOBBY
  - ACCESS_TOKEN
  - REFRESH_TOKEN

```bash
source ./token_login.sh http://localhost:8080 admin asdf1234
```

- Or refresh it:

```bash
source ./token_refresh.sh
```

- Check the proxy health api

```bash
curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS_TOKEN" $MICROLOBBY/proxy/v1/health | jq
```

- Get a list of routes

```bash
curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS_TOKEN" $MICROLOBBY/proxy/v1/routes | jq
```

- Create a game

```bash
curl -s -d @./docs/json-test/gamedb_v1_create.json -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS_TOKEN" $MICROLOBBY/gamedb/v1/ | jq
```

- List games

```bash
curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS_TOKEN" $MICROLOBBY/gamedb/v1/ | jq
```

### Remove everything or start from new

```bash
task rm
```

## Authors

- René Jochum - rene@jochum.dev
- Pastdue (ideas)

## License

Its dual licensed:

- Apache-2.0
- GPL-2.0-or-later
