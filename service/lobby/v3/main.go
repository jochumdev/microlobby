package main

import (
	"log"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	infoHandler "wz2100.net/microlobby/service/http_proxy/handler/info"
	"wz2100.net/microlobby/service/lobby/v3/config"
	lobbyHandler "wz2100.net/microlobby/service/lobby/v3/handler/lobby"
	scomponent "wz2100.net/microlobby/service/settings/v1/component"
	"wz2100.net/microlobby/shared/component"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), scomponent.NewSettingsV1())

	service := micro.NewService(
		micro.Name(config.Name),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Version(config.Version),
		micro.Flags(registry.Flags()...),
	)
	registry.SetService(service)

	routes := []*infoservicepb.RoutesReply_Route{}

	lobbyH, err := lobbyHandler.NewHandler(registry)
	if err != nil {
		log.Fatalln(err)
	}

	service.Init(
		micro.Action(func(c *cli.Context) error {
			if err := registry.Init(c); err != nil {
				return err
			}

			s := service.Server()
			infoService := infoHandler.NewHandler(registry, config.ProxyURI, "v3", routes)
			infoservicepb.RegisterInfoServiceHandler(s, infoService)

			if err := lobbyH.Start(); err != nil {
				log.Fatalln(err)
			}

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		log.Fatalln(err)
	}

	if err := lobbyH.Stop(); err != nil {
		log.Fatalln(err)
	}
}
