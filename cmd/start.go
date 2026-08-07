package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/api"
	"miaoverse/model/server/conf"
	"miaoverse/service/ConfigService"
	"miaoverse/service/MetaSync"
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

	// 启动动态计数定期校准任务（默认每 10 分钟校准一次）
	syncCtx, syncCancel := context.WithCancel(context.Background())
	MetaSync.Start(syncCtx, servants, 10*time.Minute)
	defer syncCancel()

	port := config.Server.Port
	if port == 0 {
		port = 3000
	}
	err = app.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}
}
