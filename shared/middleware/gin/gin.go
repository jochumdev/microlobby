package gin

import (
	"log"

	"net/http"

	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"wz2100.net/microlobby/shared/utils"
)

const pkgPath = "wz2100.net/microlobby/shared/middleware/gin"

// UserSrvMiddleware is a middleware for gin that gets the user from the user service.
func UserSrvMiddleware(registry *component.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _, err := utils.ExtractToken(c.GetHeader("Authorization"))
		if err != nil {
			logrusc, err := component.Logrus(registry)
			if err != nil {
				log.Fatalf("Failed to get the logger from the registry, Error was: %s", err)
				return
			}

			logrusc.WithFunc(pkgPath, "UserSrvMiddleware").
				WithFields(logrus.Fields{
					"error": err,
				}).
				Error("ExtractToken failed")
			return
		}

		user, err := auth.UserFromToken(c, token)
		if err != nil {
			logrusc, err := component.Logrus(registry)
			if err != nil {
				log.Fatalf("Failed to get the logger from the registry, Error was: %s", err)
				return
			}

			logrusc.WithFunc(pkgPath, "UserSrvMiddleware").
				WithFields(logrus.Fields{
					"error": err,
				}).
				Error("UserFromToken failed")
			return
		}

		c.Set("user", user)
	}
}

// RequireUserMiddleware aborts if we don't have a valid user.
func RequireUserMiddleware(registry *component.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "message": "Authorization failed"})
			c.Abort()
			return
		}

		// user := userIface.(*user)

		// jsonb, _ := json.Marshal(user)
		// logger.WithFunc(pkgPath, "requireUserMiddleware").
		// 	WithFields(log.Fields{
		// 		"user": string(jsonb),
		// 	}).
		// 	Debug("User")
	}
}
