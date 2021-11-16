package serviceregistry

import (
	"context"

	"go-micro.dev/v4/registry"
)

type ServiceListResult map[*registry.Service][]*registry.Endpoint

func ServiceEndpoints(service *registry.Service, reg registry.Registry, ctx context.Context) ([]*registry.Endpoint, error) {
	if len(service.Endpoints) > 0 {
		eps := append([]*registry.Endpoint{}, service.Endpoints...)
		return eps, nil
	}
	// lookup the endpoints otherwise
	newServices, err := reg.GetService(service.Name, registry.GetContext(ctx))
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

func ServiceListEndpoints(reg registry.Registry, ctx context.Context) (ServiceListResult, error) {
	services, err := reg.ListServices(registry.ListContext(ctx))
	if err != nil {
		return nil, err
	}

	endpoints := make(ServiceListResult)
	for _, service := range services {
		eps, err := ServiceEndpoints(service, reg, ctx)
		if err != nil {
			continue
		}

		endpoints[service] = eps
	}

	return endpoints, nil
}

func ServiceFindByEndpoint(endpoint string, reg registry.Registry, ctx context.Context) ([]*registry.Service, error) {
	services, err := ServiceListEndpoints(reg, ctx)
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
