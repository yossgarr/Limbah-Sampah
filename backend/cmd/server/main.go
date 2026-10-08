package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/joho/godotenv"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"daur-ulang/pkg/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, membaca dari OS Environment Variable")
	}

	// 3. AMBIL DATABASE_URL DARI .ENV
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL wajib diisi di file .env atau environment variable")
	}

	// Buat koneksi ke database
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	defer db.Close()

	// Eksekusi auto migration dari file SQL
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Gagal menjalankan migrasi database: %v", err)
	}
	
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Get("/api/v1/waste-reports", func(c *fiber.Ctx) error {
        rows, err := db.Query("SELECT id, category_code, input_amount, result_total, activity_date FROM waste_reports ORDER BY activity_date DESC")
        if err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": err.Error(),
            })
        }
        defer rows.Close()

        reports := []fiber.Map{}
        for rows.Next() {
            var id, categoryCode, activityDate string
            var inputAmount, resultTotal float64

            if err := rows.Scan(&id, &categoryCode, &inputAmount, &resultTotal, &activityDate); err == nil {
                reports = append(reports, fiber.Map{
                    "id":            id,
                    "category_code": categoryCode,
                    "input_amount":  inputAmount,
                    "result_total":  resultTotal,
                    "activity_date": activityDate,
                })
            }
        }

        return c.JSON(reports)
    })
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