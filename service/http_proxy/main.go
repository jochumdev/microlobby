package main

import (
	"log"

	microWeb "go-micro.dev/v4/web"

	"github.com/gin-gonic/gin"
	ginlogrus "github.com/toorop/gin-logrus"
	"github.com/urfave/cli/v2"

	"wz2100.net/microlobby/service/http_proxy/version"
	"wz2100.net/microlobby/service/http_proxy/web"
	"wz2100.net/microlobby/service/http_proxy/web/proxy"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	_ "wz2100.net/microlobby/shared/micro_plugins"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut())

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	service := microWeb.NewService(
		microWeb.Name(defs.ServiceHttpProxy),
		microWeb.Version(version.Version),
		microWeb.Handler(router),
		microWeb.Flags(registry.Flags()...),
	)

	if err := service.Init(microWeb.Action(func(c *cli.Context) {
		if err := registry.Init(c); err != nil {
			log.Fatal(err)
			return
		}

		logrusc, err := component.Logrus(registry)
		if err != nil {
			log.Fatal(err)
			return
		}

		router.Use(ginlogrus.Logger(logrusc.Logger()), gin.Recovery())

		web.ConfigureRouter(registry, router)
		proxy.ConfigureRouter(registry, router)
	})); err != nil {
		log.Fatal(err)
	}

	// Run server
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
