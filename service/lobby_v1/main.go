package main

import (
	"log"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"wz2100.net/microlobby/service/lobby_v1/version"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/infoservice"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	infoSvcPb "wz2100.net/microlobby/shared/proto/infoservice"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut())

	service := micro.NewService(
		micro.Name(defs.ServiceLobbyV1),
		micro.Version(version.Version),
		micro.Flags(registry.Flags()...),
	)

	routes := []*infoSvcPb.RoutesReply_Route{}

	service.Init(
		micro.Action(func(c *cli.Context) error {
			if err := registry.Init(c); err != nil {
				return err
			}

			s := service.Server()
			infoService := infoservice.NewHandler(registry, "v1", routes)
			infoSvcPb.RegisterInfoServiceHandler(s, infoService)

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		log.Fatalln(err)
	}
}