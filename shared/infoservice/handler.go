package infoservice

import (
	"context"

	empty "github.com/golang/protobuf/ptypes/empty"
	"wz2100.net/microlobby/shared/component"
	pb "wz2100.net/microlobby/shared/proto/infoservice"
)

// Handler is the handler for minadmin.io/shared/proto/maservice
type Handler struct {
	comRegistry *component.Registry
	routes      []*pb.RoutesReply_Route
}

// NewHandler returns a new srv/user pb handler
func NewHandler(comRegistry *component.Registry, routes []*pb.RoutesReply_Route) *Handler {
	return &Handler{comRegistry, routes}
}

// Health returns information about the health of this service.
func (h *Handler) Health(ctx context.Context, req *empty.Empty, rsp *pb.HealthReply) error {
	healthMap := h.comRegistry.Health(ctx)

	hasError := false

	rsp.Infos = make(map[string]*pb.HealthReply_HealthInfo)
	for name, info := range healthMap {
		if info.IsError {
			hasError = true
		}

		rsp.Infos[name] = &pb.HealthReply_HealthInfo{
			Message: info.Message,
			IsError: info.IsError,
		}
	}

	rsp.HasError = hasError

	return nil
}

// Routes returns the registered routes
func (h *Handler) Routes(ctx context.Context, req *empty.Empty, rsp *pb.RoutesReply) error {
	rsp.Routes = h.routes

	return nil
}
