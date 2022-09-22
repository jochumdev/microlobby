package serviceregistry

import (
	"context"

	"go-micro.dev/v4/client"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/registry"
	"wz2100.net/microlobby/shared/component"
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
