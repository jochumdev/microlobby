package main

import (
	"log"
	"net/http"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	settingsHandler "wz2100.net/microlobby/service/settings/v1/handler/settings"
	"wz2100.net/microlobby/service/settings/v1/version"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/infoservice"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
	"wz2100.net/microlobby/shared/utils"
)

const pkgPath = "wz2100.net/microlobby/service/settings/v1"

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), component.NewBUN())

	service := micro.NewService(
		micro.Name(defs.ServiceSettingsV1),
		micro.Version(version.Version),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Flags(registry.Flags()...),
		micro.WrapHandler(component.RegistryMicroHdlWrapper(registry)),
	)
	registry.SetService(service)

	routes := []*infoservicepb.RoutesReply_Route{
		{
			Method:          http.MethodGet,
			Path:            "/",
			Endpoint:        utils.ReflectFunctionName(settingsservicepb.SettingsV1Service.List),
			IntersectsRoles: []string{auth.ROLE_USER, auth.ROLE_SERVICE},
			Params:          []string{"id", "ownerId", "service", "name", "limit", "offset"},
		},
		{
			Method:          http.MethodPost,
			Path:            "/",
			Endpoint:        utils.ReflectFunctionName(settingsservicepb.SettingsV1Service.Create),
			IntersectsRoles: []string{auth.ROLE_ADMIN, auth.ROLE_SERVICE},
		},
		{
			Method:          http.MethodGet,
			Path:            "/:id",
			Endpoint:        utils.ReflectFunctionName(settingsservicepb.SettingsV1Service.Get),
			IntersectsRoles: []string{auth.ROLE_USER, auth.ROLE_SERVICE},
			Params:          []string{"id", "ownerId", "service", "name"},
		},
		{
			Method:          http.MethodPut,
			Path:            "/:id",
			Endpoint:        utils.ReflectFunctionName(settingsservicepb.SettingsV1Service.Update),
			IntersectsRoles: []string{auth.ROLE_USER, auth.ROLE_SERVICE},
			Params:          []string{"id"},
		},
	}

	service.Init(
		micro.Action(func(c *cli.Context) error {
			if err := registry.Init(c); err != nil {
				return err
			}

			logrus, err := component.Logrus(registry)
			if err != nil {
				log.Fatal(err)
				return err
			}

			s := service.Server()
			infoService := infoservice.NewHandler(registry, defs.ProxyURISettings, "v1", routes)
			infoservicepb.RegisterInfoServiceHandler(s, infoService)

			settingsH, err := settingsHandler.NewHandler()
			if err != nil {
				logrus.WithFunc(pkgPath, "main").Fatal(err)
				return err
			}
			settingsservicepb.RegisterSettingsV1ServiceHandler(s, settingsH)

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		log.Fatalln(err)
	}
}
