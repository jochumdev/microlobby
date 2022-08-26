package main

import (
	"log"
	"net/http"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	authSvc "wz2100.net/microlobby/service/badwords/v1/service/auth"
	bwSvc "wz2100.net/microlobby/service/badwords/v1/service/badwords"
	"wz2100.net/microlobby/service/badwords/v1/version"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/infoservice"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/authservicepb/v1"
	"wz2100.net/microlobby/shared/proto/badwordspb/v1"
	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
	"wz2100.net/microlobby/shared/utils"
)

const pkgPath = version.PkgPath

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), component.NewSettingsV1())

	service := micro.NewService(
		micro.Name(defs.ServiceBadwordsV1),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Version(version.Version),
		micro.Flags(registry.Flags()...),
		micro.WrapHandler(component.RegistryMicroHdlWrapper(registry)),
	)
	registry.SetService(service)

	routes := []*infoservicepb.RoutesReply_Route{
		{
			Method:          http.MethodGet,
			Path:            "/check/:request",
			Endpoint:        utils.ReflectFunctionName(badwordspb.BadwordsV1Service.Check),
			Params:          []string{"request"},
			IntersectsRoles: []string{auth.ROLE_ADMIN, auth.ROLE_SERVICE},
		},
		{
			Method:          http.MethodPost,
			Path:            "/check",
			Endpoint:        utils.ReflectFunctionName(badwordspb.BadwordsV1Service.Check),
			IntersectsRoles: []string{auth.ROLE_ADMIN, auth.ROLE_SERVICE},
		},
	}

	authH, err := authSvc.NewHandler(registry)
	if err != nil {
		log.Fatalln(err)
	}
	bwH, err := bwSvc.NewHandler(registry)
	if err != nil {
		log.Fatalln(err)
	}

	service.Init(
		micro.Action(func(c *cli.Context) error {
			if err := registry.Init(c); err != nil {
				return err
			}

			s := service.Server()
			infoService := infoservice.NewHandler(registry, defs.ProxyURIBadwords, "v1", routes)
			infoservicepb.RegisterInfoServiceHandler(s, infoService)

			if err := authH.Start(); err != nil {
				log.Fatalln(err)
			}
			authservicepb.RegisterAuthV1PreServiceHandler(s, authH)

			if err := bwH.Start(); err != nil {
				log.Fatalln(err)
			}
			badwordspb.RegisterBadwordsV1ServiceHandler(s, bwH)

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		log.Fatalln(err)
	}

	if err := authH.Stop(); err != nil {
		log.Fatalln(err)
	}
	if err := bwH.Stop(); err != nil {
		log.Fatalln(err)
	}
}