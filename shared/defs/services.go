package defs

// Internal service name - also for settings
const ServiceHttpProxy = "microlobby.proxy"
const ServiceGameDBV1 = "microlobby.gamedb.v1"
const ServiceLobbyV3 = "microlobby.lobby.v3"
const ServiceAuthV1 = "microlobby.auth.v1"
const ServiceSettingsV1 = "microlobby.settings.v1"
const ServiceBadwordsV1 = "microlobby.badwords.v1"

// The part in the url of the service
const ProxyURIHttpProxy = "proxy"
const ProxyURIGameDB = "gamedb"
const ProxyURILobby = "lobby"
const ProxyURIAuth = "auth"
const ProxyURISettings = "settings"
const ProxyURIBadwords = "badwords"

// These Services are required on health checks in http_proxy
var ServicesRequired = [...]string{ServiceGameDBV1, ServiceLobbyV3, ServiceAuthV1, ServiceSettingsV1}
