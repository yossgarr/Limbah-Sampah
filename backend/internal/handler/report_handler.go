package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type reportRequest struct {
	CategoryCode string  `json:"category_code"`
	ActivityDate string  `json:"activity_date"`
	InputAmount  float64 `json:"input_amount"`
}

type factor struct {
	Value   float64 `db:"value"`
	Unit    string  `db:"unit"`
	Source  string  `db:"source_ref"`
	Version string  `db:"version"`
}

func errJSON(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{"error": msg})
}

// ownerFilter: nil (tanpa filter) untuk superadmin, selain itu id user pemilik.
func ownerFilter(c *fiber.Ctx) any {
	if c.Locals("role") == RoleSuperadmin {
		return nil
	}
	return c.Locals("user_id")
}

const (
	maxInputAmountKg = 1000.0 // batas jumlah per laporan
	maxBackdateDays  = 365    // batas tanggal ke belakang
)

func validateReport(req reportRequest) string {
	if req.CategoryCode == "" {
		return "Kategori limbah wajib diisi"
	}
	if math.IsNaN(req.InputAmount) || req.InputAmount <= 0 {
		return "Jumlah sampah harus lebih dari 0"
	}
	if req.InputAmount > maxInputAmountKg {
		return fmt.Sprintf("Jumlah sampah maksimal %.0f kg per laporan", maxInputAmountKg)
	}
	d, err := time.Parse("2006-01-02", req.ActivityDate)
	if err != nil {
		return "Format tanggal harus YYYY-MM-DD"
	}
	now := time.Now()
	if d.After(now) {
		return "Tanggal laporan tidak boleh melebihi hari ini"
	}
	if d.Before(now.AddDate(0, 0, -maxBackdateDays)) {
		return fmt.Sprintf("Tanggal laporan tidak boleh lebih dari %d hari ke belakang", maxBackdateDays)
	}
	return ""
}

func (h *Handler) activeFactor(code string) (factor, error) {
	var f factor
	err := h.DB.Get(&f, `
		SELECT f.value, f.unit, f.source_ref, fs.version_name AS version
		FROM factors f JOIN factor_sets fs ON fs.id = f.set_id
		WHERE fs.is_active AND f.category_code = $1`, code)
	return f, err
}

// ListFactors: daftar kategori + faktor aktif (sumber dropdown di frontend).
func (h *Handler) ListFactors(c *fiber.Ctx) error {
	rows := []struct {
		CategoryCode string  `db:"category_code" json:"category_code"`
		Name         string  `db:"name" json:"name"`
		Value        float64 `db:"value" json:"value"`
		Unit         string  `db:"unit" json:"unit"`
	}{}
	err := h.DB.Select(&rows, `
		SELECT f.category_code, f.name, f.value, f.unit
		FROM factors f JOIN factor_sets fs ON fs.id = f.set_id
		WHERE fs.is_active ORDER BY f.category_code`)
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Gagal mengambil faktor")
	}
	return c.JSON(rows)
}

// ListReports: user hanya melihat laporannya sendiri; observer & superadmin melihat semua.
func (h *Handler) ListReports(c *fiber.Ctx) error {
	var owner any
	if c.Locals("role") == RoleUser {
		owner = c.Locals("user_id")
	}

	reports := []struct {
		ID           string  `db:"id" json:"id"`
		CategoryCode string  `db:"category_code" json:"category_code"`
		InputAmount  float64 `db:"input_amount" json:"input_amount"`
		ResultTotal  float64 `db:"result_total" json:"result_total"`
		ActivityDate string  `db:"activity_date" json:"activity_date"`
	}{}
	err := h.DB.Select(&reports, `
		SELECT id, category_code, input_amount, result_total,
		       to_char(activity_date, 'YYYY-MM-DD') AS activity_date
		FROM waste_reports
		WHERE is_deleted = false AND ($1::uuid IS NULL OR user_id = $1::uuid)
		ORDER BY waste_reports.activity_date DESC, created_at DESC`, owner)
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Gagal mengambil laporan")
	}
	return c.JSON(reports)
}

