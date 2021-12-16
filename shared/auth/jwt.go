package auth

import "github.com/golang-jwt/jwt"

type JWTMicrolobbyClaims struct {
	*jwt.StandardClaims
	Roles []string `json:"roles,omitempty"`
}
