package cmd

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"miaoverse/api"
	"miaoverse/model/server/conf"
	"miaoverse/service/ConfigService"
)

func ServerStart(app *fiber.App, config *conf.AppConfig) {
	servants, err := ConfigService.ConfToServants(config)
	if err != nil {
		panic(err)
	}
	api.Initial(app, servants)
	port := config.Server.Port
	if port == 0 {
		port = 3000
	}
	err = app.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}
}
