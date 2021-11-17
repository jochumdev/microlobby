package web

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	"go-micro.dev/v4/cmd"
	"wz2100.net/microlobby/shared/serviceregistry"

	infoServiceProto "wz2100.net/microlobby/shared/proto/infoservice"
)

// const pkgPath = "wz2100.net/microlobby/service_main/web"

type Handler struct{}

func ConfigureRouter(r *gin.Engine) {
	h := &Handler{}

	r.GET("/health", h.getHealth)
}

func (h *Handler) getHealth(c *gin.Context) {
	allFine := true

	services, err := serviceregistry.ServiceFindByEndpoint("InfoService.Health", *cmd.DefaultOptions().Registry, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  500,
			"message": "We see errors",
		})
		return
	}

	for _, s := range services {
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
