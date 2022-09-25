package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4/errors"
	"jochum.dev/jo-micro/auth2"
	"jochum.dev/jo-micro/components"
	"jochum.dev/jo-micro/logruscomponent"
	"wz2100.net/microlobby/service/settings"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
)

type Config struct {
	Host string `json:"host"`
	Port int32  `json:"port"`
}

const Name = "lobbyV3Handler"

type Handler struct {
	cReg        *components.Registry
	initialized bool

	config   Config
	listener net.Listener
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

	ctx, err := auth2.ClientAuthMustReg(h.cReg).Plugin().ServiceContext(context.Background())
	if err != nil {
		return errors.FromError(err)
	}

	s := settings.MustReg(h.cReg)
	if err != nil {
		return errors.FromError(err)
	}

	var c Config
	se, err := s.Get(ctx, "", "", h.cReg.Service().Name(), "config")
	if err == nil {
		err = json.Unmarshal(se.Content, &c)
		if err != nil {
			return errors.FromError(err)
		}
	} else {
		c.Host = "0.0.0.0"
		c.Port = 9990

		craw, err := json.Marshal(&c)
		if err != nil {
			return errors.FromError(err)
		}

		if _, err = s.Upsert(ctx, &settingsservicepb.UpsertRequest{
			Service:     h.cReg.Service().Name(),
			Name:        "config",
			Content:     craw,
			RolesRead:   []string{auth2.ROLE_ADMIN, auth2.ROLE_SERVICE},
			RolesUpdate: []string{auth2.ROLE_ADMIN, auth2.ROLE_SERVICE},
		}); err != nil {
			return errors.FromError(err)
		}
	}
	h.config = c

	logruscomponent.MustReg(h.cReg).Logger().Infof("Lobbyserver listening on: %s:%d", h.config.Host, h.config.Port)
	h.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", h.config.Host, h.config.Port))
	if err != nil {
		return errors.FromError(err)
	}

	go func() {
		for {
			conn, err := h.listener.Accept()
			if err != nil {
				logruscomponent.MustReg(h.cReg).Logger().Error(err)
				break
			}

			sh, err := NewConnHandler(h.cReg, conn)
			if err != nil {
				logruscomponent.MustReg(h.cReg).Logger().Error(err)
				continue
			}

			go sh.Serve()
		}
	}()

	h.initialized = true
	return nil
}

func (h *Handler) Stop() error {
	if h.listener != nil {
		h.listener.Close()
		h.listener = nil
	}
	return nil
}

func (h *Handler) Flags(r *components.Registry) []cli.Flag {
	return []cli.Flag{}
}

func (h *Handler) Health(context context.Context) error {
	return nil
}
