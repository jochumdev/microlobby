package config

const (
	Name     = "microlobby.proxy"
	ProxyURI = "proxy"
	Version  = "not set"
	PkgPath  = "wz2100.net/microlobby/service/http_proxy"

	SettingNameJWTRefreshTokenPub  = "jwt|refreshtoken|pub"
	SettingNameJWTRefreshTokenPriv = "jwt|refreshtoken|priv"
	SettingNameJWTAccessTokenPub   = "jwt|accesstoken|pub"
	SettingNameJWTAccessTokenPriv  = "jwt|accesstoken|priv"
)

// These Services are required on health checks
var ServicesRequired = [...]string{"microlobby.gamedb.v1", "microlobby.lobby.v3", "microlobby.auth.v1", "microlobby.settings.v1"}
