package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	userPB "wz2100.net/microlobby/shared/proto/user"
)

func UserFromGinContext(c *gin.Context) (*userPB.User, error) {
	userIface, ok := c.Get("user")
	if !ok {
		return nil, errors.New("Failed to get user from context")
	}

	return userIface.(*userPB.User), nil
}
