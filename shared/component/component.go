package component

import (
	"context"
	"errors"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/server"
)

// const pkgPath = "wz2100.net/microlobby/shared/component"

var (
	errorRetrievingRegistry = errors.New("retrieving registry")
	errorRegistryIsNil      = errors.New("registry is nil")
)

type Component interface {
	Key() interface{}
	Priority() int8
	Name() string
	Flags() []cli.Flag
	Initialized() bool
	Init(registry *Registry, context *cli.Context) error
	Health(context context.Context) (string, bool)
}

type LogrusComponent interface {
	Logger() *logrus.Logger
	WithFunc(pkgPath string, function string) *logrus.Entry
	WithClassFunc(pkgPath, class, function string) *logrus.Entry
}

type RegistryKey struct{}

type Registry struct {
	components map[interface{}]Component

	Service micro.Service
	Client  client.Client

	Logrus LogrusComponent
}

type HealthInfo struct {
	Message string
	IsError bool
}

type HealthInfoMap map[string]HealthInfo

func RegistryMicroHdlWrapper(reg *Registry) func(server.HandlerFunc) server.HandlerFunc {
	return func(in server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			ctx = context.WithValue(ctx, RegistryKey{}, reg)
			return in(ctx, req, rsp)
		}
	}
}

func RegistryFromContext(ctx context.Context) (*Registry, error) {
	reg, ok := ctx.Value(RegistryKey{}).(*Registry)
	if !ok {
		return nil, errorRetrievingRegistry
	}

	if reg == nil {
		return nil, errorRegistryIsNil
	}

	return reg, nil
}

func RegistryToContext(ctx context.Context, reg *Registry) context.Context {
	return context.WithValue(ctx, RegistryKey{}, reg)
}

func NewRegistry(components ...Component) *Registry {
	reg := &Registry{components: make(map[interface{}]Component)}

	reg.Add(components...)

	return reg
}

func (r *Registry) SetService(service micro.Service) {
	r.Service = service
	r.Client = service.Client()
}

func (r *Registry) Add(components ...Component) {
	for _, component := range components {
		if component == nil {
			continue
		}

		if _, ok := r.components[component.Key()]; ok {
			continue
		}

		r.components[component.Key()] = component
	}
}

func (r *Registry) Get(key interface{}) (Component, error) {
	if c, ok := r.components[key]; ok {
		return c, nil
	}

	return nil, errors.New("not found")
}

func (r *Registry) Flags() []cli.Flag {
	flags := []cli.Flag{}
	for _, c := range r.components {
		flags = append(flags, c.Flags()...)
	}

	return flags
}

func (r *Registry) Initialized() bool {
	for _, c := range r.components {
		if !c.Initialized() {
			return false
		}
	}

	return true
}

func (r *Registry) Init(context *cli.Context) error {
	// Sort Components by Priority ASC
	components := make([]Component, len(r.components))
	for _, com := range r.components {
		components = append(components, com)
	}
	sort.Slice(components, func(i, j int) bool {
		if components[i] == nil || components[j] == nil {
			return false
		}

		return components[i].Priority() < components[j].Priority()
	})

	// Init them sorted now
	for _, c := range components {
		if c == nil {
			continue
		}

		if err := c.Init(r, context); err != nil {
			return err
		}
	}

	return nil
}

func (r *Registry) Health(context context.Context) HealthInfoMap {
	result := make(HealthInfoMap, len(r.components))

	for _, c := range r.components {
		m, e := c.Health(context)
		result[c.Name()] = HealthInfo{Message: m, IsError: e}
	}

	return result
}
