package badwords

import (
	"context"

	goaway "github.com/TwiN/go-away"

	"wz2100.net/microlobby/service/badwords/v1/version"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/proto/badwordspb/v1"
)

const pkgPath = version.PkgPath + "/service/badwords"

type Handler struct {
	cRegistry *component.Registry
	svcName   string
}

func NewHandler(cregistry *component.Registry) (*Handler, error) {
	h := &Handler{
		cRegistry: cregistry,
		svcName:   cregistry.Service.Name(),
	}

	return h, nil
}

func (s *Handler) IsProfane(ctx context.Context, in *badwordspb.StringRequest, out *badwordspb.BoolResponse) error {
	out.Response = goaway.IsProfane(in.Request)
	return nil
}

func (s *Handler) ExtractProfanity(ctx context.Context, in *badwordspb.StringRequest, out *badwordspb.StringResponse) error {
	out.Response = goaway.ExtractProfanity(in.Request)
	return nil
}

func (s *Handler) Censor(ctx context.Context, in *badwordspb.StringRequest, out *badwordspb.StringResponse) error {
	out.Response = goaway.Censor(in.Request)
	return nil
}

func (s *Handler) Check(ctx context.Context, in *badwordspb.StringRequest, out *badwordspb.CheckResponse) error {
	out.Profane = goaway.IsProfane(in.Request)
	out.Extracted = goaway.ExtractProfanity(in.Request)
	out.Censored = goaway.Censor(in.Request)
	return nil
}
