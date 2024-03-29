# Microlobby API V1 Documentation

This document describes the V1 API of microlobby.

## General Notes

**SERVER NOTES:**

- Server must **only** accept connections via HTTPS.
- Server must limit requests, by default: **60 per hour**

**CLIENT NOTES:**

- To start, the client should **not** offer a "remember me" option that saves the authentication token. Why?
  - Because the authenticated token is equivalent to the user's password, and storing it securely is problematic (depending on platform):
    - on macOS: We can securely store it using the Keychain
    - on Windows: We can use `CryptProtectData` to protect it from other user's acccounts, *but any application running as the original user can decrypt it*. To securely store it, we need to protect it with another secret (i.e. a passcode or similar, which could be passed to `CryptProtectData` but the user would need to re-enter every app start).
    - on Linux: We would need to protect it with an additional secret, that the user would need to re-enter every app start.
  - Since there is no cross-platform secure way of storing it that doesn't require entering some other password/passcode to decrypt it... let's just make the user enter their *account* password for now. We can support this later.

## JWT Format

**SERVER TODO:** Store the private key / secret without storing it in the server source code / files that are publicly available. (Example: Wrapper that looks in a predefined location on server startup for an existing key and, if not present, generates a new random one & persists it on the server securely.)

### access_token:

More info's about [JWT Signatures](https://auth0.com/blog/rs256-vs-hs256-whats-the-difference/).

JWT header:

```json
{
  "alg": "EdDSA",
  "typ": "JWT",
}
```

JWT payload:

```json
{
  "iss": "go.micro.auth",
  "sub": "[KING]Fast",
  "aud": [
    "https://lobby.wz2100.net"
  ],
  "exp": 1663905852,
  "nbf": 1663902252,
  "iat": 1663902252,
  "jti": "xxxxxxxx-934d-4cd2-84fb-bd650d3a1ded",
  "roles": [
    "user"
  ]
}
```

We store roles in the JWT that makes it simpler to check rights on the server without an extra call to Inspect the token each time.

The client DOES NOT need to read the JWT, it only needs to forward it.

## /api/auth/v1/register : POST

Register a new username within the lobby.

**SERVER NOTES**:

