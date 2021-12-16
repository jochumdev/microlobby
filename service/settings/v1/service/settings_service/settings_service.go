package settingsService

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"
	"wz2100.net/microlobby/service/settings/v1/db"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
)

type Handler struct{}

func NewHandler() (*Handler, error) {
	return &Handler{}, nil
}

func (h *Handler) translateDBSettingToPB(dbs *db.Setting, out *settingsservicepb.Setting) {
	out.Id = dbs.ID.String()
	out.OwnerId = dbs.OwnerID.String()
	out.Service = dbs.Service
	out.Name = dbs.Name
	out.Content = dbs.Content
	out.CreatedAt = timestamppb.New(dbs.CreatedAt)
	if !dbs.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(dbs.UpdatedAt.Time)
	}
}

func (h *Handler) Create(ctx context.Context, in *settingsservicepb.CreateRequest, out *settingsservicepb.Setting) error {
	result, err := db.SettingsCreate(ctx, in)
	if err != nil {
		return err
	}

	h.translateDBSettingToPB(result, out)
	return nil
}

func (h *Handler) Update(ctx context.Context, in *settingsservicepb.UpdateRequest, out *settingsservicepb.Setting) error {
	result, err := db.SettingsUpdate(ctx, in.Id, in.Content)
	if err != nil {
		return err
	}

	h.translateDBSettingToPB(result, out)
	return nil
}

func (h *Handler) Get(ctx context.Context, in *settingsservicepb.GetRequest, out *settingsservicepb.Setting) error {
	result, err := db.SettingsGet(ctx, in.Id, in.OwnerId, in.Service, in.Name)
	if err != nil {
		return err
	}

	h.translateDBSettingToPB(result, out)
	return nil
}

func (h *Handler) List(ctx context.Context, in *settingsservicepb.ListRequest, out *settingsservicepb.SettingsList) error {
	results, err := db.SettingsList(ctx, in.Id, in.OwnerId, in.Service, in.Name, in.Limit, in.Offset)
	if err != nil {
		return err
	}

	// Copy the data to the result
	for _, result := range results {
		row := &settingsservicepb.Setting{}
		h.translateDBSettingToPB(&result, row)
		out.Data = append(out.Data, row)
	}

	return nil
}
