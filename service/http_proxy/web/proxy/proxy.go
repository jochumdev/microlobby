package proxy

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/util/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"wz2100.net/microlobby/service/http_proxy/config"
	scomponent "wz2100.net/microlobby/service/settings/v1/component"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	middlewareGin "wz2100.net/microlobby/shared/middleware/gin"
	"wz2100.net/microlobby/shared/serviceregistry"
	"wz2100.net/microlobby/shared/utils"

	"wz2100.net/microlobby/shared/proto/infoservicepb/v1"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
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
	settings         map[string][]byte
}

// ConfigureRouter creates a new Handler and configures it on the router
func ConfigureRouter(cregistry *component.Registry, r *gin.Engine) *Handler {
	h := &Handler{
		cRegistry:        cregistry,
		registeredRoutes: make(map[string]bool),
		routingengine:    r,
		settings:         make(map[string][]byte),
	}

	globalGroup := r.Group("")
	globalGroup.Use(middlewareGin.UserSrvMiddleware(cregistry))

	proxyAuthGroup := globalGroup.Group(fmt.Sprintf("/%s", config.ProxyURI))
	proxyAuthGroup.Use(middlewareGin.RequireUserMiddleware(cregistry))

	proxyAuthGroup.GET("/v1/health", h.getHealth)
	proxyAuthGroup.GET("/v1/routes", h.getRoutes)

	globalAuthGroup := globalGroup.Group("")
	globalAuthGroup.Use(middlewareGin.RequireUserMiddleware(cregistry))

	// Refresh routes for the proxy every 10 seconds
	go func() {
		ctx := context.Background()

		for {
			services, err := serviceregistry.FindByEndpoint(ctx, h.cRegistry, "InfoService.Routes")
			if err != nil {
				return
			}

			for _, s := range services {
				client := infoservicepb.NewInfoService(s.Name, cregistry.Client)
				resp, err := client.Routes(ctx, &emptypb.Empty{})
				if err != nil {
					// failure in getting routes, silently ignore
					continue
				}

				serviceGroup := globalGroup.Group(fmt.Sprintf("/%s/%s", resp.GetProxyURI(), resp.GetApiVersion()))
				serviceAuthGroup := globalAuthGroup.Group(fmt.Sprintf("/%s/%s", resp.GetProxyURI(), resp.GetApiVersion()))

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

	go func() {
		ctx := component.RegistryToContext(utils.CtxForService(context.Background()), cregistry)
		s, err := scomponent.SettingsV1(cregistry)
		if err != nil {
			panic(err)
		}

		_, ok := h.settings[config.SettingNameJWTRefreshTokenPub]
		_, ok2 := h.settings[config.SettingNameJWTRefreshTokenPriv]
		if !ok || !ok2 {
			spub, epub := s.Get(ctx, "", "", config.Name, config.SettingNameJWTRefreshTokenPub)
			spri, epri := s.Get(ctx, "", "", config.Name, config.SettingNameJWTRefreshTokenPriv)
			if epub != nil || epri != nil {
				pubKey, privKey, err := ed25519.GenerateKey(nil)
				if err != nil {
					panic(err)
				}

				npri := privKey
				npub := pubKey

				spri, epri = s.Upsert(ctx, &settingsservicepb.UpsertRequest{
					Service:     config.Name,
					Name:        config.SettingNameJWTRefreshTokenPriv,
					Content:     npri,
					RolesRead:   []string{auth.ROLE_SUPERADMIN, auth.ROLE_SERVICE},
					RolesUpdate: []string{auth.ROLE_SUPERADMIN, auth.ROLE_SERVICE},
				})
				if epri != nil {
					if cLogrus, lErr := component.Logrus(cregistry); lErr == nil {
						cLogrus.WithFunc(pkgPath, "ConfigureRouter").
							WithField("error", epri).
							WithField("setting", config.SettingNameJWTRefreshTokenPriv).
							Error(epri)
					} else {
						log.Error(epri)
					}
					return
				}

				spub, epub = s.Upsert(ctx, &settingsservicepb.UpsertRequest{
					Service:     config.Name,
					Name:        config.SettingNameJWTRefreshTokenPub,
					Content:     npub,
					RolesRead:   []string{auth.ROLE_USER, auth.ROLE_SERVICE},
					RolesUpdate: []string{auth.ROLE_SUPERADMIN, auth.ROLE_SERVICE},
				})
				if epub != nil {
					if cLogrus, lErr := component.Logrus(cregistry); lErr == nil {
						cLogrus.WithFunc(pkgPath, "ConfigureRouter").
							WithField("error", epub).
							WithField("setting", config.SettingNameJWTRefreshTokenPub).
							Error(epub)
					} else {
						log.Error(epub)
					}
					return
				}
			}

			h.settings[config.SettingNameJWTRefreshTokenPub] = spub.Content
			h.settings[config.SettingNameJWTRefreshTokenPriv] = spri.Content
		}

		_, ok = h.settings[config.SettingNameJWTAccessTokenPub]
		_, ok2 = h.settings[config.SettingNameJWTAccessTokenPriv]
		if !ok || !ok2 {
			spub, epub := s.Get(ctx, "", "", config.Name, config.SettingNameJWTAccessTokenPub)
			spri, epri := s.Get(ctx, "", "", config.Name, config.SettingNameJWTAccessTokenPriv)
			if epub != nil || epri != nil {
				pubKey, privKey, err := ed25519.GenerateKey(nil)
				if err != nil {
					panic(err)
				}

				npri := privKey
				npub := pubKey

				spri, epri = s.Upsert(ctx, &settingsservicepb.UpsertRequest{
					Service:     config.Name,
					Name:        config.SettingNameJWTAccessTokenPriv,
					Content:     npri,
					RolesRead:   []string{auth.ROLE_SUPERADMIN, auth.ROLE_SERVICE},
					RolesUpdate: []string{auth.ROLE_SUPERADMIN, auth.ROLE_SERVICE},
				})
				if epri != nil {
					if cLogrus, lErr := component.Logrus(cregistry); lErr == nil {
						cLogrus.WithFunc(pkgPath, "ConfigureRouter").
							WithField("error", epri).
							WithField("setting", config.SettingNameJWTAccessTokenPriv).
							Error(epri)
					} else {
						log.Error(epri)
					}
					return
				}

				spub, epub = s.Upsert(ctx, &settingsservicepb.UpsertRequest{
					Service:     config.Name,
					Name:        config.SettingNameJWTAccessTokenPub,
					Content:     npub,
					RolesRead:   []string{auth.ROLE_USER, auth.ROLE_SERVICE},
					RolesUpdate: []string{auth.ROLE_SUPERADMIN, auth.ROLE_SERVICE},
				})
				if epub != nil {
					if cLogrus, lErr := component.Logrus(cregistry); lErr == nil {
						cLogrus.WithFunc(pkgPath, "ConfigureRouter").
							WithField("error", epub).
							WithField("setting", config.SettingNameJWTAccessTokenPub).
							Error(epub)
					} else {
						log.Error(epub)
					}
					return
				}
			}

			h.settings[config.SettingNameJWTAccessTokenPub] = spub.Content
			h.settings[config.SettingNameJWTAccessTokenPriv] = spri.Content
		}
	}()

	return h
}

func (h *Handler) proxy(serviceName string, route *infoservicepb.RoutesReply_Route) func(*gin.Context) {
	return func(c *gin.Context) {
		// Check if the user has the required role
		if len(route.RequireRole) > 0 || len(route.IntersectsRoles) > 0 {
			u, err := middlewareGin.UserFromContext(c)
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

		req := h.cRegistry.Client.NewRequest(serviceName, route.GetEndpoint(), request, client.WithContentType("application/json"))

		ctx := utils.CtxFromRequest(c, c.Request)

		// remote call
		var response json.RawMessage
		err := h.cRegistry.Client.Call(ctx, req, &response)
		if err != nil {
			if cLogrus, lErr := component.Logrus(h.cRegistry); lErr == nil {
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
	// Check permissions ( ROLE_ADMIN )
	if ok := middlewareGin.ForceRole(c, auth.ROLE_ADMIN); !ok {
		return
	}

	allFine := true

	services, err := serviceregistry.FindByEndpoint(context.TODO(), h.cRegistry, "InfoService.Health")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "We see errors",
		})
		return
	}

	servicesStatus := gin.H{}
	for _, s := range services {
		client := infoservicepb.NewInfoService(s.Name, h.cRegistry.Client)
		resp, err := client.Health(c, &emptypb.Empty{})

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
	// Check permissions ( ROLE_ADMIN )
	if ok := middlewareGin.ForceRole(c, auth.ROLE_ADMIN); !ok {
		return
	}

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
