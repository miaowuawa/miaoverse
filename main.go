package main

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/cmd"
	"miaoverse/service/ConfigService"
)

const multipartBodyOverhead = 1024 * 1024

func main() {
	err, conf := ConfigService.InitConfig("./")
	if err != nil {
		panic(err)
	}
	app := fiber.New(fiber.Config{
		BodyLimit: int(conf.UploadMaxFileSizeBytes()) + multipartBodyOverhead,
	})
	cmd.ServerStart(app, conf)
}
