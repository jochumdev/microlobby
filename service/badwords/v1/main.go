package main

import (
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"jochum.dev/jo-micro/auth2"
	jwtClient "jochum.dev/jo-micro/auth2/plugins/client/jwt"
	"jochum.dev/jo-micro/auth2/plugins/verifier/endpointroles"
	"jochum.dev/jo-micro/components"
	"jochum.dev/jo-micro/logruscomponent"
	"jochum.dev/jo-micro/router"
	"wz2100.net/microlobby/service/badwords/v1/badwordshandler"
	"wz2100.net/microlobby/service/badwords/v1/config"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/badwordspb/v1"
)

func main() {
	service := micro.NewService()
	cReg := components.New(
		service,
		"badwords",
		logruscomponent.New(),
		auth2.ClientAuthComponent(),
		badwordshandler.New(),
		router.New(),
	)

	auth2ClientReg := auth2.ClientAuthMustReg(cReg)
	auth2ClientReg.Register(jwtClient.New())

	service.Init(
		micro.Name(config.Name),
		micro.Version(config.Version),
		micro.Flags(cReg.AppendFlags([]cli.Flag{})...),
		micro.WrapHandler(auth2ClientReg.WrapHandler()),
		micro.Action(func(c *cli.Context) error {
			// Start/Init the components
			if err := cReg.Init(c); err != nil {
				logger.Fatal(err)
				return err
			}

			logger := logruscomponent.MustReg(cReg).Logger()

			authVerifier := endpointroles.NewVerifier(
				endpointroles.WithLogrus(logger),
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
			auth2.ClientAuthMustReg(cReg).Plugin().SetVerifier(authVerifier)

			return nil
		}),
	)

	// Run the server
	if err := service.Run(); err != nil {
		logruscomponent.MustReg(cReg).Logger().Fatal(err)
		return
	}

	// Stop the components
	if err := cReg.Stop(); err != nil {
		logger.Fatal(err)
		return
	}
}
