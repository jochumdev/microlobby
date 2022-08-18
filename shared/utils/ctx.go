package utils

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go-micro.dev/v4/metadata"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/proto/userpb/v1"
)

// CtxFromRequest adds HTTP request headers to the context as metadata
func CtxFromRequest(c *gin.Context, r *http.Request) context.Context {
	md := make(metadata.Metadata, len(r.Header)+1)
	for k, v := range r.Header {
		if k == "Authorization" {
			k = "X-Microlobby-Authorization"
		}
		md[k] = strings.Join(v, ",")
	}

	userIface, ok := c.Get("user")
	if ok {
		jsonb, err := json.Marshal(userIface.(*userpb.User))
		if err == nil {
			md["user"] = string(jsonb)
		}
	}

	return metadata.NewContext(c, md)
}

func CtxMetadataUser(ctx context.Context) (*userpb.User, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, errors.New("no metadata?")
	}

	data, ok := md.Get("user")
	if !ok || len(data) < 1 {
		return nil, errors.New("no user in context")
	}

	user := userpb.User{}
	err := json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func CtxForService(ctx context.Context) context.Context {
	user := &userpb.User{
		Id:       "00000000-0000-0000-0000-000000000001",
		Username: "service",
		Email:    "service@wz2100.net",
		Roles:    []string{auth.ROLE_SERVICE},
	}

	md := make(metadata.Metadata, 1)
	jsonb, err := json.Marshal(user)
	if err == nil {
		md["user"] = string(jsonb)
	}

	return metadata.MergeContext(ctx, md, true)
}
