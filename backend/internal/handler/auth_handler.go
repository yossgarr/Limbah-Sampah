package handler

import (
	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(router fiber.Router) {
	auth := router.Group("/auth")
	
	// FR-AUTH-2: Endpoint callback SSO[cite: 1]
	auth.Get("/sso/callback", handleSSOCallback)
	
	// FR-AUTH-9: Login khusus akun eksternal[cite: 1]
	auth.Post("/external/login", handleExternalLogin)
}

func handleSSOCallback(c *fiber.Ctx) error {
	// 1. Terima code & state dari Keycloak
	// 2. Tukar code dengan token Keycloak
	// 3. Baca 'sub' dan atribut analitik (Satuan Kerja, dll)[cite: 1]
	// 4. Buat user baru jika 'sub' belum ada, default role: 'user' (R-2)[cite: 1]
	// 5. Kembalikan token akses & refresh internal
	
	return c.JSON(fiber.Map{
		"access_token":  "jwt-token-sso",
		"refresh_token": "refresh-token-sso",
	})
}

func handleExternalLogin(c *fiber.Ctx) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format tidak valid"})
	}

	// Sesuai FR-AUTH-11: Pesan error harus generik agar email tidak bocor[cite: 1]
	genericError := fiber.Map{"error": "Email atau password salah"}

	// 1. Cari user dengan auth_source = 'external' dan email = req.Email
	// 2. Jika tidak ada / nonaktif / kedaluwarsa -> kembalikan genericError
	// 3. Cek locked_until. Jika terkunci -> kembalikan genericError
	// 4. Verifikasi PasswordHash dengan Argon2id. Jika gagal -> tambah failed_login_count.
	// 5. Jika gagal 5x -> set locked_until = now + 15 menit[cite: 1]
	
	// Jika sukses:
	return c.JSON(fiber.Map{
		"access_token":  "jwt-token-external",
		"refresh_token": "refresh-token-external",
		"role":          "observer", // Akun eksternal selalu observer (R-8)[cite: 1]
	})
}