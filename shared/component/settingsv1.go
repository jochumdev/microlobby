package component

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4/cmd"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
	"wz2100.net/microlobby/shared/utils"
)

type SettingsV1Key struct{}

type SettingsV1Handler struct {
	initialized   bool
	cacheGetLock  *sync.RWMutex
	cacheGet      map[string]*settingsservicepb.Setting
	cacheListLock *sync.RWMutex
	cacheList     map[string][]*settingsservicepb.Setting
}

func SettingsV1FromContext(ctx context.Context) (*SettingsV1Handler, error) {
	reg, err := RegistryFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s, err := SettingsV1(reg)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func SettingsV1(reg *Registry) (*SettingsV1Handler, error) {
	sc, err := reg.Get(SettingsV1Key{})
	if err != nil {
		return nil, err
	}

	result := sc.(*SettingsV1Handler)
	if result == nil {
		return nil, errors.New("settingsv1 is nil")
	}

	return result, nil
}

// NewLog creates a new component
func NewSettingsV1() *SettingsV1Handler {
	return &SettingsV1Handler{
		initialized:   false,
		cacheGetLock:  &sync.RWMutex{},
		cacheGet:      make(map[string]*settingsservicepb.Setting),
		cacheListLock: &sync.RWMutex{},
		cacheList:     make(map[string][]*settingsservicepb.Setting),
	}
}

func (c *SettingsV1Handler) Priority() int8 {
	return 30
}

func (c *SettingsV1Handler) Key() interface{} {
	return SettingsV1Key{}
}

func (c *SettingsV1Handler) Name() string {
	return "shared.settingsv1"
}

func (c *SettingsV1Handler) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:  "settings-cachetime",
			Value: 3600,
			Usage: "Time in seconds where settingsV1 caches your request",
		},
	}
}

func (c *SettingsV1Handler) Initialized() bool {
	return c.initialized
}

func (c *SettingsV1Handler) Init(registry *Registry, cli *cli.Context) error {
	if c.initialized {
		return nil
	}

	c.initialized = true
	return nil
}

func (c *SettingsV1Handler) Health(context context.Context) (string, bool) {
	if !c.Initialized() {
		return "Not initialized", true
	}

	return "All fine", false
}

func (c *SettingsV1Handler) Get(ctx context.Context, id, ownerId, service, name string) (*settingsservicepb.Setting, error) {
	// Build the request
	req := &settingsservicepb.GetRequest{}
	cacheKey := ""
	if len(id) > 0 {
		req.Id = id
		cacheKey = id
	} else if len(ownerId) > 0 {
		req.OwnerId = ownerId
		if len(name) > 0 {
			req.Name = name
			cacheKey = fmt.Sprintf("%s-%s", req.OwnerId, req.Name)
		} else {
			cacheKey = req.OwnerId
		}
	} else if len(service) > 0 {
		req.Service = service
		if len(name) > 0 {
			req.Name = name
			cacheKey = fmt.Sprintf("%s-%s", req.Service, req.Name)
		} else {
			cacheKey = req.Service
		}
	} else {
		return nil, errors.New("invalid arguments")
	}

	// Check cache and return from cache
	c.cacheGetLock.RLock()
	if result, ok := c.cacheGet[cacheKey]; ok {
		c.cacheGetLock.RUnlock()
		return result, nil
	}
	c.cacheGetLock.RUnlock()

	client := settingsservicepb.NewSettingsV1Service(defs.ServiceSettingsV1, *cmd.DefaultOptions().Client)
	result, err := client.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", cacheKey, err)
	}

	// Store the result in cache
	c.cacheGetLock.Lock()
	c.cacheGet[cacheKey] = result
	c.cacheGetLock.Unlock()

	return result, nil
}

func (c *SettingsV1Handler) List(ctx context.Context, id, ownerId, service, name string) ([]*settingsservicepb.Setting, error) {
	// Build the request
	req := &settingsservicepb.ListRequest{}
	cacheKey := ""
	if len(id) > 0 {
		req.Id = id
		cacheKey = id
	} else if len(service) > 0 {
		req.Service = service
		if len(name) > 0 {
			req.Name = name
			cacheKey = fmt.Sprintf("%s-%s", req.Service, req.Name)
		} else {
			cacheKey = req.Service
		}
	} else if len(ownerId) > 0 {
		req.OwnerId = ownerId
		if len(name) > 0 {
			req.Name = name
			cacheKey = fmt.Sprintf("%s-%s", req.OwnerId, req.Name)
		} else {
			cacheKey = req.OwnerId
		}
	} else {
		return nil, errors.New("invalid arguments")
	}

	// Check cache and return from cache
	c.cacheListLock.RLock()
	if result, ok := c.cacheList[cacheKey]; ok {
		c.cacheListLock.RUnlock()
		return result, nil
	}
	c.cacheListLock.RUnlock()

	// Wait until the service is here
	_, err := utils.ServiceRetryGet(defs.ServiceSettingsV1, 10)
	if err != nil {
		return nil, err
	}

	// Fetch
	client := settingsservicepb.NewSettingsV1Service(defs.ServiceSettingsV1, *cmd.DefaultOptions().Client)
	result, err := client.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// Store the result in cache
	c.cacheListLock.Lock()
	c.cacheList[cacheKey] = result.Data
	c.cacheListLock.Unlock()

	return result.Data, nil
}

func (c *SettingsV1Handler) Create(ctx context.Context, req *settingsservicepb.CreateRequest) (*settingsservicepb.Setting, error) {
	// Wait until the service is here
	_, err := utils.ServiceRetryGet(defs.ServiceSettingsV1, 10)
	if err != nil {
		return nil, err
	}

	// Create
	client := settingsservicepb.NewSettingsV1Service(defs.ServiceSettingsV1, *cmd.DefaultOptions().Client)
	return client.Create(ctx, req)
}

func (c *SettingsV1Handler) Update(ctx context.Context, req *settingsservicepb.UpdateRequest) (*settingsservicepb.Setting, error) {
	// Wait until the service is here
	_, err := utils.ServiceRetryGet(defs.ServiceSettingsV1, 10)
	if err != nil {
		return nil, err
	}

	// Update
	client := settingsservicepb.NewSettingsV1Service(defs.ServiceSettingsV1, *cmd.DefaultOptions().Client)
	return client.Update(ctx, req)
}

func (c *SettingsV1Handler) Upsert(ctx context.Context, req *settingsservicepb.CreateRequest) (*settingsservicepb.Setting, error) {
	// Wait until the service is here
	_, err := utils.ServiceRetryGet(defs.ServiceSettingsV1, 10)
	if err != nil {
		return nil, err
	}

	// Upsert
	client := settingsservicepb.NewSettingsV1Service(defs.ServiceSettingsV1, *cmd.DefaultOptions().Client)
	if s, err := c.Get(ctx, "", req.OwnerId, req.Service, req.Name); err == nil {
		// Update
		ureq := &settingsservicepb.UpdateRequest{Id: s.Id, Content: req.Content}
		return client.Update(ctx, ureq)
	}

	// Create
	return client.Create(ctx, req)
}
