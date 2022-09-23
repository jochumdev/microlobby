package main

import (
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"jochum.dev/jo-micro/auth2"
	"jochum.dev/jo-micro/auth2/plugins/verifier/endpointroles"
	"wz2100.net/microlobby/service/lobby/v3/config"
	lobbyHandler "wz2100.net/microlobby/service/lobby/v3/handler/lobby"
	scomponent "wz2100.net/microlobby/service/settings/component"
	"wz2100.net/microlobby/shared/component"
	_ "wz2100.net/microlobby/shared/micro_plugins"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), scomponent.NewSettingsV1())

	auth2ClientReg := auth2.ClientAuthRegistry()

	service := micro.NewService(
		micro.Name(config.Name),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Version(config.Version),
		micro.Flags(auth2ClientReg.MergeFlags(registry.Flags())...),
	)
	registry.SetService(service)

	lobbyH, err := lobbyHandler.NewHandler(registry)
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
			}

			if err := auth2ClientReg.Init(auth2.CliContext(c), auth2.Service(service), auth2.Logrus(cLogrus.Logger())); err != nil {
				cLogrus.Logger().Fatal(err)
			}

			authVerifier := endpointroles.NewVerifier(
				endpointroles.WithLogrus(cLogrus.Logger()),
			)
			authVerifier.AddRules(
				endpointroles.RouterRule,
			)
			auth2ClientReg.Plugin().SetVerifier(authVerifier)

			if err := lobbyH.Start(); err != nil {
				logger.Fatal(err)
			}

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		logger.Fatal(err)
	}

	if err := lobbyH.Stop(); err != nil {
		logger.Fatal(err)
	}

	if err := auth2ClientReg.Stop(); err != nil {
		logger.Fatal(err)
	}
}
