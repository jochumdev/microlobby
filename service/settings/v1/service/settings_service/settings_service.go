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

func (h *Handler) Create(ctx context.Context, in *settingsservicepb.CreateRequest, out *settingsservicepb.Setting) error {
	return nil
}

func (h *Handler) Update(ctx context.Context, in *settingsservicepb.UpdateRequest, out *settingsservicepb.Setting) error {
	return nil
}

func (h *Handler) Get(ctx context.Context, in *settingsservicepb.GetRequest, out *settingsservicepb.Setting) error {
	result, err := db.SettingsGet(ctx, in.Id, in.OwnerId, in.Service, in.Name)
	if err != nil {
		return err
	}

	out.Id = result.ID.String()
	out.OwnerId = result.OwnerID.String()
	out.Service = result.Service
	out.Name = result.Name
	out.Content = result.Content
	out.CreatedAt = timestamppb.New(result.CreatedAt)
	if !result.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(result.UpdatedAt.Time)
	}

	return nil
}

func (h *Handler) List(ctx context.Context, in *settingsservicepb.ListRequest, out *settingsservicepb.SettingsList) error {
	results, err := db.SettingsList(ctx, in.Id, in.OwnerId, in.Service, in.Name, in.Limit, in.Offset)
	if err != nil {
		return err
	}

	// Copy the data to the result
	for _, result := range results {
		row := &settingsservicepb.Setting{
			Id:        result.ID.String(),
			OwnerId:   result.OwnerID.String(),
			Service:   result.Service,
			Name:      result.Name,
			Content:   result.Content,
			CreatedAt: timestamppb.New(result.CreatedAt),
		}
		if !result.UpdatedAt.IsZero() {
			row.UpdatedAt = timestamppb.New(result.UpdatedAt.Time)
		}
		out.Data = append(out.Data, row)
	}

	return nil
}
