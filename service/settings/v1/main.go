package main

import (
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"jochum.dev/jo-micro/auth2"
	"jochum.dev/jo-micro/auth2/plugins/verifier/endpointroles"
	"jochum.dev/jo-micro/router"
	"wz2100.net/microlobby/service/settings/v1/config"
	settingsHandler "wz2100.net/microlobby/service/settings/v1/handler/settings"
	"wz2100.net/microlobby/shared/component"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
)

const pkgPath = "wz2100.net/microlobby/service/settings/v1"

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), component.NewBUN())

	auth2ClientReg := auth2.ClientAuthRegistry()

	service := micro.NewService(
		micro.Name(config.Name),
		micro.Version(config.Version),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Flags(auth2ClientReg.MergeFlags(registry.Flags())...),
	)
	registry.SetService(service)

	service.Init(
		micro.WrapHandler(component.RegistryMicroHdlWrapper(registry), auth2ClientReg.Wrapper()),
		micro.Action(func(c *cli.Context) error {
			if err := registry.Init(c); err != nil {
				return err
			}

			cLogrus, err := component.Logrus(registry)
			if err != nil {
				logger.Fatal(err)
			}

			if err := auth2ClientReg.Init(c, service); err != nil {
				cLogrus.Logger().Fatal(err)
			}

			authVerifier := endpointroles.NewVerifier(
				endpointroles.WithLogrus(cLogrus.Logger()),
			)
			authVerifier.AddRules(
				endpointroles.RouterRule,
				endpointroles.NewRule(
					endpointroles.Endpoint(settingsservicepb.SettingsV1Service.List),
					endpointroles.RolesAllow(auth2.RolesServiceAndUsersAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(settingsservicepb.SettingsV1Service.Create),
					endpointroles.RolesAllow(auth2.RolesServiceAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(settingsservicepb.SettingsV1Service.Get),
					endpointroles.RolesAllow(auth2.RolesServiceAndUsersAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(settingsservicepb.SettingsV1Service.Update),
					endpointroles.RolesAllow(auth2.RolesServiceAndUsersAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(settingsservicepb.SettingsV1Service.Upsert),
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
					router.Endpoint(settingsservicepb.SettingsV1Service.List),
					router.Params("id", "ownerId", "service", "name", "limit", "offset"),
				),
				router.NewRoute(
					router.Method(router.MethodPost),
					router.Path("/"),
					router.Endpoint(settingsservicepb.SettingsV1Service.Create),
				),
				router.NewRoute(
					router.Method(router.MethodGet),
					router.Path("/:id"),
					router.Endpoint(settingsservicepb.SettingsV1Service.Get),
					router.Params("id", "ownerId", "service", "name"),
				),
				router.NewRoute(
					router.Method(router.MethodPut),
					router.Path("/:id"),
					router.Endpoint(settingsservicepb.SettingsV1Service.Update),
					router.Params("id"),
				),
			)
			r.RegisterWithServer(service.Server())

			settingsH, err := settingsHandler.NewHandler()
			if err != nil {
				cLogrus.WithFunc(pkgPath, "main").Fatal(err)
				return err
			}
			settingsservicepb.RegisterSettingsV1ServiceHandler(service.Server(), settingsH)

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		logger.Fatal(err)
	}

	if err := auth2ClientReg.Stop(); err != nil {
		logger.Fatal(err)
	}
}
