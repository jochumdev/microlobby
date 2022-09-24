package main

import (
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"jochum.dev/jo-micro/auth2"
	"jochum.dev/jo-micro/auth2/plugins/verifier/endpointroles"
	"jochum.dev/jo-micro/router"
	"wz2100.net/microlobby/service/gamedb/v1/config"
	gamedbHandler "wz2100.net/microlobby/service/gamedb/v1/handler/gamedb"
	scomponent "wz2100.net/microlobby/service/settings/component"
	"wz2100.net/microlobby/shared/component"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/gamedbpb/v1"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), component.NewBUN(), scomponent.NewSettingsV1())

	auth2ClientReg := auth2.ClientAuthRegistry()

	service := micro.NewService(
		micro.Name(config.Name),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Version(config.Version),
		micro.Flags(auth2ClientReg.MergeFlags(registry.MergeFlags([]cli.Flag{}))...),
	)
	registry.SetService(service)

	gdbH, err := gamedbHandler.NewHandler(registry)
	if err != nil {
		logger.Fatal(err)
	}

	service.Init(
		micro.WrapHandler(auth2ClientReg.Wrapper()),
		micro.Action(func(c *cli.Context) error {
			if err := registry.Init(c); err != nil {
				return err
			}

			cLogrus, err := component.Logrus(registry)
			if err != nil {
				logger.Fatal(err)
				return err
			}

			if err := auth2ClientReg.Init(auth2.CliContext(c), auth2.Service(service), auth2.Logrus(cLogrus.Logger())); err != nil {
				cLogrus.Logger().Fatal(err)
				return err
			}

			authVerifier := endpointroles.NewVerifier(
				endpointroles.WithLogrus(cLogrus.Logger()),
			)
			authVerifier.AddRules(
				endpointroles.RouterRule,
				endpointroles.NewRule(
					endpointroles.Endpoint(gamedbpb.GameDBV1Service.List),
					endpointroles.RolesAllow(auth2.RolesServiceAndUsersAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(gamedbpb.GameDBV1Service.Create),
					endpointroles.RolesAllow(auth2.RolesServiceAndUsersAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(gamedbpb.GameDBV1Service.Update),
					endpointroles.RolesAllow(auth2.RolesServiceAndUsersAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(gamedbpb.GameDBV1Service.Delete),
					endpointroles.RolesAllow(auth2.RolesServiceAndUsersAndAdmin),
				),
			)
			auth2ClientReg.Plugin().SetVerifier(authVerifier)

			// Register with https://jochum.dev/jo-micro/router
			r := router.NewHandler(
				config.RouterURI,
				router.NewRoute(
					router.Method(router.MethodGet),
					router.Path("/"),
					router.Endpoint(gamedbpb.GameDBV1Service.List),
					router.Params("id", "history", "name", "limit", "offset"),
					router.AuthRequired(),
				),
				router.NewRoute(
					router.Method(router.MethodPost),
					router.Path("/"),
					router.Endpoint(gamedbpb.GameDBV1Service.Create),
					router.AuthRequired(),
				),
				router.NewRoute(
					router.Method(router.MethodPut),
					router.Path("/:id"),
					router.Endpoint(gamedbpb.GameDBV1Service.Update),
					router.Params("id"),
					router.AuthRequired(),
				),
				router.NewRoute(
					router.Method(router.MethodDelete),
					router.Path("/:id"),
					router.Endpoint(gamedbpb.GameDBV1Service.Delete),
					router.Params("id"),
					router.AuthRequired(),
				),
			)
			r.RegisterWithServer(service.Server())

			if err := gdbH.Start(); err != nil {
				cLogrus.Logger().Fatal(err)
				return err
			}
			gamedbpb.RegisterGameDBV1ServiceHandler(service.Server(), gdbH)

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		logger.Fatal(err)
	}

	if err := gdbH.Stop(); err != nil {
		logger.Fatal(err)
	}

	if err := auth2ClientReg.Stop(); err != nil {
		logger.Fatal(err)
	}
}
