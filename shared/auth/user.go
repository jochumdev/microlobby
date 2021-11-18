package auth

import (
	"context"

	"go-micro.dev/v4/cmd"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/proto/authservice"
	spb "wz2100.net/microlobby/shared/proto/user"
)

func UserFromContext(ctx context.Context) (*spb.User, error) {
	clientUser := authservice.NewAuthService(defs.ServiceAuthV1, *cmd.DefaultOptions().Client)

	me := "me"
	user, err := clientUser.UserDetail(ctx, &authservice.UserIDRequest{UserId: &me})
	if err != nil {
		return nil, err
	}

	return user, nil
}
