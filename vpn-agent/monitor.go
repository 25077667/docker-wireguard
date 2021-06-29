package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func get_sys_info() string {
	// TODO: Get the runtime metrics
	return ""
}

func main() {
	app := fiber.New()
	app.Get("/stat", func(c *fiber.Ctx) error {
		return c.SendString(get_sys_info())
	})
	app.Get("/key", func(c *fiber.Ctx) error {
		id := c.Get("id", "") // Suppose it is a string of int
		path := fmt.Sprintf("/config/peer%s/peer%s.conf", id, id)
		return c.SendFile(path) // give Sendfile to check if the file is exist
	})

	app.Listen(":8080")
}
