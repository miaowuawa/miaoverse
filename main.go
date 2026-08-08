package main

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/cmd"
	"miaoverse/consts"
	"miaoverse/service/ConfigService"
)

func main() {
	err, conf := ConfigService.InitConfig("./")
	if err != nil {
		panic(err)
	}
	app := fiber.New(fiber.Config{
		BodyLimit: int(conf.UploadMaxFileSizeBytes()) + consts.MultipartBodyOverhead,
	})
	cmd.ServerStart(app, conf)
}
