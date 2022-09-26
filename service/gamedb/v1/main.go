package main

import (
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"jochum.dev/jo-micro/auth2"
	jwtClient "jochum.dev/jo-micro/auth2/plugins/client/jwt"
	"jochum.dev/jo-micro/auth2/plugins/verifier/endpointroles"
	"jochum.dev/jo-micro/buncomponent"
	"jochum.dev/jo-micro/components"
	"jochum.dev/jo-micro/logruscomponent"
	"jochum.dev/jo-micro/router"
	"wz2100.net/microlobby/service/gamedb/v1/config"
	"wz2100.net/microlobby/service/gamedb/v1/gamedbhandler"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/gamedbpb/v1"
)

func main() {
	service := micro.NewService()
	cReg := components.New(
		service,
		"gamedb",
		logruscomponent.New(),
		auth2.ClientAuthComponent(),
		buncomponent.New(),
		gamedbhandler.New(),
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

			authVerifier := endpointroles.NewVerifier(
				endpointroles.WithLogrus(logruscomponent.MustReg(cReg).Logger()),
			)
			authVerifier.AddRules(
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
			auth2ClientReg.Plugin().AddVerifier(authVerifier)

			return nil
		}),
	)

	// Run the server
	if err := service.Run(); err != nil {
		logruscomponent.MustReg(cReg).Logger().Fatal(err)
	}

	// Stop the components
	if err := cReg.Stop(); err != nil {
		logger.Fatal(err)
		return
	}
}
