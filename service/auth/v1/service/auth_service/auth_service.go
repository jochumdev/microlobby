package authService

import (
	"context"
	"errors"

	"github.com/golang/protobuf/ptypes/empty"
	"go-micro.dev/v4/metadata"
	"wz2100.net/microlobby/service/auth/v1/db"
	"wz2100.net/microlobby/shared/proto/authservicepb/v1"
	"wz2100.net/microlobby/shared/proto/userpb/v1"
	"wz2100.net/microlobby/shared/utils"
)

type Handler struct{}

func NewHandler() (*Handler, error) {
	return &Handler{}, nil
}

func (s *Handler) UserList(ctx context.Context, in *authservicepb.ListRequest, out *authservicepb.UserListResponse) error {
	results, err := db.UserList(ctx, in.Limit, in.Offset)
	if err != nil {
		return err
	}

	// Copy the data to the result
	for _, result := range results {
		out.Data = append(out.Data, &authservicepb.UserListResponse_User{
			Id:       result.ID.String(),
			Username: result.Username,
			Email:    result.Email,
		})
	}

	return nil
}

func (s *Handler) UserDetail(ctx context.Context, in *authservicepb.UserIDRequest, out *userpb.User) error {
	result, err := db.UserDetail(ctx, in.UserId)
	if err != nil {
		return err
	}

	out.Id = result.ID.String()
	out.Email = result.Email
	out.Username = result.Username
	out.Roles = result.Roles

	return nil
}

func (s *Handler) UserDelete(ctx context.Context, in *authservicepb.UserIDRequest, out *empty.Empty) error {
	err := db.UserDelete(ctx, in.UserId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Handler) UserUpdateRoles(ctx context.Context, in *authservicepb.UpdateRolesRequest, out *userpb.User) error {
	result, err := db.UserUpdateRoles(ctx, in.UserId, in.Roles)
	if err != nil {
		return err
	}

	out.Id = result.ID.String()
	out.Email = result.Email
	out.Username = result.Username
	out.Roles = result.Roles

	return nil
}

func (s *Handler) TokenList(ctx context.Context, in *authservicepb.ListRequest, out *authservicepb.TokenListResponse) error {
	out.Data = append(out.Data, &authservicepb.TokenListResponse_TokenData{
		UserId: "1",
		Token:  "123456",
	}, &authservicepb.TokenListResponse_TokenData{
		UserId: "2",
		Token:  "234567",
	})
	return nil
}

func (s *Handler) TokenDetail(ctx context.Context, in *authservicepb.Token, out *userpb.User) error {
	return nil
}

func (s *Handler) TokenDelete(ctx context.Context, in *authservicepb.Token, out *empty.Empty) error {
	return nil
}

func (s *Handler) TokenRefresh(ctx context.Context, in *authservicepb.Token, out *authservicepb.Token) error {
	return nil
}

func (s *Handler) Register(ctx context.Context, in *userpb.User, out *empty.Empty) error {
	return nil
}

func (s *Handler) Login(ctx context.Context, in *authservicepb.LoginRequest, out *authservicepb.Token) error {
	return nil
}

func (s *Handler) Logout(ctx context.Context, in *empty.Empty, out *empty.Empty) error {
	// Extract the token
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return errors.New("failed to get metadata from context")
	}
	authStr, ok := md.Get("X-Microlobby-Authorization")
	if !ok {
		return errors.New("failed to geth auth string from context")
	}
	token, _, err := utils.ExtractToken(authStr)
	if err != nil {
		return err
	}

	// Delete the token
	req := &authservicepb.Token{
		Token: token,
	}
	reqOut := &empty.Empty{}
	return s.TokenDelete(ctx, req, reqOut)
}
