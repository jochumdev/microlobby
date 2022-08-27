package main

import (
	"log"
	"net/http"

	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"wz2100.net/microlobby/service/auth/v1/config"
	authHandler "wz2100.net/microlobby/service/auth/v1/handler/auth"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/infoservice"
	_ "wz2100.net/microlobby/shared/micro_plugins"
	"wz2100.net/microlobby/shared/proto/authservicepb/v1"
	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
	"wz2100.net/microlobby/shared/serviceregistry"
	"wz2100.net/microlobby/shared/utils"
)

func main() {
	registry := component.NewRegistry(component.NewLogrusStdOut(), component.NewBUN(), component.NewSettingsV1())

	service := micro.NewService(
		micro.Name(defs.ServiceAuthV1),
		micro.Client(client.NewClient(client.ContentType("application/grpc+proto"))),
		micro.Version(config.Version),
		micro.Flags(registry.Flags()...),
		micro.WrapHandler(component.RegistryMicroHdlWrapper(registry)),
	)
	registry.SetService(service)

	routes := []*infoservicepb.RoutesReply_Route{
		{
			Method:          http.MethodGet,
			Path:            "/user",
			Endpoint:        utils.ReflectFunctionName(authservicepb.AuthV1Service.UserList),
			RequireRole:     auth.ROLE_ADMIN,
			Params:          []string{"limit", "offset"},
			IntersectsRoles: []string{auth.ROLE_USER, auth.ROLE_SERVICE},
		},
		{
			Method:          http.MethodDelete,
			Path:            "/user/:userId",
			Endpoint:        utils.ReflectFunctionName(authservicepb.AuthV1Service.UserDelete),
			PreEndpoint:     utils.ReflectFunctionName(authservicepb.AuthV1PreService.UserDelete),
			Params:          []string{"userId"},
			IntersectsRoles: []string{auth.ROLE_USER, auth.ROLE_SERVICE},
		},
		{
			Method:          http.MethodGet,
			Path:            "/user/:userId",
			Endpoint:        utils.ReflectFunctionName(authservicepb.AuthV1Service.UserDetail),
			Params:          []string{"userId"},
			IntersectsRoles: []string{auth.ROLE_USER, auth.ROLE_SERVICE},
		},
		{
			Method:          http.MethodPut,
			Path:            "/user/:userId/roles",
			Endpoint:        utils.ReflectFunctionName(authservicepb.AuthV1Service.UserUpdateRoles),
			PreEndpoint:     utils.ReflectFunctionName(authservicepb.AuthV1PreService.UserUpdateRoles),
			IntersectsRoles: []string{auth.ROLE_SUPERADMIN},
			Params:          []string{"userId"},
		},
		{
			Method:      http.MethodPost,
			Path:        "/login",
			Endpoint:    utils.ReflectFunctionName(authservicepb.AuthV1Service.Login),
			PreEndpoint: utils.ReflectFunctionName(authservicepb.AuthV1PreService.Login),
		},
		{
			Method:      http.MethodPost,
			Path:        "/register",
			Endpoint:    utils.ReflectFunctionName(authservicepb.AuthV1Service.Register),
			PreEndpoint: utils.ReflectFunctionName(authservicepb.AuthV1PreService.Register),
		},
		{
			Method:      http.MethodPost,
			Path:        "/refresh",
			Endpoint:    utils.ReflectFunctionName(authservicepb.AuthV1Service.Refresh),
			PreEndpoint: utils.ReflectFunctionName(authservicepb.AuthV1PreService.Refresh),
		},
	}

	authH, err := authHandler.NewHandler(registry)
	if err != nil {
		log.Fatalln(err)
	}

	service.Init(
		micro.WrapHandler(
			serviceregistry.NewHandlerWrapper(
				registry,
				routes,
			),
		),
		micro.Action(func(c *cli.Context) error {
			if err := registry.Init(c); err != nil {
				return err
			}

			s := service.Server()
			infoService := infoservice.NewHandler(registry, defs.ProxyURIAuth, "v1", routes)
			infoservicepb.RegisterInfoServiceHandler(s, infoService)

			if err := authH.Start(); err != nil {
				log.Fatalln(err)
			}
			authservicepb.RegisterAuthV1ServiceHandler(s, authH)

			return nil
		}),
	)

	if err := service.Run(); err != nil {
		log.Fatalln(err)
	}

	if err := authH.Stop(); err != nil {
		log.Fatalln(err)
	}
}
