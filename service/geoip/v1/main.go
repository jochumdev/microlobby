package main

import (
	"log"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/logger"
	"jochum.dev/jo-micro/auth2"
	"jochum.dev/jo-micro/auth2/plugins/verifier/endpointroles"
	"jochum.dev/jo-micro/router"
	"wz2100.net/microlobby/service/geoip/v1/config"
	"wz2100.net/microlobby/service/geoip/v1/handler/geoip"
	"wz2100.net/microlobby/service/geoip/v1/proto/geoippb"
	scomponent "wz2100.net/microlobby/service/settings/component"
	"wz2100.net/microlobby/shared/component"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/utils"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), scomponent.NewSettingsV1())

	auth2ClientReg := auth2.ClientAuthRegistry()

	service := micro.NewService(
		micro.Name(config.Name),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Version(config.Version),
		micro.Flags(auth2ClientReg.MergeFlags(registry.MergeFlags([]cli.Flag{
			&cli.IntFlag{
				Name:    "geoip_refreshdb",
				Usage:   "Refresh the DB every x seconds, default is every half day",
				EnvVars: []string{"GEOIP_REFRESHDB"},
				Value:   43200,
			},
			&cli.StringFlag{
				Name:    "geoip_data_directory",
				Usage:   "GeoIP data directory",
				EnvVars: []string{"GEOIP_DATA_DIRECTORY"},
				Value:   "/data",
			},
			&cli.StringFlag{
				Name:    "geoip_maxmind_account_id",
				Usage:   "Maxmind account ID",
				EnvVars: []string{"GEOIP_MAXMIND_ACCOUNT_ID"},
			},
			&cli.StringFlag{
				Name:    "geoip_maxmind_license_key",
				Usage:   "Maxmind license key",
				EnvVars: []string{"GEOIP_MAXMIND_LICENSE_KEY"},
			},
		}))...),
		micro.WrapHandler(auth2ClientReg.Wrapper()),
	)
	registry.SetService(service)

	geoipH, err := geoip.NewHandler(registry)
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

			if c.String("geoip_maxmind_account_id") == "" || c.String("geoip_maxmind_license_key") == "" {
				err := errors.InternalServerError("NO_MAXMIND_ACCOUNT", "you have to provide geoip_maxmind_account_id and geoip_maxmind_license_key")
				cLogrus.Logger().Fatal(err)
				return err
			}

			if err := utils.IsDirectoryAndWriteable(c.String("geoip_data_directory")); err != nil {
				cLogrus.Logger().Fatal(err)
				return err
			}

			if err := auth2ClientReg.Init(
				auth2.CliContext(c),
				auth2.Service(service),
				auth2.Logrus(cLogrus.Logger()),
			); err != nil {
				cLogrus.Logger().Fatal(err)
				return err
			}

			authVerifier := endpointroles.NewVerifier(
				endpointroles.WithLogrus(cLogrus.Logger()),
			)
			authVerifier.AddRules(
				endpointroles.RouterRule,
				endpointroles.NewRule(
					endpointroles.Endpoint(geoippb.GeoIPV1Service.Country),
					endpointroles.RolesAllow(auth2.RolesServiceAndAdmin),
				),
				endpointroles.NewRule(
					endpointroles.Endpoint(geoippb.GeoIPV1Service.City),
					endpointroles.RolesAllow(auth2.RolesServiceAndAdmin),
				),
			)
			auth2ClientReg.Plugin().SetVerifier(authVerifier)

			s := service.Server()
			r := router.NewHandler(
				config.RouterURI,
				router.NewRoute(
					router.Method(router.MethodGet),
					router.Path("/country/:ip"),
					router.Endpoint(geoippb.GeoIPV1Service.Country),
					router.Params("ip"),
					router.AuthRequired(),
					router.RatelimitUser("1-S", "10-M"),
				),
				router.NewRoute(
					router.Method(router.MethodGet),
					router.Path("/country/:ip/:commaLanguages"),
					router.Endpoint(geoippb.GeoIPV1Service.Country),
					router.Params("ip", "commaLanguages"),
					router.AuthRequired(),
					router.RatelimitUser("1-S", "10-M"),
				),
				router.NewRoute(
					router.Method(router.MethodGet),
					router.Path("/city/:ip"),
					router.Endpoint(geoippb.GeoIPV1Service.City),
					router.Params("ip"),
					router.AuthRequired(),
					router.RatelimitUser("1-S", "10-M"),
				),
				router.NewRoute(
					router.Method(router.MethodGet),
					router.Path("/city/:ip/:commaLanguages"),
					router.Endpoint(geoippb.GeoIPV1Service.City),
					router.Params("ip", "commaLanguages"),
					router.AuthRequired(),
					router.RatelimitUser("1-S", "10-M"),
				),
			)
			r.RegisterWithServer(s)

			if err := geoipH.Start(geoip.Config{
				DataDirectory:  c.String("geoip_data_directory"),
				RefreshSeconds: c.Int("geoip_refreshdb"),
				AccountID:      c.String("geoip_maxmin_account_id"),
				LicenseKey:     c.String("geoip_maxmind_license_key"),
			}); err != nil {
				cLogrus.Logger().Fatal(err)
				return err
			}
			geoippb.RegisterGeoIPV1ServiceHandler(s, geoipH)

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		log.Fatalln(err)
	}

	if err := geoipH.Stop(); err != nil {
		log.Fatalln(err)
	}

	if err := auth2ClientReg.Stop(); err != nil {
		logger.Fatal(err)
	}
}
