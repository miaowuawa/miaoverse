package cmd

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"miaoverse/api"
	"miaoverse/model/server/conf"
	"miaoverse/service/ConfigService"
	"miaoverse/service/i18n"
)

func ServerStart(app *fiber.App, config *conf.AppConfig) {
	if config.I18n.DefaultLanguage != "" {
		i18n.SetDefaultLanguage(config.I18n.DefaultLanguage)
	}
	langDir := config.I18n.Dir
	if langDir == "" {
		langDir = "./locales"
	}
	if err := i18n.LoadDir(langDir); err != nil {
		panic(err)
	}

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
