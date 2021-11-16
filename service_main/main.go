package main

import (
	"log"

	microWeb "go-micro.dev/v4/web"

	"github.com/gin-gonic/gin"
	ginlogrus "github.com/toorop/gin-logrus"
	"github.com/urfave/cli/v2"

	"wz2100.net/microlobby/service_main/version"
	"wz2100.net/microlobby/shared/logger"
)

func main() {
	logger.Setup()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(ginlogrus.Logger(logger.Logger), gin.Recovery())

	service := microWeb.NewService(
		microWeb.Name("microlobby.api.main"),
		microWeb.Version(version.Version),
		microWeb.Handler(router),
	)

	// initialise flags
	if err := service.Init(microWeb.Action(func(c *cli.Context) {})); err != nil {
		log.Fatal(err)
	}

	// start the service
	service.Run()
}
