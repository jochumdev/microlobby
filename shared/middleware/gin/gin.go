package middlewareGin

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"log"

	"net/http"

	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
	"wz2100.net/microlobby/shared/proto/userpb/v1"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"

	"wz2100.net/microlobby/shared/utils"
)

const pkgPath = "wz2100.net/microlobby/shared/middleware/gin"

// UserSrvMiddleware is a middleware for gin that gets the user from the access token.
func UserSrvMiddleware(registry *component.Registry) gin.HandlerFunc {
	logrusc, err := component.Logrus(registry)
	if err != nil {
		log.Printf("Failed to get the logger from the registry, Error was: %s", err)
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": "Internal server error"})
			c.Abort()
		}
	}

	ctx := component.RegistryToContext(utils.CtxForService(context.Background()), registry)
	s, err := component.SettingsV1(registry)
	if err != nil {
		logrusc.WithFunc(pkgPath, "UserSrvMiddleware").
			WithFields(logrus.Fields{
				"error": err,
			}).
			Error("failed to get the settings component")
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": "Internal server error"})
			c.Abort()
		}
	}

	var sJWTAccessTokenPub *settingsservicepb.Setting

	return func(c *gin.Context) {
		if sJWTAccessTokenPub == nil {
			sJWTAccessTokenPub, err = s.Get(ctx, "", "", defs.ServiceHttpProxy, defs.SettingNameJWTAccessTokenPub)
			if err != nil {
				logrusc.WithFunc(pkgPath, "UserSrvMiddleware").
					WithFields(logrus.Fields{
						"error": err,
					}).
					Error("failed to get the accesstoken public key")

				c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": "Internal server error"})
				c.Abort()
				return
			}
		}

		aTokenString, _, err := utils.ExtractToken(c.GetHeader("Authorization"))
		if err != nil {
			logrusc.WithFunc(pkgPath, "UserSrvMiddleware").
				WithFields(logrus.Fields{
					"error": err,
				}).
				Error("ExtractToken failed")
			return
		}

		claims := auth.JWTMicrolobbyClaims{}
		_, err = jwt.ParseWithClaims(aTokenString, &claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return ed25519.PublicKey(sJWTAccessTokenPub.Content), nil
		})
		if err != nil {
			logrusc.WithFunc(pkgPath, "UserSrvMiddleware").
				WithFields(logrus.Fields{
					"error": err,
				}).
				Error("validate token failed")
			return
		}

		// Check claims (expiration)
		if err = claims.Valid(); err != nil {
			logrusc.WithFunc(pkgPath, "UserSrvMiddleware").
				WithFields(logrus.Fields{
					"error": err,
				}).
				Error("validate claims failed")
			return

		}

		user := userpb.User{}
		user.Id = claims.Id
		user.Username = claims.Subject
		user.Roles = claims.Roles

		c.Set("user", &user)
	}
}

// RequireUserMiddleware aborts if we don't have a valid user.
func RequireUserMiddleware(registry *component.Registry) gin.HandlerFunc {
	// logrusc, err := component.Logrus(registry)
	// if err != nil {
	// 	log.Printf("Failed to get the logger from the registry, Error was: %s", err)
	// 	return func(c *gin.Context) {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": "Internal server error"})
	// 		c.Abort()
	// 	}
	// }

	return func(c *gin.Context) {
		_, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "message": "Authorization failed"})
			c.Abort()
			return
		}

		// user := userIface.(*userpb.User)
		// jsonb, _ := json.Marshal(user)

		// logrusc.WithFunc(pkgPath, "RequireUserMiddleware").
		// 	WithFields(logrus.Fields{
		// 		"user": string(jsonb),
		// 	}).
		// 	Debug("User")
	}
}

func UserFromContext(c *gin.Context) (*userpb.User, error) {
	userIface, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "message": "Authorization failed"})
		c.Abort()
		return nil, errors.New("user not in context")
	}

	return userIface.(*userpb.User), nil
}

func ForceRole(c *gin.Context, role string) bool {
	user, err := UserFromContext(c)
	if err != nil {
		return false
	}

	if !auth.HasRole(user, role) {
		c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "message": "Authorization failed"})
		c.Abort()
		return false
	}

	return true
}
