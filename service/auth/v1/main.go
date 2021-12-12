package main

import (
	"log"
	"net/http"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	authService "wz2100.net/microlobby/service/auth/v1/service/auth_service"
	"wz2100.net/microlobby/service/auth/v1/version"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/infoservice"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/authservicepb/v1"
	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
	"wz2100.net/microlobby/shared/utils"
)

const pkgPath = "wz2100.net/microlobby/service/auth/v1"

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), component.NewBUN())

	service := micro.NewService(
		micro.Name(defs.ServiceAuthV1),
		micro.Version(version.Version),
		micro.Flags(registry.Flags()...),
		micro.WrapHandler(component.BunMicroHdlWrapper(registry)),
	)

	routes := []*infoservicepb.RoutesReply_Route{
		{
			Method:      http.MethodGet,
			Path:        "/user",
			Endpoint:    utils.ReflectFunctionName(authservicepb.AuthService.UserList),
			RequireRole: auth.ROLE_ADMIN,
			Params:      []string{"limit", "offset"},
		},
		{
			Method:   http.MethodDelete,
			Path:     "/user/:userId",
			Endpoint: utils.ReflectFunctionName(authservicepb.AuthService.UserDelete),
			Params:   []string{"userId"},
		},
		{
			Method:   http.MethodGet,
			Path:     "/user/:userId",
			Endpoint: utils.ReflectFunctionName(authservicepb.AuthService.UserDetail),
			Params:   []string{"userId"},
		},
		{
			Method:      http.MethodPut,
			Path:        "/user/:userId/roles",
			Endpoint:    utils.ReflectFunctionName(authservicepb.AuthService.UserUpdateRoles),
			RequireRole: auth.ROLE_SUPERADMIN,
			Params:      []string{"userId"},
		},
		{
			Method:   http.MethodPost,
			Path:     "/login",
			Endpoint: utils.ReflectFunctionName(authservicepb.AuthService.Login),
		},
		{
			Method:   http.MethodPost,
			Path:     "/logout",
			Endpoint: utils.ReflectFunctionName(authservicepb.AuthService.Logout),
		},
		{
			Method:      http.MethodGet,
			Path:        "/token",
			Endpoint:    utils.ReflectFunctionName(authservicepb.AuthService.TokenList),
			RequireRole: auth.ROLE_ADMIN,
		},
		{
			Method:   http.MethodGet,
			Path:     "/token/:token",
			Endpoint: utils.ReflectFunctionName(authservicepb.AuthService.TokenDetail),
			Params:   []string{"token"},
		},
		{
			Method:   http.MethodPut,
			Path:     "/token/:token",
			Endpoint: utils.ReflectFunctionName(authservicepb.AuthService.TokenRefresh),
			Params:   []string{"token"},
		},
		{
			Method:   http.MethodDelete,
			Path:     "/token/:token",
			Endpoint: utils.ReflectFunctionName(authservicepb.AuthService.TokenDelete),
			Params:   []string{"token"},
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
			infoService := infoservice.NewHandler(registry, defs.ProxyURIAuth, "v1", routes)
			infoservicepb.RegisterInfoServiceHandler(s, infoService)

			authH, err := authService.NewHandler()
			if err != nil {
				logrus.WithFunc(pkgPath, "main").Fatal(err)
				return err
			}
			authservicepb.RegisterAuthServiceHandler(s, authH)

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		log.Fatalln(err)
	}
}
