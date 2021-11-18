package proxy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/cmd"
	"go-micro.dev/v4/errors"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	middlewaregin "wz2100.net/microlobby/shared/middleware/gin"
	"wz2100.net/microlobby/shared/serviceregistry"
	"wz2100.net/microlobby/shared/utils"

	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
)

// const pkgPath = "wz2100.net/microlobby/service_main/web/proxy"

// Handler is the handler for the proxy
type Handler struct {
	registerdServices []string
}

// ConfigureRouter creates a new Handler and configures it on the router
func ConfigureRouter(cregistry *component.Registry, r *gin.Engine) *Handler {
	h := &Handler{}

	proxyAuthGroup := r.Group(fmt.Sprintf("/%s", defs.ProxyURIHttpProxy))
	// proxyAuthGroup.Use(middlewaregin.UserSrvMiddleware(cregistry))
	// proxyAuthGroup.Use(middlewaregin.RequireUserMiddleware(cregistry))

	proxyAuthGroup.GET("/v1/health", h.getHealth)

	globalGroup := r.Group("")
	globalAuthGroup := r.Group("")
	globalAuthGroup.Use(middlewaregin.UserSrvMiddleware(cregistry))
	globalAuthGroup.Use(middlewaregin.RequireUserMiddleware(cregistry))

	// Refresh routes for the proxy every 30 seconds
	go func() {
		ctx := context.Background()

		for {
			services, err := serviceregistry.ServicesFindByEndpoint("InfoService.Routes", *cmd.DefaultOptions().Registry, ctx)
			if err != nil {
				return
			}

			for _, s := range services {
				i := sort.SearchStrings(h.registerdServices, s.Name)
				if i < len(h.registerdServices) && h.registerdServices[i] == s.Name {
					// already registered
					continue
				}

				client := infoservicepb.NewInfoService(s.Name, *cmd.DefaultOptions().Client)
				resp, err := client.Routes(ctx, &empty.Empty{})
				if err != nil {
					// failure in getting routes, silently ignore
					continue
				}

				serviceGroup := r.Group(fmt.Sprintf("/%s/%s", resp.GetProxyURI(), resp.GetApiVersion()))
				serviceAuthGroup := r.Group(fmt.Sprintf("/%s/%s", resp.GetProxyURI(), resp.GetApiVersion()))
				serviceAuthGroup.Use(middlewaregin.UserSrvMiddleware(cregistry))
				serviceAuthGroup.Use(middlewaregin.RequireUserMiddleware(cregistry))

				for _, route := range resp.Routes {
					if route.GetGlobalRoute() {
						if len(route.RequireRole) > 0 || len(route.IntersectsRoles) > 0 {
							globalAuthGroup.Handle(route.GetMethod(), route.GetPath(), h.proxy(s.Name, route))
						} else {
							globalGroup.Handle(route.GetMethod(), route.GetPath(), h.proxy(s.Name, route))
						}
					} else {
						if len(route.RequireRole) > 0 || len(route.IntersectsRoles) > 0 {
							serviceAuthGroup.Handle(route.GetMethod(), route.GetPath(), h.proxy(s.Name, route))
						} else {
							serviceGroup.Handle(route.GetMethod(), route.GetPath(), h.proxy(s.Name, route))
						}
					}
				}

				h.registerdServices = append(h.registerdServices, s.Name)
				sort.Strings(h.registerdServices)
			}

			time.Sleep(30 * time.Second)
		}
	}()

	return h
}

func (h *Handler) proxy(serviceName string, route *infoservicepb.RoutesReply_Route) func(*gin.Context) {
	return func(c *gin.Context) {
		// Check if the user has the required role
		if len(route.RequireRole) > 0 || len(route.IntersectsRoles) > 0 {
			u, err := auth.UserFromGinContext(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  http.StatusInternalServerError,
					"message": err,
				})
				return
			}

			if len(route.RequireRole) > 0 {
				if !auth.HasRole(u, route.GetRequireRole()) {
					c.JSON(http.StatusForbidden, gin.H{
						"status":  http.StatusForbidden,
						"message": "You don't have the privileges to access this resource",
					})
					return
				}
			}

			if len(route.IntersectsRoles) > 0 {
				if !auth.IntersectsRoles(u, route.GetIntersectsRoles()...) {
					c.JSON(http.StatusForbidden, gin.H{
						"status":  http.StatusForbidden,
						"message": "You don't have the privileges to access this resource",
					})
					return
				}
			}
		}

		// Map query/path params
		params := make(map[string]string)
		for _, p := range route.Params {
			if len(c.Query(p)) > 0 {
				params[p] = c.Query(p)
			}
		}
		for _, p := range route.Params {
			if len(c.Param(p)) > 0 {
				params[p] = c.Param(p)
			}
		}

		// Bind the request if POST/PATCH/PUT
		request := gin.H{}
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodPut {
			mf, err := c.MultipartForm()
			if err == nil {
				for k, files := range mf.File {
					for _, file := range files {
						fp, err := file.Open()
						if err != nil {
							continue
						}
						data, err := ioutil.ReadAll(fp)
						if err != nil {
							continue
						}

						if len(files) > 1 {
							if _, ok := request[k]; !ok {
								request[k] = []string{base64.StdEncoding.EncodeToString(data)}
							} else {
								request[k] = append(request[k].([]string), base64.StdEncoding.EncodeToString(data))
							}
						} else {
							request[k] = base64.StdEncoding.EncodeToString(data)
						}
					}
				}

				for k, v := range mf.Value {
					if len(v) > 1 {
						request[k] = v
					} else {
						request[k] = v[0]
					}

				}
			} else {
				c.ShouldBind(&request)
			}
		}

		// Set query/route params to the request
		for pn, p := range params {
			request[pn] = p
		}

		req := (*cmd.DefaultOptions().Client).NewRequest(serviceName, route.GetEndpoint(), request, client.WithContentType("application/json"))

		ctx := utils.RequestToContext(c, c.Request)

		// remote call
		var response json.RawMessage
		err := (*cmd.DefaultOptions().Client).Call(ctx, req, &response)
		if err != nil {
			pErr := errors.FromError(err)
			code := int32(http.StatusInternalServerError)
			if pErr.Code != 0 {
				code = pErr.Code
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  code,
				"message": pErr.Detail,
			})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func (h *Handler) getHealth(c *gin.Context) {
	allFine := true

	services, err := serviceregistry.ServicesFindByEndpoint("InfoService.Health", *cmd.DefaultOptions().Registry, context.TODO())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "We see errors",
		})
		return
	}

	servicesStatus := gin.H{}
	for _, s := range services {
		client := infoservicepb.NewInfoService(s.Name, *cmd.DefaultOptions().Client)
		resp, err := client.Health(c, &empty.Empty{})

		if err != nil {
			allFine = false

			servicesStatus[s.Name] = err.Error()
		} else {
			if resp.GetHasError() {
				allFine = false
			}

			servicesStatus[s.Name] = resp.Infos
		}
	}

	if !allFine {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":   500,
			"message":  "We see errors",
			"services": servicesStatus,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   200,
		"message":  "Everything is fine",
		"services": servicesStatus,
	})
}
