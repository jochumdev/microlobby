package badwordshandler

import (
	"context"

	goaway "github.com/TwiN/go-away"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4/server"

	"jochum.dev/jo-micro/components"
	"jochum.dev/jo-micro/router"
	"wz2100.net/microlobby/shared/proto/badwordspb/v1"
)

const Name = "badwordsHandler"

type Handler struct {
	cReg        *components.Registry
	initialized bool
}

func New() *Handler {
	return &Handler{initialized: false}
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
			router.Path("/check/:request"),
			router.Endpoint(badwordspb.BadwordsV1Service.Check),
			router.Params("request"),
		),
		router.NewRoute(
			router.Method(router.MethodPost),
			router.Path("/check"),
			router.Endpoint(badwordspb.BadwordsV1Service.Check),
		),
	)

	badwordspb.RegisterBadwordsV1ServiceHandler(h.cReg.Service().Server(), h)

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

func (h *Handler) WrapHandlerFunc(ctx context.Context, req server.Request, rsp interface{}) error {
	return nil
}

func (h *Handler) IsProfane(ctx context.Context, in *badwordspb.StringRequest, out *badwordspb.BoolResponse) error {
	out.Response = goaway.IsProfane(in.Request)
	return nil
}

func (h *Handler) ExtractProfanity(ctx context.Context, in *badwordspb.StringRequest, out *badwordspb.StringResponse) error {
	out.Response = goaway.ExtractProfanity(in.Request)
	return nil
}

func (h *Handler) Censor(ctx context.Context, in *badwordspb.StringRequest, out *badwordspb.StringResponse) error {
	out.Response = goaway.Censor(in.Request)
	return nil
}

func (h *Handler) Check(ctx context.Context, in *badwordspb.StringRequest, out *badwordspb.CheckResponse) error {
	out.Profane = goaway.IsProfane(in.Request)
	out.Extracted = goaway.ExtractProfanity(in.Request)
	out.Censored = goaway.Censor(in.Request)
	return nil
}
