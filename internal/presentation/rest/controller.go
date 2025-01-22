package rest

import "github.com/gofiber/fiber/v3"

type Controller interface {
	Init(root fiber.Router)
}