func (h *Handler) CreateReport(c *fiber.Ctx) error {
	var req reportRequest
	if err := c.BodyParser(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if msg := validateReport(req); msg != "" {
		return errJSON(c, fiber.StatusBadRequest, msg)
	}

	f, err := h.activeFactor(req.CategoryCode)
	if errors.Is(err, sql.ErrNoRows) {
		return errJSON(c, fiber.StatusBadRequest, "Kategori limbah tidak dikenal")
	}
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Gagal mengambil faktor")
	}
	result := math.Round(req.InputAmount*f.Value*1e6) / 1e6

	// Atribut analitik user disalin ke laporan (snapshot, PRD 2.4)
	var id string
	err = h.DB.Get(&id, `
		INSERT INTO waste_reports (
			user_id, category_code, activity_date, input_amount,
			factor_value, factor_unit, factor_source, factor_version, result_total,
			user_satuan_kerja, user_unit_kerja, user_status_pegawai, user_golongan_group, user_age_band
		)
		SELECT id, $2, $3::date, $4::numeric, $5::numeric, $6, $7, $8, $9::numeric,
		       satuan_kerja, unit_kerja, status_pegawai, golongan_group, age_band
		FROM users WHERE id = $1::uuid
		RETURNING id`,
		c.Locals("user_id"), req.CategoryCode, req.ActivityDate, req.InputAmount,
		f.Value, f.Unit, f.Source, f.Version, result)
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Gagal menyimpan laporan")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Laporan berhasil ditambahkan", "id": id})
}

func (h *Handler) UpdateReport(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return errJSON(c, fiber.StatusNotFound, "Laporan tidak ditemukan")
	}

	var req reportRequest
	if err := c.BodyParser(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if msg := validateReport(req); msg != "" {
		return errJSON(c, fiber.StatusBadRequest, msg)
	}

	f, err := h.activeFactor(req.CategoryCode)
	if errors.Is(err, sql.ErrNoRows) {
		return errJSON(c, fiber.StatusBadRequest, "Kategori limbah tidak dikenal")
	}
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Gagal mengambil faktor")
	}
	result := math.Round(req.InputAmount*f.Value*1e6) / 1e6

	res, err := h.DB.Exec(`
		UPDATE waste_reports SET
			category_code = $1, activity_date = $2::date, input_amount = $3::numeric,
			factor_value = $4::numeric, factor_unit = $5, factor_source = $6, factor_version = $7,
			result_total = $8::numeric, updated_at = NOW()
		WHERE id = $9::uuid AND is_deleted = false AND ($10::uuid IS NULL OR user_id = $10::uuid)`,
		req.CategoryCode, req.ActivityDate, req.InputAmount,
		f.Value, f.Unit, f.Source, f.Version, result, id, ownerFilter(c))
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Gagal memperbarui laporan")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errJSON(c, fiber.StatusNotFound, "Laporan tidak ditemukan")
	}

	return c.JSON(fiber.Map{"message": "Laporan berhasil diperbarui"})
}

// DeleteReport: soft delete (kolom is_deleted).
func (h *Handler) DeleteReport(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return errJSON(c, fiber.StatusNotFound, "Laporan tidak ditemukan")
	}

	res, err := h.DB.Exec(`
		UPDATE waste_reports SET is_deleted = true, updated_at = NOW()
		WHERE id = $1::uuid AND is_deleted = false AND ($2::uuid IS NULL OR user_id = $2::uuid)`,
		id, ownerFilter(c))
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Gagal menghapus laporan")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errJSON(c, fiber.StatusNotFound, "Laporan tidak ditemukan")
	}

	return c.JSON(fiber.Map{"message": "Laporan berhasil dihapus"})
}
