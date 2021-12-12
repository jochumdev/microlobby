package proxy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/cmd"
	"go-micro.dev/v4/errors"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/serviceregistry"
	"wz2100.net/microlobby/shared/utils"

	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
)

const pkgPath = "wz2100.net/microlobby/http_proxy/web/proxy"

type JSONRoute struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// Handler is the handler for the proxy
type Handler struct {
	cRegistry        *component.Registry
	registeredRoutes map[string]bool
	routingengine    *gin.Engine
}

// ConfigureRouter creates a new Handler and configures it on the router
func ConfigureRouter(cregistry *component.Registry, r *gin.Engine) *Handler {
	h := &Handler{
		cRegistry:        cregistry,
		registeredRoutes: make(map[string]bool),
		routingengine:    r,
	}

	proxyAuthGroup := r.Group(fmt.Sprintf("/%s", defs.ProxyURIHttpProxy))
	// proxyAuthGroup.Use(middlewaregin.UserSrvMiddleware(cregistry))
	// proxyAuthGroup.Use(middlewaregin.RequireUserMiddleware(cregistry))

	proxyAuthGroup.GET("/v1/health", h.getHealth)
	proxyAuthGroup.GET("/v1/routes", h.getRoutes)

	globalGroup := r.Group("")
	globalAuthGroup := r.Group("")
	// globalAuthGroup.Use(middlewaregin.UserSrvMiddleware(cregistry))
	// globalAuthGroup.Use(middlewaregin.RequireUserMiddleware(cregistry))

	// Refresh routes for the proxy every 10 seconds
	go func() {
		ctx := context.Background()

		for {
			services, err := serviceregistry.ServicesFindByEndpoint("InfoService.Routes", *cmd.DefaultOptions().Registry, ctx)
			if err != nil {
				return
			}

			for _, s := range services {
				client := infoservicepb.NewInfoService(s.Name, *cmd.DefaultOptions().Client)
				resp, err := client.Routes(ctx, &empty.Empty{})
				if err != nil {
					// failure in getting routes, silently ignore
					continue
				}

				serviceGroup := r.Group(fmt.Sprintf("/%s/%s", resp.GetProxyURI(), resp.GetApiVersion()))
				serviceAuthGroup := r.Group(fmt.Sprintf("/%s/%s", resp.GetProxyURI(), resp.GetApiVersion()))
				// serviceAuthGroup.Use(middlewaregin.UserSrvMiddleware(cregistry))
				// serviceAuthGroup.Use(middlewaregin.RequireUserMiddleware(cregistry))

				for _, route := range resp.Routes {
					var g *gin.RouterGroup = nil

					if route.GetGlobalRoute() {
						if len(route.RequireRole) > 0 || len(route.IntersectsRoles) > 0 {
							g = globalAuthGroup
						} else {
							g = globalGroup
						}
					} else {
						if len(route.RequireRole) > 0 || len(route.IntersectsRoles) > 0 {
							g = serviceAuthGroup
						} else {
							g = serviceGroup
						}
					}

					// Calculate the path of the route and register it if it's not registered yet
					path := fmt.Sprintf("%s: %s/%s", route.Method, g.BasePath(), route.Path)
					if _, ok := h.registeredRoutes[path]; !ok {
						g.Handle(route.GetMethod(), route.GetPath(), h.proxy(s.Name, route))
						h.registeredRoutes[path] = true
					}
				}
			}

			time.Sleep(10 * time.Second)
		}
	}()

	return h
}

func (h *Handler) proxy(serviceName string, route *infoservicepb.RoutesReply_Route) func(*gin.Context) {
	return func(c *gin.Context) {
		// Check if the user has the required role
		// if len(route.RequireRole) > 0 || len(route.IntersectsRoles) > 0 {
		// 	u, err := auth.UserFromGinContext(c)
		// 	if err != nil {
		// 		c.JSON(http.StatusInternalServerError, gin.H{
		// 			"status":  http.StatusInternalServerError,
		// 			"message": err,
		// 		})
		// 		return
		// 	}

		// 	if len(route.RequireRole) > 0 {
		// 		if !auth.HasRole(u, route.GetRequireRole()) {
		// 			c.JSON(http.StatusForbidden, gin.H{
		// 				"status":  http.StatusForbidden,
		// 				"message": "You don't have the privileges to access this resource",
		// 			})
		// 			return
		// 		}
		// 	}

		// 	if len(route.IntersectsRoles) > 0 {
		// 		if !auth.IntersectsRoles(u, route.GetIntersectsRoles()...) {
		// 			c.JSON(http.StatusForbidden, gin.H{
		// 				"status":  http.StatusForbidden,
		// 				"message": "You don't have the privileges to access this resource",
		// 			})
		// 			return
		// 		}
		// 	}
		// }

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
				if c.ContentType() == "" {
					c.JSON(http.StatusUnsupportedMediaType, gin.H{
						"status":  http.StatusUnsupportedMediaType,
						"message": "provide a content-type header",
					})
					return
				}
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
			if cLogrus, lErr := component.Logrus(h.cRegistry); lErr != nil {
				cLogrus.WithClassFunc(pkgPath, "Handler", "proxy").Error(err)
			}

			pErr := errors.FromError(err)
			code := int(http.StatusInternalServerError)
			if pErr.Code != 0 {
				code = int(pErr.Code)
			}
			c.JSON(code, gin.H{
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

func (h *Handler) getRoutes(c *gin.Context) {
	ginRoutes := h.routingengine.Routes()
	rRoutes := []JSONRoute{}
	for _, route := range ginRoutes {
		rRoutes = append(rRoutes, JSONRoute{Method: route.Method, Path: route.Path})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "Dumping the routes *lalalala*",
		"data":    rRoutes,
	})
}
