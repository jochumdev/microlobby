package defs

const ServiceHttpProxy = "microlobby.proxy"
const ServiceLobbyV1 = "microlobby.lobby.v1"
const ServiceAuthV1 = "microlobby.auth.v1"
const ServiceSettingsV1 = "microlobby.settings.v1"

const ProxyURIHttpProxy = "proxy"
const ProxyURILobby = "lobby"
const ProxyURIAuth = "auth"
const ProxyURISettings = "settings"

var ServicesRequired = [...]string{ServiceLobbyV1, ServiceAuthV1, ServiceSettingsV1}
