package settings

import (
	"context"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4/util/log"
	"google.golang.org/protobuf/types/known/timestamppb"
	"jochum.dev/jo-micro/components"
	"jochum.dev/jo-micro/router"
	"wz2100.net/microlobby/service/settings/v1/db"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
)

const Name = "settingsHandler"

type Handler struct {
	cReg        *components.Registry
	initialized bool
}

func New() *Handler {
	return &Handler{}
}

func MustReg(cReg *components.Registry) *Handler {
	return cReg.Must(Name).(*Handler)
}

func (h *Handler) Name() string {
	return Name
}

func (h *Handler) Priority() int {
	return 100
}

func (h *Handler) Initialized() bool {
	return h.initialized
}

func (h *Handler) Init(components *components.Registry, cli *cli.Context) error {
	if h.initialized {
		return nil
	}

	h.cReg = components

	r := router.MustReg(h.cReg)
	r.Add(
		router.NewRoute(
			router.Method(router.MethodGet),
			router.Path("/"),
			router.Endpoint(settingsservicepb.SettingsV1Service.List),
			router.Params("id", "ownerId", "service", "name", "limit", "offset"),
		),
		router.NewRoute(
			router.Method(router.MethodPost),
			router.Path("/"),
			router.Endpoint(settingsservicepb.SettingsV1Service.Create),
		),
		router.NewRoute(
			router.Method(router.MethodGet),
			router.Path("/:id"),
			router.Endpoint(settingsservicepb.SettingsV1Service.Get),
			router.Params("id", "ownerId", "service", "name"),
		),
		router.NewRoute(
			router.Method(router.MethodPut),
			router.Path("/:id"),
			router.Endpoint(settingsservicepb.SettingsV1Service.Update),
			router.Params("id"),
		),
	)

	settingsservicepb.RegisterSettingsV1ServiceHandler(h.cReg.Service().Server(), h)

	h.initialized = true
	return nil
}

func (h *Handler) Stop() error {
	return nil
}

func (h *Handler) Flags(r *components.Registry) []cli.Flag {
	return []cli.Flag{}
}

func (h *Handler) Health(context context.Context) error {
	return nil
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
	result, err := db.SettingsCreate(h.cReg, ctx, in)
	if err != nil {
		return err
	}

	h.translateDBSettingToPB(result, out)
	return nil
}

func (h *Handler) Update(ctx context.Context, in *settingsservicepb.UpdateRequest, out *settingsservicepb.Setting) error {
	result, err := db.SettingsUpdate(h.cReg, ctx, in.Id, "", "", "", in.Content)
	if err != nil {
		return err
	}

	h.translateDBSettingToPB(result, out)
	return nil
}

func (h *Handler) Upsert(ctx context.Context, in *settingsservicepb.UpsertRequest, out *settingsservicepb.Setting) error {
	result, err := db.SettingsUpsert(h.cReg, ctx, in)
	if err != nil {
		return err
	}

	h.translateDBSettingToPB(result, out)
	log.Error(out.String())
	return nil
}

func (h *Handler) Get(ctx context.Context, in *settingsservicepb.GetRequest, out *settingsservicepb.Setting) error {
	result, err := db.SettingsGet(h.cReg, ctx, in.Id, in.OwnerId, in.Service, in.Name)
	if err != nil {
		return err
	}

	h.translateDBSettingToPB(result, out)
	return nil
}

func (h *Handler) List(ctx context.Context, in *settingsservicepb.ListRequest, out *settingsservicepb.SettingsList) error {
	results, err := db.SettingsList(h.cReg, ctx, in.Id, in.OwnerId, in.Service, in.Name, in.Limit, in.Offset)
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
