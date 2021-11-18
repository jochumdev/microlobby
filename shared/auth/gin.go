package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"wz2100.net/microlobby/shared/proto/userpb/v1"
)

func UserFromGinContext(c *gin.Context) (*userpb.User, error) {
	userIface, ok := c.Get("user")
	if !ok {
		return nil, errors.New("failed to get user from context")
	}

	return userIface.(*userpb.User), nil
}
