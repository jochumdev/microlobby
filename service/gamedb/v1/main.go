package main

import (
	"log"
	"net/http"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"wz2100.net/microlobby/service/gamedb/v1/service/gamedb"
	"wz2100.net/microlobby/service/gamedb/v1/version"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/infoservice"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/gamedbpb/v1"
	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
	"wz2100.net/microlobby/shared/utils"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), component.NewBUN(), component.NewSettingsV1())

	service := micro.NewService(
		micro.Name(defs.ServiceGameDBV1),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Version(version.Version),
		micro.Flags(registry.Flags()...),
	)
	registry.SetService(service)

	routes := []*infoservicepb.RoutesReply_Route{
		{
			Method:          http.MethodGet,
			Path:            "/",
			Endpoint:        utils.ReflectFunctionName(gamedbpb.GameDBV1Service.List),
			IntersectsRoles: auth.AllowServiceAndUsers,
			Params:          []string{"id", "history", "name", "limit", "offset"},
		},
		{
			Method:          http.MethodPost,
			Path:            "/",
			Endpoint:        utils.ReflectFunctionName(gamedbpb.GameDBV1Service.Create),
			IntersectsRoles: auth.AllowServiceAndAdmin,
		},
		// {
		// 	Method:          http.MethodGet,
		// 	Path:            "/:id",
		// 	Endpoint:        utils.ReflectFunctionName(settingsservicepb.SettingsV1Service.Get),
		// 	IntersectsRoles: auth.AllowServiceAndUsers,
		// 	Params:          []string{"id"},
		// },
		{
			Method:          http.MethodPut,
			Path:            "/:id",
			Endpoint:        utils.ReflectFunctionName(gamedbpb.GameDBV1Service.Update),
			IntersectsRoles: auth.AllowServiceAndUsers,
			Params:          []string{"id"},
		},
	}

	gdbH, err := gamedb.NewHandler(registry)
	if err != nil {
		log.Fatalln(err)
	}

	service.Init(
		micro.Action(func(c *cli.Context) error {
			if err := registry.Init(c); err != nil {
				return err
			}

			s := service.Server()
			infoService := infoservice.NewHandler(registry, defs.ProxyURIGameDB, "v1", routes)
			infoservicepb.RegisterInfoServiceHandler(s, infoService)

			if err := gdbH.Start(); err != nil {
				log.Fatalln(err)
			}
			gamedbpb.RegisterGameDBV1ServiceHandler(s, gdbH)

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		log.Fatalln(err)
	}

	if err := gdbH.Stop(); err != nil {
		log.Fatalln(err)
	}
}
