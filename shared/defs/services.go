package defs

// Internal service name - also for settings
const ServiceHttpProxy = "microlobby.proxy"
const ServiceLobbyV1 = "microlobby.lobby.v1"
const ServiceAuthV1 = "microlobby.auth.v1"
const ServiceSettingsV1 = "microlobby.settings.v1"
const ServiceBadwordsV1 = "microlobby.badwords.v1"

// The part in the url of the service
const ProxyURIHttpProxy = "proxy"
const ProxyURILobby = "lobby"
const ProxyURIAuth = "auth"
const ProxyURISettings = "settings"
const ProxyURIBadwords = "badwords"

// These Services are required on health checks
var ServicesRequired = [...]string{ServiceLobbyV1, ServiceAuthV1, ServiceSettingsV1}
