package cmd

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/api"
	"miaoverse/model/server/conf"
	"miaoverse/service/ConfigService"
)

func ServerStart(app *fiber.App, config *conf.AppConfig) {
	err, servants := ConfigService.ConfToServants(config)
	if err != nil {
		panic(err)
	}
	api.Initial(app, servants)
	err = app.Listen(":3000")
	if err != nil {
		return
	}
}
