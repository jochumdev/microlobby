package auth

import (
	"context"
	"net/http"

	goaway "github.com/TwiN/go-away"

	"go-micro.dev/v4/errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"wz2100.net/microlobby/service/badwords/v1/config"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/proto/authservicepb/v1"
)

const pkgPath = config.PkgPath + "/handler/auth"

type Handler struct {
	cRegistry *component.Registry
	logrus    component.LogrusComponent

	svcName string
}

func NewHandler(cregistry *component.Registry) (*Handler, error) {
	h := &Handler{
		cRegistry: cregistry,
		svcName:   cregistry.Service.Name(),
	}

	return h, nil
}

func (h *Handler) Start() error {
	logrus, err := component.Logrus(h.cRegistry)
	if err != nil {
		return errors.FromError(err)
	}
	h.logrus = logrus

	return nil
}

func (h *Handler) Stop() error { return nil }

func (h *Handler) UserDelete(ctx context.Context, in *authservicepb.UserIDRequest, out *emptypb.Empty) error {
	return nil
}

func (h *Handler) UserUpdateRoles(ctx context.Context, in *authservicepb.UpdateRolesRequest, out *emptypb.Empty) error {
	return nil
}

func (h *Handler) Register(ctx context.Context, in *authservicepb.RegisterRequest, out *emptypb.Empty) error {
	if goaway.IsProfane(in.Username) {
		return errors.New(h.svcName, "Badword filter matched", http.StatusBadRequest)
	}

	return nil
}

func (h *Handler) Login(ctx context.Context, in *authservicepb.LoginRequest, out *emptypb.Empty) error {
	return nil
}

func (h *Handler) Refresh(ctx context.Context, in *authservicepb.Token, out *emptypb.Empty) error {
	return nil
}
