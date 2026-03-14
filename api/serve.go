package api

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

func Initial(app *fiber.App) {
	app.Get("/", func(c fiber.Ctx) error {
		time := time.Now().Format("2006-01-02 15:04:05")
		return c.SendString("Miaoverse Content/User API Resp at " + time)
	})

}
