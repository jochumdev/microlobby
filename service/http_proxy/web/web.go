package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	"go-micro.dev/v4/cmd"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/serviceregistry"

	infoServiceProto "wz2100.net/microlobby/shared/proto/infoservice"
)

// const pkgPath = "wz2100.net/microlobby/service_main/web"

type Handler struct{}

func ConfigureRouter(cregistry *component.Registry, r *gin.Engine) {
	h := &Handler{}

	r.GET("/health", h.getHealth)
}

func (h *Handler) getHealth(c *gin.Context) {
	allFine := true

	services, err := serviceregistry.ServicesFindByEndpoint("InfoService.Health", *cmd.DefaultOptions().Registry, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  500,
			"message": "We see errors",
		})
		return
	}

	foundServices := make([]string, len(services))

	for _, s := range services {
		foundServices = append(foundServices, s.Name)

		client := infoServiceProto.NewInfoService(s.Name, *cmd.DefaultOptions().Client)
		resp, err := client.Health(context.TODO(), &empty.Empty{})

		if err != nil {
			allFine = false
		} else {
			if resp.GetHasError() {
				allFine = false
			}
		}
	}

	servicesNotThere := []string{}
	for _, svcReq := range defs.ServicesRequired {
		found := false
		for _, svcFound := range foundServices {
			if svcFound == svcReq {
				found = true
				break
			}
		}

		if !found {
			servicesNotThere = append(servicesNotThere, svcReq)
		}
	}

	if len(servicesNotThere) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  500,
			"message": fmt.Sprintf("We see errors, services %s are not there.", servicesNotThere),
		})
		return
	}

	if !allFine {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  500,
			"message": "We see errors",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "Everything is fine",
	})
}
