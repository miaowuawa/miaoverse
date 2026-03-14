package miaoverse

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/cmd"
	"miaoverse/service/ConfigService"
)

func main() {
	err, conf := ConfigService.InitConfig("./")
	if err != nil {
		panic(err)
	}
	app := fiber.New()
	cmd.ServerStart(app, conf)
}
