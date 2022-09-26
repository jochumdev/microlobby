package main

import (
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"jochum.dev/jo-micro/auth2"
	jwtClient "jochum.dev/jo-micro/auth2/plugins/client/jwt"
	"jochum.dev/jo-micro/buncomponent"
	"jochum.dev/jo-micro/components"
	"jochum.dev/jo-micro/logruscomponent"
	"jochum.dev/jo-micro/router"
	"wz2100.net/microlobby/service/settings/v1/config"
	"wz2100.net/microlobby/service/settings/v1/settingshandler"
	_ "wz2100.net/microlobby/shared/micro_plugins"
)

func main() {
	service := micro.NewService()
	cReg := components.New(service, "settings",
		logruscomponent.New(),
		auth2.ClientAuthComponent(),
		buncomponent.New(),
		router.New(),
		settingshandler.New(),
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
