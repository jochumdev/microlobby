package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/serviceregistry"

	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
)

// const pkgPath = "wz2100.net/microlobby/service_main/web"

type Handler struct {
	cRegistry *component.Registry
}

func ConfigureRouter(cregistry *component.Registry, r *gin.Engine) {
	h := &Handler{cRegistry: cregistry}

	r.GET("/health", h.getHealth)
}

func (h *Handler) getHealth(c *gin.Context) {
	allFine := true

	services, err := serviceregistry.FindByEndpoint(c, h.cRegistry, "InfoService.Health")
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

		client := infoservicepb.NewInfoService(s.Name, h.cRegistry.Client)
		resp, err := client.Health(context.TODO(), &emptypb.Empty{})

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
