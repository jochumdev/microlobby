package defs

const ServiceHttpProxy = "proxy"
const ServiceLobbyV1 = "lobby.v1"
const ServiceAuthV1 = "auth.v1"
const ServiceSettingsV1 = "settings.v1"

var ServicesRequired = [...]string{ServiceLobbyV1, ServiceAuthV1, ServiceSettingsV1}
