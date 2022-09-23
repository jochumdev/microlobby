package main

import (
	"log"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"jochum.dev/jo-micro/auth2"
	"jochum.dev/jo-micro/auth2/plugins/verifier/endpointroles"
	"jochum.dev/jo-micro/router"
	"wz2100.net/microlobby/service/badwords/v1/config"
	bwHandler "wz2100.net/microlobby/service/badwords/v1/handler/badwords"
	scomponent "wz2100.net/microlobby/service/settings/component"
	"wz2100.net/microlobby/shared/component"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/badwordspb/v1"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), scomponent.NewSettingsV1())

	auth2ClientReg := auth2.ClientAuthRegistry()

	service := micro.NewService(
		micro.Name(config.Name),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Version(config.Version),
		micro.Flags(auth2ClientReg.MergeFlags(registry.Flags())...),
		micro.WrapHandler(component.RegistryMicroHdlWrapper(registry), auth2ClientReg.Wrapper()),
	)
	registry.SetService(service)

	bwH, err := bwHandler.NewHandler(registry)
	if err != nil {
		log.Fatalln(err)
	}

	service.Init(
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
				endpointroles.NewRule(
					endpointroles.Endpoint(badwordspb.BadwordsV1Service.Censor),
					endpointroles.RolesAllow(auth2.RolesServiceAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(badwordspb.BadwordsV1Service.Check),
					endpointroles.RolesAllow(auth2.RolesServiceAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(badwordspb.BadwordsV1Service.ExtractProfanity),
					endpointroles.RolesAllow(auth2.RolesServiceAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(badwordspb.BadwordsV1Service.IsProfane),
					endpointroles.RolesAllow(auth2.RolesServiceAndAdmin),
				),
			)
			auth2ClientReg.Plugin().SetVerifier(authVerifier)

			s := service.Server()
			// ab
			r := router.NewHandler(
				config.RouterURI,
				router.NewRoute(
					router.Method(router.MethodGet),
					router.Path("/check/:request"),
					router.Endpoint(badwordspb.BadwordsV1Service.Check),
					router.Params("request"),
				),
				router.NewRoute(
					router.Method(router.MethodPost),
					router.Path("/check"),
					router.Endpoint(badwordspb.BadwordsV1Service.Check),
				),
			)
			r.RegisterWithServer(s)

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

	if err := bwH.Stop(); err != nil {
		log.Fatalln(err)
	}

	if err := auth2ClientReg.Stop(); err != nil {
		logger.Fatal(err)
	}
}
