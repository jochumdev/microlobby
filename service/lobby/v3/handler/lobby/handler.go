package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"go-micro.dev/v4/errors"
	"wz2100.net/microlobby/service/lobby/v3/config"
	scomponent "wz2100.net/microlobby/service/settings/v1/component"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
	"wz2100.net/microlobby/shared/utils"
)

const pkgPath = config.PkgPath + "/handler/lobby"

type Config struct {
	Host string `json:"host"`
	Port int32  `json:"port"`
}

type Handler struct {
	cRegistry *component.Registry
	svcName   string
	logrus    component.LogrusComponent
	config    Config
	listener  net.Listener
}

func NewHandler(cregistry *component.Registry) (*Handler, error) {
	h := &Handler{
		cRegistry: cregistry,
		svcName:   cregistry.Service.Name(),
	}

	return h, nil
}

func (h *Handler) Start() error {
	logrus, err := component.Logrus(h.cRegistry)
	if err != nil {
		return errors.FromError(err)
	}
	h.logrus = logrus

	ctx := component.RegistryToContext(utils.CtxForService(context.Background()), h.cRegistry)
	s, err := scomponent.SettingsV1(h.cRegistry)
	if err != nil {
		return errors.FromError(err)
	}

	var c Config
	se, err := s.Get(ctx, "", "", h.svcName, "config")
	if err == nil {
		err = json.Unmarshal(se.Content, &c)
	} else {
		c.Host = "0.0.0.0"
		c.Port = 9990

		craw, err := json.Marshal(&c)
		if err != nil {
			return errors.FromError(err)
		}

		if _, err = s.Upsert(ctx, &settingsservicepb.UpsertRequest{
			Service:     h.svcName,
			Name:        "config",
			Content:     craw,
			RolesRead:   []string{auth.ROLE_ADMIN, auth.ROLE_SERVICE},
			RolesUpdate: []string{auth.ROLE_ADMIN, auth.ROLE_SERVICE},
		}); err != nil {
			return errors.FromError(err)
		}
	}
	h.config = c

	h.logrus.WithClassFunc(pkgPath, "Handler", "Start").Infof("Lobbyserver listening on: %s:%d", h.config.Host, h.config.Port)
	h.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", h.config.Host, h.config.Port))
	if err != nil {
		return errors.FromError(err)
	}

	go func() {
		for {
			conn, err := h.listener.Accept()
			if err != nil {
				h.logrus.WithClassFunc(pkgPath, "Handler", "listener").Error(err)
				break
			}

			sh, err := NewConnHandler(h.cRegistry, conn)
			if err != nil {
				h.logrus.WithClassFunc(pkgPath, "Handler", "listener").Error(err)
				continue
			}

			go sh.Serve()
		}
	}()

	return nil
}

func (h *Handler) Stop() error {
	if h.listener != nil {
		h.listener.Close()
		h.listener = nil
	}
	return nil
}
