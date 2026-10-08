package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"daur-ulang/pkg/database"
)

type WasteReportReq struct {
	CategoryCode string  `json:"category_code"`
	ActivityDate string  `json:"activity_date"`
	InputAmount  float64 `json:"input_amount"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, membaca dari OS Environment Variable")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL wajib diisi di file .env atau environment variable")
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Gagal menjalankan migrasi database: %v", err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(logger.New())

	api := app.Group("/api/v1")

	// GET: Fetch Semua Laporan
	api.Get("/waste-reports", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, category_code, input_amount, result_total, activity_date FROM waste_reports ORDER BY activity_date DESC")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
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

	// POST: Tambah Laporan Baru
	api.Post("/waste-reports", func(c *fiber.Ctx) error {
		var req WasteReportReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Format request tidak valid"})
		}

		// Validasi Server-Side
		if req.CategoryCode == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Kategori limbah wajib diisi"})
		}
		if req.InputAmount <= 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Jumlah sampah harus lebih dari 0"})
		}
		parsedDate, err := time.Parse("2006-01-02", req.ActivityDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Format tanggal harus YYYY-MM-DD"})
		}
		if parsedDate.After(time.Now()) {
			return c.Status(400).JSON(fiber.Map{"error": "Tanggal laporan tidak boleh melebihi hari ini"})
		}

		factorValue := 1.0
		resultTotal := req.InputAmount * factorValue
		dummyUserID := "00000000-0000-0000-0000-000000000001"

		var createdID string
		query := `
			INSERT INTO waste_reports (user_id, category_code, activity_date, input_amount, factor_value, factor_unit, result_total)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id`

		err = db.QueryRow(query, dummyUserID, req.CategoryCode, req.ActivityDate, req.InputAmount, factorValue, "kg", resultTotal).Scan(&createdID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan laporan: " + err.Error()})
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "Laporan berhasil ditambahkan",
			"id":      createdID,
		})
	})

	// PUT: Update Laporan
	api.Put("/waste-reports/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var req WasteReportReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Format request tidak valid"})
		}

		if req.InputAmount <= 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Jumlah sampah harus lebih dari 0"})
		}

		factorValue := 1.0
		resultTotal := req.InputAmount * factorValue

		query := `
			UPDATE waste_reports 
			SET category_code = $1, activity_date = $2, input_amount = $3, result_total = $4
			WHERE id = $5`

		res, err := db.Exec(query, req.CategoryCode, req.ActivityDate, req.InputAmount, resultTotal, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui laporan"})
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Laporan tidak ditemukan"})
		}

		return c.JSON(fiber.Map{"message": "Laporan berhasil diperbarui"})
	})

	// DELETE: Hapus Laporan
	api.Delete("/waste-reports/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		res, err := db.Exec("DELETE FROM waste_reports WHERE id = $1", id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus laporan"})
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Laporan tidak ditemukan"})
		}

		return c.JSON(fiber.Map{"message": "Laporan berhasil dihapus"})
	})

	// Health Check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Auth Group
	auth := api.Group("/auth")
	auth.Get("/sso/callback", handleSSOCallback)
	auth.Post("/login-external", handleExternalLogin)

	log.Fatal(app.Listen(":8080"))
}

func handleSSOCallback(c *fiber.Ctx) error {
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

	return c.JSON(fiber.Map{
		"access_token":  "dummy-ext-access-token",
		"refresh_token": "dummy-ext-refresh-token",
		"role":          "observer",
	})
}