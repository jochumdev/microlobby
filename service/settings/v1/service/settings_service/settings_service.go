package settingsService

import (
	"context"

	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
)

type Handler struct{}

func NewHandler() (*Handler, error) {
	return &Handler{}, nil
}

func (h *Handler) Create(context.Context, *settingsservicepb.CreateRequest, *settingsservicepb.Setting) error {
	return nil
}

func (h *Handler) Update(context.Context, *settingsservicepb.UpdateRequest, *settingsservicepb.Setting) error {
	return nil
}

func (h *Handler) Get(context.Context, *settingsservicepb.GetRequest, *settingsservicepb.Setting) error {
	return nil
}

func (h *Handler) List(context.Context, *settingsservicepb.ListRequest, *settingsservicepb.SettingsList) error {
	return nil
}
