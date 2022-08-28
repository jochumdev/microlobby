package main

import (
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"

	"github.com/gin-gonic/gin"
	"github.com/go-micro/plugins/v4/server/http"
	ginlogrus "github.com/toorop/gin-logrus"
	"wz2100.net/microlobby/service/http_proxy/config"
	"wz2100.net/microlobby/service/http_proxy/web"
	"wz2100.net/microlobby/service/http_proxy/web/proxy"
	scomponent "wz2100.net/microlobby/service/settings/v1/component"
	"wz2100.net/microlobby/shared/component"
	_ "wz2100.net/microlobby/shared/micro_plugins"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), scomponent.NewSettingsV1())

	srv := micro.NewService(
		micro.Name(config.Name),
		micro.Version(config.Version),
		micro.Server(http.NewServer(server.Name(config.Name))),
	)

	opts := []micro.Option{
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Flags(registry.Flags()...),
		micro.Address(":8080"),
		micro.Action(func(c *cli.Context) error {
			gin.SetMode(gin.ReleaseMode)

			registry.SetService(srv)
			if err := registry.Init(c); err != nil {
				return err
			}

			logrusc, err := component.Logrus(registry)
			if err != nil {
				return err
			}

			router := gin.New()
			router.Use(ginlogrus.Logger(logrusc.Logger()), gin.Recovery())
			web.ConfigureRouter(registry, router)
			proxy.ConfigureRouter(registry, router)

			if err := micro.RegisterHandler(srv.Server(), router); err != nil {
				logger.Fatal(err)
			}

			return nil
		}),
	}
	srv.Init(opts...)

	// Run server
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