- The server must then **additionally** use appropriate measures, such as a password-hashing-function like bcrypt / scrypt (or whatever ends up supported on the environments we're looking into) to store the password / password-derived-key that the client sends.
  - We have choosen argon2-id here, if it's to slow/takes to much compute power we switch to bcrypt.
- Must implement rate-limiting by (at least):
  - Account
  - IP address

**ACL**: Verify authentication token does *not* exist in request, rate-limited by account &amp; IP

**RATE Limit** 10 per day/1 per minute

**returns**: a JSON object, containing (on success) an `accessToken` JWT used for authenticating future requests

On success:

HTTP Code: 200

```json
{
  "id"                                : "<users-uuid-here>",
  "accessToken"                       : "<b64-encoded-JWT-token>",
  "accessTokenExpiresAt"              : "<unix-timestamp>",
  "refreshToken"                      : "<b64-encoded-JWT-token>",
  "refreshTokenExpiresAt"             : "<unix-timestamp>"
}
```

On failure:

HTTP Code: 401

```json
{
    "errors"          : [
        {
            "id"          : "USER_EXISTS",
            "message"     : "That user name exists, try a different user name"
        }
    ]
}
```

## /api/auth/v1/login : POST

Login to lobby with the given username and password-derived-key.

**CLIENT NOTES**:

- **Before** sending to the server, the user's password is run through the key derivation function **Argon2id** (with parameters to be specified here after testing performance on various systems). That key is then effectively the user's password.

**SERVER NOTES**:

- The server must then **additionally** use appropriate measures, such as a password-hashing-function like bcrypt / scrypt (or whatever ends up supported on the environments we're looking into) to store the password / password-derived-key that the client sends.
- Must implement rate-limiting by (at least):
  - Account
  - IP address

**ACL**: Verify authentication token does *not* exist in request, rate-limited by account &amp; IP

**RATE Limit** 100 per day/30 per hour/10 per minute

**returns**: a JSON object, containing (on success) an `access_token` JWT used for authenticating future requests

On success:

HTTP Code: 200

```json
{
  "id"                                : "<users-uuid-here>",
  "accessToken"                       : "<b64-encoded-JWT-token>",
  "accessTokenExpiresAt"              : "<unix-timestamp>",
  "refreshToken"                      : "<b64-encoded-JWT-token>",
  "refreshTokenExpiresAt"             : "<unix-timestamp>"
}
```

On failure:

HTTP Code: 401

```json
{
    "errors"          : [
        {
            "id"         : "INVALID_USERNAME_OR_PASSWORD",
            "message"    : "Invalid username and/or password."
        }
    ]
}
```

## /api/gamedb/v1/ : GET

Lists the currently-active (joinable / not-yet-started) games in the Lobby.

**ACL**: Public

**returns**: object

```json
{
    "success": true,
    "data": {
        "games": [
            {
                "gameUUID"       : "Game's UUID",
                "host"           :
                {
                        "availability"  : [ "ipv4", "ipv6" ],
                        "country"       : "UK (FIPS 10-4 country code)",
                        "player"        : {
                            "name" : "Fastdeath",
                            "rank": "not-sure-what-to-put-here"
                        },
                },
                "description"    : "Test 1",
                "currentPlayers" : 1,
                "maxPlayers"     : 3,
                "multiVer"       : "Warzone 2100 master",
                "wzVerMajor"     : 0x1000,
                "wzVerMinor"     : 0,
                "isPrivate"      : false,
                "modlist"        : "",
                "mapname"        : "Sk-Rush-T1",
                "limits"         : 0x0
            },
            {
                "more_games"    : "here"
            }
        ]
    }
}
```

## /api/gamedb/v1/ : POST

Creates a new game in the lobby (hosted by the current, authenticated user).

**CLIENT NOTES:**

- As is the case with the current lobby server, only direct connections to the server are supported. (i.e. No HTTP proxy support - this must be disabled in libcurl.)

**ACL**: Any authenticated user with "host" privileges

**input**:
    "gamedetails": a JSON object containing the game information


**returns**: a JSON object

On success:

```json
{
    "uuid" : "<UUID>",
}
```

On failure:

```json
{
    "error"          : {
        "id"         : "FAILED_VERIFY_HOST_CONNECTION",
        "message"    : "Unable to connect to host. A firewall may be blocking access"
    }
}
```

## /api/gamedb/v1/&lt;GAME-UUID&gt;/ : GET

Get detailed information about a game in the Lobby.

**ACL**: Any authenticated user

**returns**: a JSON object

```json
{
    "host"           : {
        "availability"  : [ "ipv4", "ipv6" ],
        "country"       : "UK",
        "player"        : {
            "name" : "Fastdeath",
            "rank": "not-sure-what-to-put-here"
            // ...
        },
    },
    "description"    : "Test 1",
    "currentPlayers" : 1,
    "maxPlayers"     : 3,
    "multiVer"       : "Warzone 2100 master",
    "wzVerMajor"     : 0x1000,
    "wzVerMinor"     : 0,
    "isPrivate"      : false,
    "modlist"        : "",
    "mapname"        : "Sk-Rush-T1",
    "limits"         : 0x0,
    "players":  [
        {
            "name": "Fastdeath",
            "rank": "not-sure-what-to-put-here",
            "team": "a",
            "isAI": false,
            "available": false
        },
        {
            "name": "NullBot",
            "team": "b",
            "isAI": true,
            "available": false
        },
        {
            "name": "pastdue",
            "rank": "not-sure-what-to-put-here",
            "team": "c",
            "isAI": false,
            "available": false
        }
        // ...
    ]
}
```

## /api/gamedb/v1/&lt;GAME-UUID&gt;/ : PUT

Changes a game.

**ACL**: The game's host (authenticated user) only.

**optional arguments**: `description`, `isPrivate`, `mapname`

**returns**: NONE

## /api/gamedb/v1/&lt;GAME-UUID&gt;/ : DELETE

Deletes the game from the Lobby (prior to it starting).

**ACL**: The game's host (authenticated user), or an admin.

**returns**: boolean

## /api/gamedb/v1/&lt;GAME-UUID&gt;/ip : POST

Called by the host to add a host IP address to their game.

The server obtains the connecting IP address, (potentially) verifies the host is accessible, and returns a JSON object.

**CLIENT NOTES**:

- The client **MUST** try to initalize a connection with IPv4 and IPv6 at the same time, whatever succeds will be added.

**SERVER NOTES:**

- The server obtains the connecting (public) IP address, and checks whether it is IPv4 or IPv6. This information is returned to the client (as `host/availability`).
- The verifiable connecting IP address must be that of the host. Thus, we cannot support HTTP proxies (which can claim to be forwarding for an IP address, but can lie), but we *can* support environments like AppEngine that can verifiably provide the connecting IP.
- The server can also geolocate the IP address to determine the host's country.
  - One possible Python option: https://pythonhosted.org/python-geoip/
  - There's also a whole webservice that uses the Geolite2 database available, but it's probably more than we need (https://github.com/maxmind/GeoIP2-python).
- The server can also attempt a connection to the host, to verify that it's accessible, and return an error if not.
  - Outbound sockets [work on AppEngine](https://cloud.google.com/appengine/docs/standard/python/sockets/), so can be used to connect to the host to verify.
- A game should be only listed when the client added an IP, not before that.

**ACL**: The game's host (authenticated user) only

**returns**: JSON object

On success:

```json
{
    "host" : {
        "availability"  : [ "ipv4", "ipv6" ],
        "newavailability" : [ "ipv6" ]
    }
}
```

On failure:

```json
{
    "errors"  : [
        {
            "id"         : "FAILED_VERIFY_HOST_CONNECTION",
            "message"    : "Unable to connect to host. A firewall may be blocking access"
        }
    ]
}
```

## /api/gamedb/v1/&lt;GAME-UUID&gt;/join/ : POST

Called by an authenticated user to "join" a game. This registers the intent on the server, and returns an object containing the information required for the client to connect to the game host.

**NOTES**:

- This is the only way to obtain the game host's IP address(es). They must not be available via any other method.
- The `secret` that is returned is a unique, unpredictable, short-lived, game-scoped identifier (==could be a UUID or a separate authenticated JWT==) that the client then sends to the host when connecting. (The host can then use this to identify the client to the lobby server. It cannot be used by any other user/host to identify a client.)
  - This helps prevent: (a) users from spoofing other users' identities when connecting to a host, (b) hosts from spoofing user joins.

**CLIENT NOTES**:

- The client needs to store the secret for other game related actions.

**SERVER NOTES**:

- Stores the secret in within the gamedb, for future actions.
- Must implement rate-limiting by (at least):
  - Account

**ACL**: Any authenticated user with "join" privileges, rate-limited

**RATE-Limit** user: 5 per minute, 60 per hour

**returns**: JSON object

```json
{
    "ips": [
        ["127.0.0.1", 2100]
        ["0:0:0:0:0:0:0:1", 2100],
    ],
    "secret": "<CLIENT_CONNECT_ID_TOKEN>"
}
```

On failure:

Headers:

```text
X-UserRateLimit-Limit: 5
X-UserRateLimit-Remaining: 0
X-UserRateLimit-Reset: <unix timestamp from server>
```

```json
{
    "errors" : [
        {
            "id"         : "TO_MANY_REQUESTS",
            "message"    : "To many requests"
        }
    ]
}
```


## /api/gamedb/v1/&lt;GAME-UUID&gt;/player/ : POST

Authenticate the join request that the host received from a new player, using the `secret` that the client transmitted to the host. If the join request is valid for this game, the associated player is added to the game &amp; the player details are returned to the host.

**ACL**: The game's host (authenticated user) only

**arguments**: `player-uuid`, `secret`, `slot`, `team`

**returns**: a JSON object, containing (on success) the validated player details

On success:

HTTP Code: 200

```json
{
    "player": {
        "uuid": "<player-uuid-here>",
        "name": "Fastdeath",
        "rank": "not-sure-what-to-put-here",
        "stats": {
            "lifetime": {
                "games": 101,
                "hosted": 12,
                "reports": 1,
                "abandons": 2,
                "players_kicked": 7
            },
            "recent": {
                "games": 20,
                "hosted": 2,
                "reports": 0,
                "abandons": 2,
                "players_kicked": 0
            }
        }
    }
}
```

## /api/gamedb/v1/&lt;GAME-UUID&gt;/player/&lt;PLAYER-UUID&gt; : PUT

Update a player.

**optional arguments**: `slot`, `team`, `name`

## /api/gamedb/v1/&lt;GAME-UUID&gt;/player/&lt;PLAYER-UUID&gt; : DELETE

Delete (leave game during lobby) a player from the game, game owners can delete any player by given then `slot` argument.

**ACL**: The game's host (authenticated user) can delete any player; players joined to the game can only delete themselves.

**optional arguments**: `slot`


## /api/gamedb/v1/&lt;GAME-UUID&gt;/kicked/ : POST

Let the lobbyserver know that the connecting player have been kicked during the Game.

**CLIENT NOTES**:

- The client **MUST** call this endpoint whenever he got kicked from the game, this is important for Hoster stats.

**SERVER NOTES**:

- The Lobbyserver updates the host players profile and add +1 to players_kicked.

**ACL**: An authenticated user that is in that game.

Request:

```json
{
    "secret": "<secret-from-join>"
}
```

On failure:

```json
{
    "errors" : [
        {
            "id"         : "NOT_IN_THAT_GAME",
            "message"    : "Your not in that game"
        }
    ]
}
```

## /api/profile/v1/ : GET

Get detailed information about the current authenticated user's account

**CLIENT NOTES**:

- `server_messages` contains a list of messages to be displayed to the user about their account
  - These could be displayed in the console area when viewing the lobby, or in a new "player info" screen

**ACL**: Any authenticated user

**returns**: a JSON object

```json
{
    "name" : "Fastdeath",
    "rank": "not-sure-what-to-put-here",
    "stats": {
        "lifetime": {
            "games": 101,
            "hosted": 12,
            "reports": 1,
            "abandons": 2,
            "players_kicked": 7
        },
        "recent": {
            "games": 20,
            "hosted": 2,
            "reports": 0,
            "abandons": 2,
            "players_kicked": 0
        }
    },
    "currently_hosting_games": [
        "<game UUID>"
    ],
    "server_messages": [
        {
            "id"         : "ACCOUNT_TEMP_BAN_DETAILS",
            "message"    : "Your account has been temporarily banned due to: inappropriate language.\nYou will not be able to join or host games until the ban expires."
        }
    ]
}
```

## /api/profile/v1/&lt;UUID&gt; : GET

Same as profile /me but for foreign users.

**ACL**: An admin

## /api/auth/v1/password : PUT

Change the current account's password

**arguments**: `currentpassword`, `newpassword`

**CLIENT NOTES**:

- The existing password **must** be supplied, so we'll need UI that has two password fields (ideally, 3 - so the user must type the new password twice).
- As with **/api/auth/v1/login**, the "password" is actually a password-derived-key (using all the same specifics as noted above).

**SERVER NOTES**:

- The existing password **must** be supplied, and must match what is stored on the server.
- Must ensure (ex. via db row locking) an atomic UPDATE of the record / row that stores the hashed password **ONLY IF the proper current password is also supplied**. i.e. Both the check for the current password and the update to the new password should be part of the same **atomic** operation.
  - Changing an account password should invalidate all existing authentication tokens for that account.

**ACL**: Any authenticated user

**returns**: a JSON object

On success:

HTTP Code: 200

```json
{
  "id"                                : "<users-uuid-here>",
  "accessToken"                       : "<b64-encoded-JWT-token>",
  "accessTokenExpiresAt"              : "<unix-timestamp>",
  "refreshToken"                      : "<b64-encoded-JWT-token>",
  "refreshTokenExpiresAt"             : "<unix-timestamp>"
}
```

On failure:

HTTP Code: 401

```json
{
    "errors"          : [
        {
            "id"         : "INVALID_PASSWORD",
            "message"    : "The existing password does not match."
        }
    ]
}
```
