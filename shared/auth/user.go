package auth

import (
	"context"

	"go-micro.dev/v4/cmd"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/proto/authservicepb/v1"
	"wz2100.net/microlobby/shared/proto/userpb/v1"
)

func UserFromToken(ctx context.Context, token string) (*userpb.User, error) {
	clientUser := authservicepb.NewAuthV1Service(defs.ServiceAuthV1, *cmd.DefaultOptions().Client)

	user, err := clientUser.TokenDetail(ctx, &authservicepb.Token{Token: token})
	if err != nil {
		return nil, err
	}

	return user, nil
}
