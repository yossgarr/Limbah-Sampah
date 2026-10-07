package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Group Route API v1
	api := app.Group("/api/v1")
	auth := api.Group("/auth")

	// Endpoint Auth
	auth.Get("/sso/callback", handleSSOCallback)
	auth.Post("/login-external", handleExternalLogin)

	log.Fatal(app.Listen(":8080"))
}

func handleSSOCallback(c *fiber.Ctx) error {
	// Logic callback OAuth2/PKCE Keycloak
	return c.JSON(fiber.Map{
		"access_token":  "dummy-sso-access-token",
		"refresh_token": "dummy-sso-refresh-token",
	})
}

func handleExternalLogin(c *fiber.Ctx) error {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	// Logic verifikasi akun eksternal & hash argon2id
	return c.JSON(fiber.Map{
		"access_token":  "dummy-ext-access-token",
		"refresh_token": "dummy-ext-refresh-token",
		"role":          "observer",
	})
}