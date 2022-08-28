package serviceregistry

import (
	"context"

	"go-micro.dev/v4/client"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"google.golang.org/protobuf/types/known/emptypb"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
	"wz2100.net/microlobby/shared/utils"
)

type ServiceListResult map[*registry.Service][]*registry.Endpoint

type WrappedEndpoint struct {
	Pre     string
	Handler string
}

func Endpoints(ctx context.Context, cRegistry *component.Registry, service *registry.Service) ([]*registry.Endpoint, error) {
	if len(service.Endpoints) > 0 {
		eps := append([]*registry.Endpoint{}, service.Endpoints...)
		return eps, nil
	}
	// lookup the endpoints otherwise
	newServices, err := cRegistry.Service.Options().Registry.GetService(service.Name)
	if err != nil {
		return []*registry.Endpoint{}, err
	}
	if len(newServices) == 0 {
		return []*registry.Endpoint{}, err
	}

	eps := []*registry.Endpoint{}
	for _, s := range newServices {
		eps = append(eps, s.Endpoints...)
	}

	return eps, nil
}

func ListEndpoints(ctx context.Context, cRegistry *component.Registry) (ServiceListResult, error) {
	services, err := cRegistry.Service.Options().Registry.ListServices()
	if err != nil {
		return nil, err
	}

	endpoints := make(ServiceListResult)
	for _, service := range services {
		eps, err := Endpoints(ctx, cRegistry, service)
		if err != nil {
			continue
		}

		endpoints[service] = eps
	}

	return endpoints, nil
}

func FindByEndpoint(ctx context.Context, cRegistry *component.Registry, endpoint string) ([]*registry.Service, error) {
	services, err := ListEndpoints(ctx, cRegistry)
	if err != nil {
		return []*registry.Service{}, err
	}

	result := []*registry.Service{}
	for s, eps := range services {
		for _, ep := range eps {
			if ep.Name == endpoint {
				result = append(result, s)
			}
		}
	}

	return result, nil
}

func CallEndPoints(ctx context.Context, cRegistry *component.Registry, endpoint string, req, out interface{}) error {
	services, err := ListEndpoints(ctx, cRegistry)
	if err != nil {
		return errors.FromError(err)
	}

	for svc, eps := range services {
		for _, ep := range eps {
			if ep.Name == endpoint {
				req := cRegistry.Client.NewRequest(svc.Name, endpoint, req, client.WithContentType("application/json"))
				if err := cRegistry.Client.Call(ctx, req, out); err != nil {
					return errors.FromError(err)
				}
			}
		}
	}

	return nil
}

func NewHandlerWrapper(cRegistry *component.Registry, routes []*infoservicepb.RoutesReply_Route) server.HandlerWrapper {
	preEndpoints := make(map[string]string)
	postEndpoints := make(map[string]string)

	for _, r := range routes {
		if len(r.PreEndpoint) > 0 {
			preEndpoints[r.Endpoint] = r.PreEndpoint
		}

		if len(r.PostEndpoint) > 0 {
			postEndpoints[r.Endpoint] = r.PostEndpoint
		}
	}

	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// Wrap into a pre call?
			if v, ok := preEndpoints[req.Endpoint()]; ok && len(v) > 0 {
				tmp := &emptypb.Empty{}
				if err := CallEndPoints(utils.CtxForService(ctx), cRegistry, preEndpoints[req.Endpoint()], req.Body(), tmp); err != nil {
					return err
				}
			}

			// If no error happened execute the original function.
			if err := h(ctx, req, rsp); err != nil {
				return err
			}

			// Wrap into a post call?
			if v, ok := postEndpoints[req.Endpoint()]; ok && len(v) > 0 {
				tmp := &emptypb.Empty{}
				if err := CallEndPoints(utils.CtxForService(ctx), cRegistry, postEndpoints[req.Endpoint()], req.Body(), tmp); err != nil {
					return err
				}
			}

			return nil
		}
	}
}
