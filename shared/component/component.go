package component

import (
	"context"
	"errors"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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

type Registry struct {
	components map[interface{}]Component

	Logrus LogrusComponent
}

type HealthInfo struct {
	Message string
	IsError bool
}

type HealthInfoMap map[string]HealthInfo

func NewRegistry(components ...Component) *Registry {
	reg := &Registry{components: make(map[interface{}]Component)}

	reg.Add(components...)

	return reg
}

func (r *Registry) Add(components ...Component) {
	for _, component := range components {
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
		return components[i].Priority() < components[j].Priority()
	})

	// Init them sorted now
	for _, c := range components {
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
