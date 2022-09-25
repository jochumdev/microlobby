package settings

import (
	"context"
	"fmt"
	"sync"

	"go-micro.dev/v4/errors"

	"github.com/urfave/cli/v2"
	"jochum.dev/jo-micro/components"
	"wz2100.net/microlobby/service/settings/v1/config"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
	"wz2100.net/microlobby/shared/utils"
)

const Name = "settingsV1client"

type Handler struct {
	initialized   bool
	cReg          *components.Registry
	cacheGetLock  *sync.RWMutex
	cacheGet      map[string]*settingsservicepb.Setting
	cacheListLock *sync.RWMutex
	cacheList     map[string][]*settingsservicepb.Setting
}

func MustReg(cReg *components.Registry) *Handler {
	return cReg.Must(Name).(*Handler)
}

func (h *Handler) sClient() (settingsservicepb.SettingsV1Service, error) {
	// Wait until the service is here
	_, err := utils.ServiceRetryGet(h.cReg.Service(), config.Name, 10)
	if err != nil {
		return nil, err
	}

	service := settingsservicepb.NewSettingsV1Service(config.Name, h.cReg.Service().Client())
	return service, nil

}

// NewLog creates a new component
func New() *Handler {
	return &Handler{
		initialized:   false,
		cacheGetLock:  &sync.RWMutex{},
		cacheGet:      make(map[string]*settingsservicepb.Setting),
		cacheListLock: &sync.RWMutex{},
		cacheList:     make(map[string][]*settingsservicepb.Setting),
	}
}

func (c *Handler) Priority() int {
	return 30
}

func (c *Handler) Name() string {
	return Name
}

func (c *Handler) Flags(cReg *components.Registry) []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:  "settings_cachetime",
			Usage: "Time in seconds where settingsV1 caches your request",
			Value: 3600,
		},
	}
}

func (c *Handler) Initialized() bool {
	return c.initialized
}

func (h *Handler) Init(cReg *components.Registry, cli *cli.Context) error {
	if h.initialized {
		return nil
	}

	h.cReg = cReg

	h.initialized = true
	return nil
}

func (h *Handler) Stop() error {
	h.initialized = false
	return nil
}

func (c *Handler) Health(context context.Context) error {
	if !c.Initialized() {
		return errors.InternalServerError("NOT_INITIALIZED", "Not initialized")
	}

	return nil
}

func (c *Handler) Get(ctx context.Context, id, ownerId, service, name string) (*settingsservicepb.Setting, error) {
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
		return nil, errors.BadRequest("INVALID_ARGUMENTS", "invalid arguments")
	}

	// Check cache and return from cache
	c.cacheGetLock.RLock()
	if result, ok := c.cacheGet[cacheKey]; ok {
		c.cacheGetLock.RUnlock()
		return result, nil
	}
	c.cacheGetLock.RUnlock()

	client, err := c.sClient()
	if err != nil {
		return nil, err
	}

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

func (c *Handler) List(ctx context.Context, id, ownerId, service, name string) ([]*settingsservicepb.Setting, error) {
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
		return nil, errors.BadRequest("INVALID_ARGUMENTS", "invalid arguments")
	}

	// Check cache and return from cache
	c.cacheListLock.RLock()
	if result, ok := c.cacheList[cacheKey]; ok {
		c.cacheListLock.RUnlock()
		return result, nil
	}
	c.cacheListLock.RUnlock()

	// Fetch
	client, err := c.sClient()
	if err != nil {
		return nil, err
	}
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

func (c *Handler) Create(ctx context.Context, req *settingsservicepb.CreateRequest) (*settingsservicepb.Setting, error) {
	// Create
	client, err := c.sClient()
	if err != nil {
		return nil, err
	}
	return client.Create(ctx, req)
}

func (c *Handler) Update(ctx context.Context, req *settingsservicepb.UpdateRequest) (*settingsservicepb.Setting, error) {
	// Update
	client, err := c.sClient()
	if err != nil {
		return nil, err
	}
	return client.Update(ctx, req)
}

func (c *Handler) Upsert(ctx context.Context, req *settingsservicepb.UpsertRequest) (*settingsservicepb.Setting, error) {
	// Upsert
	client, err := c.sClient()
	if err != nil {
		return nil, err
	}
	return client.Upsert(ctx, req)
}
