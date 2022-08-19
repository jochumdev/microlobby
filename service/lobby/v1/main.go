package main

import (
	"log"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"wz2100.net/microlobby/service/lobby/v1/version"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/infoservice"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), component.NewSettingsV1())

	service := micro.NewService(
		micro.Name(defs.ServiceLobbyV1),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Version(version.Version),
		micro.Flags(registry.Flags()...),
	)
	registry.SetService(service)

	routes := []*infoservicepb.RoutesReply_Route{}

	service.Init(
		micro.Action(func(c *cli.Context) error {
			if err := registry.Init(c); err != nil {
				return err
			}

			s := service.Server()
			infoService := infoservice.NewHandler(registry, defs.ProxyURILobby, "v1", routes)
			infoservicepb.RegisterInfoServiceHandler(s, infoService)

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		log.Fatalln(err)
	}
}
