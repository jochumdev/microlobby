package auth

import (
	"context"

	"go-micro.dev/v4/cmd"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/proto/authservicepb/v1"
	"wz2100.net/microlobby/shared/proto/userpb/v1"
)

func UserFromContext(ctx context.Context) (*userpb.User, error) {
	clientUser := authservicepb.NewAuthV1Service(defs.ServiceAuthV1, *cmd.DefaultOptions().Client)

	user, err := clientUser.UserDetail(ctx, &authservicepb.UserIDRequest{UserId: "me"})
	if err != nil {
		return nil, err
	}

	return user, nil
}
