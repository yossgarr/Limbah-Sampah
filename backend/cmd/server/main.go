package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"daur-ulang/internal/handler"
	"daur-ulang/pkg/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, membaca dari OS Environment Variable")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL wajib diisi di file .env atau environment variable")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET wajib diisi (minimal 32 karakter)")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Gagal menjalankan migrasi database: %v", err)
	}

	h := &handler.Handler{DB: db, JWTSecret: []byte(jwtSecret)}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")
	api.Post("/auth/login-external", h.ExternalLogin)

	// Semua role boleh melihat; hanya user & superadmin yang boleh menulis.
	canWrite := handler.RequireRole(handler.RoleUser, handler.RoleSuperadmin)
	api.Get("/factors", h.Authenticate, h.ListFactors)
	api.Get("/waste-reports", h.Authenticate, h.ListReports)
	api.Post("/waste-reports", h.Authenticate, canWrite, h.CreateReport)
	api.Put("/waste-reports/:id", h.Authenticate, canWrite, h.UpdateReport)
	api.Delete("/waste-reports/:id", h.Authenticate, canWrite, h.DeleteReport)

	log.Fatal(app.Listen(":" + port))
}
